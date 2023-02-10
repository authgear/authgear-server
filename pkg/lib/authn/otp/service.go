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
		codeModel.Code = secretcode.MagicLinkOTPSecretCode.Generate()
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

func (s *Service) GenerateCode(target string, otpMode OTPMode, appID string, webSessionID string) (*Code, error) {
	return s.createCode(target, otpMode, &Code{
		AppID:        appID,
		WebSessionID: webSessionID,
	})
}

func (s *Service) GenerateWhatsappCode(target string, appID string, webSessionID string) (*Code, error) {
	return s.createCode(target, OTPModeCode, &Code{
		AppID:        appID,
		WebSessionID: webSessionID,
	})
}

func (s *Service) CanVerifyCode(target string) (bool, error) {
	_, err := s.getCode(target)
	if errors.Is(err, ErrCodeNotFound) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (s *Service) VerifyCode(target string, code string) error {
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

func (s *Service) VerifyMagicLinkCode(code string, consume bool) (*Code, error) {
	target, err := s.MagicLinkStore.Get(code)
	if errors.Is(err, ErrCodeNotFound) {
		return nil, ErrInvalidMagicLink
	} else if err != nil {
		return nil, err
	}
	return s.VerifyMagicLinkCodeByTarget(target, consume)
}

func (s *Service) VerifyMagicLinkCodeByTarget(target string, consume bool) (*Code, error) {
	codeModel, err := s.getCode(target)
	if errors.Is(err, ErrCodeNotFound) {
		return nil, ErrInvalidMagicLink
	} else if err != nil {
		return nil, err
	}

	userInputtedCode := codeModel.UserInputtedCode
	if userInputtedCode == "" {
		userInputtedCode = code
	}

	if !secretcode.MagicLinkOTPSecretCode.Compare(userInputtedCode, codeModel.Code) {
		return nil, ErrInvalidCode
	}

	if consume {
		s.deleteCode(codeModel.Code)
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

func (s *Service) SetUserInputtedMagicLinkCode(userInputtedCode string) (*Code, error) {
	target, err := s.MagicLinkStore.Get(userInputtedCode)
	if err != nil {
		return nil, ErrInvalidMagicLink
	}

	codeModel, err := s.getCode(target)
	if errors.Is(err, ErrCodeNotFound) {
		return nil, ErrInvalidMagicLink
	} else if err != nil {
		return nil, err
	}

	codeModel.UserInputtedCode = userInputtedCode
	if err := s.CodeStore.Update(codeModel.Target, codeModel); err != nil {
		return nil, err
	}

	return codeModel, nil
}
