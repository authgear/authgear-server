package otp

import (
	"errors"
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/secretcode"
)

type GenerateCodeOptions struct {
	UserID       string
	WebSessionID string
	WorkflowID   string
}

type GenerateOptions struct {
	UserID     string
	WorkflowID string
}

type VerifyOptions struct {
	UserID           string
	UseSubmittedCode bool
	SkipConsume      bool
}

type CodeStore interface {
	Create(target string, code *Code) error
	Get(target string) (*Code, error)
	Update(target string, code *Code) error
	Delete(target string) error
}

type LoginLinkStore interface {
	Create(token string, target string, expireAt time.Time) error
	Get(token string) (string, error)
	Delete(token string) error
}

type LookupStore interface {
	Create(purpose string, code string, target string, expireAt time.Time) error
	Get(purpose string, code string) (string, error)
	Delete(purpose string, code string) error
}

type AttemptTracker interface {
	ResetFailedAttempts(purpose string, target string) error
	GetFailedAttempts(purpose string, target string) (int, error)
	IncrementFailedAttempts(purpose string, target string) (int, error)
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("otp")} }

type Service struct {
	Clock clock.Clock

	AppID          config.AppID
	RemoteIP       httputil.RemoteIP
	CodeStore      CodeStore
	LoginLinkStore LoginLinkStore
	LookupStore    LookupStore
	AttemptTracker AttemptTracker
	Logger         Logger
	RateLimiter    RateLimiter
	OTPConfig      *config.OTPLegacyConfig
	Verification   *config.VerificationConfig
}

func (s *Service) isFailedAttemptRatelimitEnabled() bool {
	return s.OTPConfig.Ratelimit.FailedAttempt.Enabled
}

func (s *Service) TrackFailedAttemptBucket(target string) ratelimit.Bucket {
	config := s.OTPConfig.Ratelimit.FailedAttempt
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("otp-failed-attempt:%s", target),
		Size:        config.Size,
		ResetPeriod: config.ResetPeriod.Duration(),
		Name:        ratelimit.TrackFailedOTPAttemptBucketName,
	}
}

func (s *Service) getCode(target string) (*Code, error) {
	return s.CodeStore.Get(target)
}

func (s *Service) createCode(target string, otpMode OTPMode, codeModel *Code) (*Code, error) {
	if codeModel == nil {
		codeModel = &Code{}
	}
	codeModel.Target = target
	codeModel.ExpireAt = s.Clock.NowUTC().Add(s.Verification.CodeValidPeriod.Duration())

	switch otpMode {
	case OTPModeLoginLink:
		codeModel.Code = secretcode.LinkOTPSecretCode.Generate()
		err := s.LoginLinkStore.Create(codeModel.Code, codeModel.Target, codeModel.ExpireAt)
		if err != nil {
			return nil, err
		}
		err = s.CodeStore.Create(target, codeModel)
		if err != nil {
			return nil, err
		}
	default:
		codeModel.Code = secretcode.OOBOTPSecretCode.Generate()
		err := s.CodeStore.Create(target, codeModel)
		if err != nil {
			return nil, err
		}
	}

	// Reset failed attempt count
	if s.isFailedAttemptRatelimitEnabled() {
		err := s.RateLimiter.ClearBucket(s.TrackFailedAttemptBucket(target))
		if err != nil {
			return nil, err
		}
	}

	return codeModel, nil
}

func (s *Service) deleteCode(target string) {
	if err := s.CodeStore.Delete(target); err != nil {
		s.Logger.WithError(err).Error("failed to delete code after validation")
	}
	// No need delete from lookup store;
	// lookup entry is invalidated since target is no longer exist.
}

func (s *Service) handleFailedAttempt(target string) error {
	if s.isFailedAttemptRatelimitEnabled() {
		err := s.RateLimiter.TakeToken(s.TrackFailedAttemptBucket(target))
		if err != nil {
			return err
		}

		pass, _, err := s.RateLimiter.CheckToken(s.TrackFailedAttemptBucket(target))
		if err != nil {
			return err
		} else if !pass {
			// Maximum number of failed attempt exceeded
			s.deleteCode(target)
		}
	}

	return ErrInvalidCode
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
		return ErrTooManyAttempts
	}

	return nil
}

func (s *Service) GenerateOTP(kind Kind, target string, form Form, opts *GenerateOptions) (string, error) {
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

	code := &Code{
		AppID:      string(s.AppID),
		Target:     target,
		WorkflowID: opts.WorkflowID,
	}
	code.Target = target
	code.Purpose = kind.Purpose()
	code.Form = form
	code.Code = form.GenerateCode()
	code.ExpireAt = s.Clock.NowUTC().Add(kind.ValidPeriod())

	// TODO: lookup-able code

	err := s.CodeStore.Create(target, code)
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

func (s *Service) GenerateCode(target string, otpMode OTPMode, opt *GenerateCodeOptions) (*Code, error) {
	return s.createCode(target, otpMode, &Code{
		AppID:        string(s.AppID),
		WebSessionID: opt.WebSessionID,
		WorkflowID:   opt.WorkflowID,
	})
}

func (s *Service) GenerateWhatsappCode(target string, opt *GenerateCodeOptions) (*Code, error) {
	return s.createCode(target, OTPModeCode, &Code{
		AppID:        string(s.AppID),
		WebSessionID: opt.WebSessionID,
		WorkflowID:   opt.WorkflowID,
	})
}

func (s *Service) FailedAttemptRateLimitExceeded(target string) (bool, error) {
	if !s.isFailedAttemptRatelimitEnabled() {
		return false, nil
	}

	pass, _, err := s.RateLimiter.CheckToken(s.TrackFailedAttemptBucket(target))
	if err != nil {
		return false, err
	}
	if !pass {
		return true, nil
	}

	// We do not check the presence of the code here.
	// If we were to check that, we will have the following bug.
	// 1. Sign in.
	// 2. Sign in immediately again within OTP cooldown period.
	// 3. The code is not generated (thus absent), and we DO NOT report rate limit error.
	// 4. This function return true, the client is confused that failed attempt rate limit is exceeded.

	return false, nil
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

	code, err := s.getCode(target)
	if errors.Is(err, ErrCodeNotFound) {
		return ErrInvalidCode
	} else if err != nil {
		return err
	}

	if code.Purpose != kind.Purpose() {
		return ErrInvalidCode
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
		s.deleteCode(target)
	}

	return nil
}

func (s *Service) VerifyCode(target string, code string) error {
	if s.isFailedAttemptRatelimitEnabled() {
		bucket := s.TrackFailedAttemptBucket(target)
		pass, _, err := s.RateLimiter.CheckToken(bucket)
		if err != nil {
			return err
		}
		if !pass {
			return bucket.BucketError()
		}
	}

	codeModel, err := s.getCode(target)
	if errors.Is(err, ErrCodeNotFound) {
		return ErrInvalidCode
	} else if err != nil {
		return err
	}

	if !secretcode.OOBOTPSecretCode.Compare(code, codeModel.Code) {
		return s.handleFailedAttempt(target)
	}

	s.deleteCode(target)

	return nil
}

// VerifyLoginLinkCode verifies the code but it won't consume it
func (s *Service) VerifyLoginLinkCode(userInputtedCode string) (*Code, error) {
	target, err := s.LoginLinkStore.Get(userInputtedCode)
	if errors.Is(err, ErrCodeNotFound) {
		return nil, ErrInvalidLoginLink
	} else if err != nil {
		return nil, err
	}

	codeModel, err := s.getCode(target)
	if errors.Is(err, ErrCodeNotFound) {
		return nil, ErrInvalidLoginLink
	} else if err != nil {
		return nil, err
	}

	if !secretcode.LinkOTPSecretCode.Compare(userInputtedCode, codeModel.Code) {
		return nil, ErrInvalidLoginLink
	}

	return codeModel, nil
}

func (s *Service) VerifyLoginLinkCodeByTarget(target string, consume bool) (*Code, error) {
	codeModel, err := s.getCode(target)
	if errors.Is(err, ErrCodeNotFound) {
		return nil, ErrInvalidLoginLink
	} else if err != nil {
		return nil, err
	}

	if !secretcode.LinkOTPSecretCode.Compare(codeModel.UserInputtedCode, codeModel.Code) {
		return nil, ErrInvalidLoginLink
	}

	if consume {
		s.deleteCode(codeModel.Target)
	}

	return codeModel, nil
}

func (s *Service) VerifyWhatsappCode(target string, consume bool) error {
	codeModel, err := s.getCode(target)
	if errors.Is(err, ErrCodeNotFound) {
		return ErrInvalidCode
	} else if err != nil {
		return err
	}

	if codeModel.UserInputtedCode == "" {
		return ErrInputRequired
	}

	if !secretcode.OOBOTPSecretCode.Compare(codeModel.UserInputtedCode, codeModel.Code) {
		return s.handleFailedAttempt(target)
	}

	if consume {
		s.deleteCode(target)
	}

	return nil
}

// SetUserInputtedCode set the user inputted code without verifying it
// The code will be verified via VerifyWhatsappCode in the original interaction
func (s *Service) SetUserInputtedCode(target string, userInputtedCode string) (*Code, error) {
	codeModel, err := s.getCode(target)
	if err != nil {
		return nil, err
	}

	codeModel.UserInputtedCode = userInputtedCode
	if err := s.CodeStore.Update(target, codeModel); err != nil {
		return nil, err
	}

	return codeModel, nil
}

// SetUserInputtedLoginLinkCode set the user inputted code if the code is correct
// If the code is incorrect, error will be returned and the approval screen should show
// the error to the user
// If the code is correct, the code will be set to the user inputted code
// The code should be verified again via VerifyLoginLinkCodeByTarget in the original interaction
func (s *Service) SetUserInputtedLoginLinkCode(userInputtedCode string) (*Code, error) {
	codeModel, err := s.VerifyLoginLinkCode(userInputtedCode)
	if err != nil {
		return nil, err
	}

	codeModel.UserInputtedCode = userInputtedCode
	if err := s.CodeStore.Update(codeModel.Target, codeModel); err != nil {
		return nil, err
	}

	return codeModel, nil
}

func (s *Service) SetSubmittedCode(kind Kind, target string, code string) (*State, error) {
	codeModel, err := s.getCode(target)
	if err != nil {
		return nil, err
	}

	codeModel.UserInputtedCode = code
	if err := s.CodeStore.Update(target, codeModel); err != nil {
		return nil, err
	}

	return s.InspectState(kind, target)
}

func (s *Service) LookupCode(kind Kind, code string) (target string, err error) {
	return s.LookupStore.Get(kind.Purpose(), code)
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

	code, err := s.getCode(target)
	if errors.Is(err, ErrCodeNotFound) {
		code = nil
	} else if err != nil {
		return nil, err
	}

	if code != nil && code.Purpose != kind.Purpose() {
		return nil, ErrCodeNotFound
	}

	if code != nil {
		state.ExpireAt = code.ExpireAt
		state.SubmittedCode = code.UserInputtedCode
		state.WorkflowID = code.WorkflowID
	}

	return state, nil
}
