package otp

import (
	"context"
	"errors"
	"time"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type GenerateOptions struct {
	UserID                                 string
	WebSessionID                           string
	WorkflowID                             string
	AuthenticationFlowWebsocketChannelName string
	AuthenticationFlowType                 string
	AuthenticationFlowName                 string
	AuthenticationFlowJSONPointer          jsonpointer.T
	SkipRateLimits                         bool
}

type VerifyOptions struct {
	UserID           string
	UseSubmittedCode bool
	SkipConsume      bool
}

type CodeStore interface {
	Create(ctx context.Context, purpose Purpose, code *Code) error
	Get(ctx context.Context, purpose Purpose, target string) (*Code, error)
	Update(ctx context.Context, purpose Purpose, code *Code) error
	Delete(ctx context.Context, purpose Purpose, target string) error
}

type LookupStore interface {
	Create(ctx context.Context, purpose Purpose, code string, target string, expireAt time.Time) error
	Get(ctx context.Context, purpose Purpose, code string) (string, error)
	Delete(ctx context.Context, purpose Purpose, code string) error
}

type AttemptTracker interface {
	ResetFailedAttempts(ctx context.Context, kind Kind, target string) error
	GetFailedAttempts(ctx context.Context, kind Kind, target string) (int, error)
	IncrementFailedAttempts(ctx context.Context, kind Kind, target string) (int, error)
}

type RateLimiter interface {
	GetTimeToAct(ctx context.Context, spec ratelimit.BucketSpec) (*time.Time, error)
	Allow(ctx context.Context, spec ratelimit.BucketSpec) (*ratelimit.FailedReservation, error)
	Reserve(ctx context.Context, spec ratelimit.BucketSpec) (*ratelimit.Reservation, *ratelimit.FailedReservation, error)
	Cancel(ctx context.Context, r *ratelimit.Reservation)
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("otp")} }

type Service struct {
	Clock clock.Clock

	AppID                 config.AppID
	TestModeConfig        *config.TestModeConfig
	TestModeFeatureConfig *config.TestModeFeatureConfig
	RemoteIP              httputil.RemoteIP
	CodeStore             CodeStore
	LookupStore           LookupStore
	AttemptTracker        AttemptTracker
	Logger                Logger
	RateLimiter           RateLimiter
}

func (s *Service) getCode(ctx context.Context, purpose Purpose, target string) (*Code, error) {
	return s.CodeStore.Get(ctx, purpose, target)
}

func (s *Service) deleteCode(ctx context.Context, kind Kind, target string) error {
	if err := s.CodeStore.Delete(ctx, kind.Purpose(), target); err != nil {
		return err
	}
	// No need delete from lookup store;
	// lookup entry is invalidated since target is no longer exist.
	return nil
}

func (s *Service) consumeCode(ctx context.Context, purpose Purpose, code *Code) error {
	code.Consumed = true
	if err := s.CodeStore.Update(ctx, purpose, code); err != nil {
		return err
	}
	// No need delete from lookup store;
	// lookup entry is invalidated since target is no longer exist.
	return nil
}

func (s *Service) handleFailedAttemptsRevocation(ctx context.Context, kind Kind, target string) error {
	failedAttempts, err := s.AttemptTracker.IncrementFailedAttempts(ctx, kind, target)
	if err != nil {
		return err
	}

	maxFailedAttempts := kind.RevocationMaxFailedAttempts()
	if maxFailedAttempts != 0 && failedAttempts >= maxFailedAttempts {
		return ErrTooManyAttempts
	}

	return nil
}

func (s *Service) checkFailedAttemptsRevocation(ctx context.Context, kind Kind, target string) error {
	failedAttempts, err := s.AttemptTracker.GetFailedAttempts(ctx, kind, target)
	if err != nil {
		return err
	}

	maxFailedAttempts := kind.RevocationMaxFailedAttempts()
	if maxFailedAttempts != 0 && failedAttempts >= maxFailedAttempts {
		err = s.deleteCode(ctx, kind, target)
		if err != nil {
			s.Logger.WithError(err).Warn("failed to revoke OTP")
		}
		return ErrTooManyAttempts
	}

	return nil
}

func (s *Service) GenerateOTP(ctx context.Context, kind Kind, target string, form Form, opts *GenerateOptions) (string, error) {
	if !opts.SkipRateLimits {
		failed, err := s.RateLimiter.Allow(ctx, kind.RateLimitTriggerCooldown(target))
		if err != nil {
			return "", err
		}
		if err := failed.Error(); err != nil {
			return "", err
		}

		failed, err = s.RateLimiter.Allow(ctx, kind.RateLimitTriggerPerIP(string(s.RemoteIP)))
		if err != nil {
			return "", err
		}
		if err := failed.Error(); err != nil {
			return "", err
		}

		if opts.UserID != "" {
			failed, err := s.RateLimiter.Allow(ctx, kind.RateLimitTriggerPerUser(opts.UserID))
			if err != nil {
				return "", err
			}
			if err := failed.Error(); err != nil {
				return "", err
			}
		}
	}

	code := &Code{
		Target:   target,
		Purpose:  kind.Purpose(),
		Form:     form,
		Code:     form.GenerateCode(s.TestModeConfig, s.TestModeFeatureConfig, target, opts.UserID),
		ExpireAt: s.Clock.NowUTC().Add(kind.ValidPeriod()),

		UserID:                                 opts.UserID,
		WorkflowID:                             opts.WorkflowID,
		AuthenticationFlowWebsocketChannelName: opts.AuthenticationFlowWebsocketChannelName,
		AuthenticationFlowType:                 opts.AuthenticationFlowType,
		AuthenticationFlowName:                 opts.AuthenticationFlowName,
		AuthenticationFlowJSONPointer:          opts.AuthenticationFlowJSONPointer,
		WebSessionID:                           opts.WebSessionID,
	}

	err := s.CodeStore.Create(ctx, kind.Purpose(), code)
	if err != nil {
		return "", err
	}

	if form.AllowLookupByCode() {
		err := s.LookupStore.Create(ctx, code.Purpose, code.Code, code.Target, code.ExpireAt)
		if err != nil {
			return "", err
		}
	}

	if err := s.AttemptTracker.ResetFailedAttempts(ctx, kind, target); err != nil {
		// non-critical error; log and continue
		s.Logger.WithError(err).Warn("failed to reset failed attempts counter")
	}

	return code.Code, nil
}

func (s *Service) VerifyOTP(ctx context.Context, kind Kind, target string, otp string, opts *VerifyOptions) error {
	if err := s.checkFailedAttemptsRevocation(ctx, kind, target); err != nil {
		return err
	}

	var reservations []*ratelimit.Reservation

	isCodeValid := false
	defer func() {
		if isCodeValid {
			for _, r := range reservations {
				s.RateLimiter.Cancel(ctx, r)
			}
		}
	}()

	r1, failedR1, err := s.RateLimiter.Reserve(ctx, kind.RateLimitValidatePerIP(string(s.RemoteIP)))
	if err != nil {
		return err
	}
	if err := failedR1.Error(); err != nil {
		return err
	}
	reservations = append(reservations, r1)

	if opts.UserID != "" {
		r2, failedR2, err := s.RateLimiter.Reserve(ctx, kind.RateLimitValidatePerUserPerIP(opts.UserID, string(s.RemoteIP)))
		if err != nil {
			return err
		}
		if err := failedR2.Error(); err != nil {
			return err
		}
		reservations = append(reservations, r2)
	}

	code, err := s.getCode(ctx, kind.Purpose(), target)
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
		ferr := s.handleFailedAttemptsRevocation(ctx, kind, target)
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
		if err := s.consumeCode(ctx, kind.Purpose(), code); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) ConsumeCode(ctx context.Context, purpose Purpose, target string) error {
	code, err := s.getCode(ctx, purpose, target)
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

	return s.consumeCode(ctx, purpose, code)
}

func (s *Service) SetSubmittedCode(ctx context.Context, kind Kind, target string, code string) (*State, error) {
	codeModel, err := s.getCode(ctx, kind.Purpose(), target)
	if err != nil {
		return nil, err
	}

	codeModel.UserInputtedCode = code
	if err := s.CodeStore.Update(ctx, kind.Purpose(), codeModel); err != nil {
		return nil, err
	}

	return s.InspectState(ctx, kind, target)
}

func (s *Service) LookupCode(ctx context.Context, purpose Purpose, code string) (target string, err error) {
	return s.LookupStore.Get(ctx, purpose, code)
}

func (s *Service) InspectCode(ctx context.Context, purpose Purpose, target string) (*Code, error) {
	return s.getCode(ctx, purpose, target)
}

func (s *Service) InspectState(ctx context.Context, kind Kind, target string) (*State, error) {
	ferr := s.checkFailedAttemptsRevocation(ctx, kind, target)
	tooManyAttempts := false
	if errors.Is(ferr, ErrTooManyAttempts) {
		tooManyAttempts = true
	} else if ferr != nil {
		return nil, ferr
	}

	// This is intentionally zero.
	var canResendAt time.Time
	timeToAct, err := s.RateLimiter.GetTimeToAct(ctx, kind.RateLimitTriggerCooldown(target))
	if err != nil {
		return nil, err
	}
	canResendAt = *timeToAct

	now := s.Clock.NowUTC()

	state := &State{
		ExpireAt:        now,
		CanResendAt:     canResendAt,
		SubmittedCode:   "",
		TooManyAttempts: tooManyAttempts,
	}

	code, err := s.getCode(ctx, kind.Purpose(), target)
	if errors.Is(err, ErrCodeNotFound) {
		code = nil
	} else if err != nil {
		return nil, err
	}

	if code != nil && code.Purpose != kind.Purpose() {
		return nil, ErrCodeNotFound
	}
	if code != nil && code.Consumed {
		// Treat consumed code as not found.
		code = nil
	}

	if code != nil {
		state.ExpireAt = code.ExpireAt
		state.SubmittedCode = code.UserInputtedCode
		state.UserID = code.UserID
		state.WorkflowID = code.WorkflowID
		state.AuthenticationFlowWebsocketChannelName = code.AuthenticationFlowWebsocketChannelName
		state.AuthenticationFlowJSONPointer = code.AuthenticationFlowJSONPointer
		state.AuthenticationFlowName = code.AuthenticationFlowName
		state.AuthenticationFlowType = code.AuthenticationFlowType
		state.WebSessionID = code.WebSessionID
	}

	return state, nil
}
