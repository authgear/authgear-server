package otp

import (
	"errors"
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/secretcode"
)

type GenerateCodeOptions struct {
	WebSessionID string
	WorkflowID   string
}

type CodeStore interface {
	Create(target string, code *Code) error
	Get(target string) (*Code, error)
	Update(target string, code *Code) error
	Delete(target string) error
}

type MagicLinkStore interface {
	Create(token string, target string, expireAt time.Time) error
	Get(token string) (string, error)
	Delete(token string) error
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("otp")} }

type Service struct {
	Clock clock.Clock

	AppID          config.AppID
	CodeStore      CodeStore
	MagicLinkStore MagicLinkStore
	Logger         Logger
	RateLimiter    RateLimiter
	OTPConfig      *config.OTPConfig
	Verification   *config.VerificationConfig
}

func (s *Service) TrackFailedAttemptBucket(target string) ratelimit.Bucket {
	config := s.OTPConfig.Ratelimit.FailedAttempt
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("otp-failed-attempt:%s", target),
		Size:        config.Size,
		ResetPeriod: config.ResetPeriod.Duration(),
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
	codeModel.ExpireAt = s.Clock.NowUTC().Add(s.Verification.CodeExpiry.Duration())

	switch otpMode {
	case OTPModeMagicLink:
		codeModel.Code = secretcode.LoginLinkOTPSecretCode.Generate()
		err := s.MagicLinkStore.Create(codeModel.Code, codeModel.Target, codeModel.ExpireAt)
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
	err := s.RateLimiter.ClearBucket(s.TrackFailedAttemptBucket(target))
	if err != nil {
		return nil, err
	}

	return codeModel, nil
}

func (s *Service) deleteCode(target string) {
	if err := s.CodeStore.Delete(target); err != nil {
		s.Logger.WithError(err).Error("failed to delete code after validation")
	}
}

func (s *Service) handleFailedAttempt(target string) error {
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
	return ErrInvalidCode
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

func (s *Service) VerifyCode(target string, code string) error {
	bucket := s.TrackFailedAttemptBucket(target)
	pass, _, err := s.RateLimiter.CheckToken(bucket)
	if err != nil {
		return err
	}
	if !pass {
		return bucket.BucketError()
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

// VerifyMagicLinkCode verifies the code but it won't consume it
func (s *Service) VerifyMagicLinkCode(userInputtedCode string) (*Code, error) {
	target, err := s.MagicLinkStore.Get(userInputtedCode)
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

	if !secretcode.LoginLinkOTPSecretCode.Compare(userInputtedCode, codeModel.Code) {
		return nil, ErrInvalidLoginLink
	}

	return codeModel, nil
}

func (s *Service) VerifyMagicLinkCodeByTarget(target string, consume bool) (*Code, error) {
	codeModel, err := s.getCode(target)
	if errors.Is(err, ErrCodeNotFound) {
		return nil, ErrInvalidLoginLink
	} else if err != nil {
		return nil, err
	}

	if !secretcode.LoginLinkOTPSecretCode.Compare(codeModel.UserInputtedCode, codeModel.Code) {
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

// SetUserInputtedMagicLinkCode set the user inputted code if the code is correct
// If the code is incorrect, error will be returned and the approval screen should show
// the error to the user
// If the code is correct, the code will be set to the user inputted code
// The code should be verified again via VerifyMagicLinkCodeByTarget in the original interaction
func (s *Service) SetUserInputtedMagicLinkCode(userInputtedCode string) (*Code, error) {
	codeModel, err := s.VerifyMagicLinkCode(userInputtedCode)
	if err != nil {
		return nil, err
	}

	codeModel.UserInputtedCode = userInputtedCode
	if err := s.CodeStore.Update(codeModel.Target, codeModel); err != nil {
		return nil, err
	}

	return codeModel, nil
}
