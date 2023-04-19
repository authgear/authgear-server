package otp

import (
	"errors"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type GenerateOptions struct {
	UserID         string
	WorkflowID     string
	WebSessionID   string
	SkipRateLimits bool
}

type VerifyOptions struct {
	UserID           string
	UseSubmittedCode bool
	SkipConsume      bool
}

type CodeStore interface {
	Create(purpose Purpose, code *Code) error
	Get(purpose Purpose, target string) (*Code, error)
	Update(purpose Purpose, code *Code) error
	Delete(purpose Purpose, target string) error
}

type LookupStore interface {
	Create(purpose Purpose, code string, target string, expireAt time.Time) error
	Get(purpose Purpose, code string) (string, error)
	Delete(purpose Purpose, code string) error
}

type AttemptTracker interface {
	ResetFailedAttempts(purpose Purpose, target string) error
	GetFailedAttempts(purpose Purpose, target string) (int, error)
	IncrementFailedAttempts(purpose Purpose, target string) (int, error)
}

type RateLimiter interface {
	Allow(spec ratelimit.BucketSpec) error
	Reserve(spec ratelimit.BucketSpec) *ratelimit.Reservation
	ReserveN(spec ratelimit.BucketSpec, n int) *ratelimit.Reservation
	Cancel(r *ratelimit.Reservation) error
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("otp")} }

type Service struct {
	Clock clock.Clock

	AppID          config.AppID
	RemoteIP       httputil.RemoteIP
	CodeStore      CodeStore
	LookupStore    LookupStore
	AttemptTracker AttemptTracker
	Logger         Logger
	RateLimiter    RateLimiter
}

func (s *Service) getCode(purpose Purpose, target string) (*Code, error) {
	return s.CodeStore.Get(purpose, target)
}

func (s *Service) deleteCode(kind Kind, target string) error {
	if err := s.CodeStore.Delete(kind.Purpose(), target); err != nil {
		return err
	}
	// No need delete from lookup store;
	// lookup entry is invalidated since target is no longer exist.
	return nil
}

func (s *Service) consumeCode(purpose Purpose, code *Code) error {
	code.Consumed = true
	if err := s.CodeStore.Update(purpose, code); err != nil {
		return err
	}
	// No need delete from lookup store;
	// lookup entry is invalidated since target is no longer exist.
	return nil
}

func (s *Service) handleFailedAttemptsRevocation(kind Kind, target string) error {
	failedAttempts, err := s.AttemptTracker.IncrementFailedAttempts(kind.Purpose(), target)
	if err != nil {
		return err
	}

	maxFailedAttempts := kind.RevocationMaxFailedAttempts()
	if maxFailedAttempts != 0 && failedAttempts >= maxFailedAttempts {
		return ErrTooManyAttempts
	}

	return nil
}

func (s *Service) checkFailedAttemptsRevocation(kind Kind, target string) error {
	failedAttempts, err := s.AttemptTracker.GetFailedAttempts(kind.Purpose(), target)
	if err != nil {
		return err
	}

	maxFailedAttempts := kind.RevocationMaxFailedAttempts()
	if maxFailedAttempts != 0 && failedAttempts >= maxFailedAttempts {
		err = s.deleteCode(kind, target)
		if err != nil {
			s.Logger.WithError(err).Warn("failed to revoke OTP")
		}
		return ErrTooManyAttempts
	}

	return nil
}

func (s *Service) GenerateOTP(kind Kind, target string, form Form, opts *GenerateOptions) (string, error) {
	if !opts.SkipRateLimits {
		if err := s.RateLimiter.Allow(kind.RateLimitTriggerCooldown(target)); err != nil {
			return "", err
		}

		if err := s.RateLimiter.Allow(kind.RateLimitTriggerPerIP(string(s.RemoteIP))); err != nil {
			return "", err
		}

		if opts.UserID != "" {
			if err := s.RateLimiter.Allow(kind.RateLimitTriggerPerUser(opts.UserID)); err != nil {
				return "", err
			}
		}
	}

	code := &Code{
		Target:   target,
		Purpose:  kind.Purpose(),
		Form:     form,
		Code:     form.GenerateCode(),
		ExpireAt: s.Clock.NowUTC().Add(kind.ValidPeriod()),

		UserID:       opts.UserID,
		WorkflowID:   opts.WorkflowID,
		WebSessionID: opts.WebSessionID,
	}

	err := s.CodeStore.Create(kind.Purpose(), code)
	if err != nil {
		return "", err
	}

	if form.AllowLookupByCode() {
		err := s.LookupStore.Create(code.Purpose, code.Code, code.Target, code.ExpireAt)
		if err != nil {
			return "", err
		}
	}

	if err := s.AttemptTracker.ResetFailedAttempts(kind.Purpose(), target); err != nil {
		// non-critical error; log and continue
		s.Logger.WithError(err).Warn("failed to reset failed attempts counter")
	}

	return code.Code, nil
}

func (s *Service) VerifyOTP(kind Kind, target string, otp string, opts *VerifyOptions) error {
	if err := s.checkFailedAttemptsRevocation(kind, target); err != nil {
		return err
	}

	isCodeValid := false

	reservation1 := s.RateLimiter.Reserve(kind.RateLimitValidatePerIP(string(s.RemoteIP)))
	if err := reservation1.Error(); err != nil {
		return err
	}
	defer func() {
		if isCodeValid {
			if err := s.RateLimiter.Cancel(reservation1); err != nil {
				// non-critical error; log and continue
				s.Logger.WithError(err).Warn("failed to return rate limit tokens")
			}
		}
	}()

	if opts.UserID != "" {
		reservation2 := s.RateLimiter.Reserve(kind.RateLimitValidatePerUserPerIP(opts.UserID, string(s.RemoteIP)))
		if err := reservation2.Error(); err != nil {
			return err
		}
		defer func() {
			if isCodeValid {
				if err := s.RateLimiter.Cancel(reservation2); err != nil {
					// non-critical error; log and continue
					s.Logger.WithError(err).Warn("failed to return rate limit tokens")
				}
			}
		}()
	}

	code, err := s.getCode(kind.Purpose(), target)
	if errors.Is(err, ErrCodeNotFound) {
		return ErrInvalidCode
	} else if err != nil {
		return err
	}

	if code.Purpose != kind.Purpose() {
		return ErrInvalidCode
	}
	if code.Consumed {
		return ErrConsumedCode
	}

	codeToVerify := otp
	if opts.UseSubmittedCode {
		codeToVerify = code.UserInputtedCode
	}

	if !code.Form.VerifyCode(codeToVerify, code.Code) {
		ferr := s.handleFailedAttemptsRevocation(kind, target)
		if errors.Is(ferr, ErrTooManyAttempts) {
			return ferr
		} else if ferr != nil {
			// log the error, and return original error
			s.Logger.WithError(ferr).Warn("failed to handle failed attempts")
		}
		return ErrInvalidCode
	}

	// Set flag to return reserved rate limit tokens
	isCodeValid = true

	if !opts.SkipConsume {
		if err := s.consumeCode(kind.Purpose(), code); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) ConsumeCode(purpose Purpose, target string) error {
	code, err := s.getCode(purpose, target)
	if errors.Is(err, ErrCodeNotFound) {
		return nil
	} else if err != nil {
		return err
	}

	if code.Purpose != purpose {
		return nil
	}
	if code.Consumed {
		return nil
	}

	return s.consumeCode(purpose, code)
}

func (s *Service) SetSubmittedCode(kind Kind, target string, code string) (*State, error) {
	codeModel, err := s.getCode(kind.Purpose(), target)
	if err != nil {
		return nil, err
	}

	codeModel.UserInputtedCode = code
	if err := s.CodeStore.Update(kind.Purpose(), codeModel); err != nil {
		return nil, err
	}

	return s.InspectState(kind, target)
}

func (s *Service) LookupCode(purpose Purpose, code string) (target string, err error) {
	return s.LookupStore.Get(purpose, code)
}

// InspectWhatsappOTP is deprecated. Whatsapp OTP is to be revamped.
func (s *Service) InspectWhatsappOTP(kind Kind, target string) (string, error) {
	code, err := s.getCode(kind.Purpose(), target)
	if err != nil {
		return "", err
	}

	if code != nil && code.Purpose != kind.Purpose() {
		return "", ErrCodeNotFound
	}

	return code.Code, nil
}

func (s *Service) InspectState(kind Kind, target string) (*State, error) {
	ferr := s.checkFailedAttemptsRevocation(kind, target)
	tooManyAttempts := false
	if errors.Is(ferr, ErrTooManyAttempts) {
		tooManyAttempts = true
	} else if ferr != nil {
		return nil, ferr
	}

	// Inspect rate limit state by reserving no tokens.
	reservation := s.RateLimiter.ReserveN(kind.RateLimitTriggerCooldown(target), 0)
	now := s.Clock.NowUTC()
	canResendAt := now.Add(reservation.DelayFrom(now))
	if err := s.RateLimiter.Cancel(reservation); err != nil {
		return nil, err
	}

	state := &State{
		ExpireAt:        now,
		CanResendAt:     canResendAt,
		SubmittedCode:   "",
		TooManyAttempts: tooManyAttempts,
	}

	code, err := s.getCode(kind.Purpose(), target)
	if errors.Is(err, ErrCodeNotFound) {
		code = nil
	} else if err != nil {
		return nil, err
	}

	if code != nil && code.Purpose != kind.Purpose() {
		return nil, ErrCodeNotFound
	}
	if code != nil && code.Consumed {
		return nil, ErrConsumedCode
	}

	if code != nil {
		state.ExpireAt = code.ExpireAt
		state.SubmittedCode = code.UserInputtedCode
		state.UserID = code.UserID
		state.WorkflowID = code.WorkflowID
		state.WebSessionID = code.WebSessionID
	}

	return state, nil
}
