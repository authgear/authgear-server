package otp

import (
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/secretcode"
)

type CodeStore interface {
	Create(target string, code *Code) error
	Get(target string) (*Code, error)
	Update(target string, code *Code) error
	Delete(target string) error
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("otp")} }

type Service struct {
	Clock clock.Clock

	CodeStore   CodeStore
	Logger      Logger
	RateLimiter RateLimiter
}

func TrackFailedAttemptBucket(target string) ratelimit.Bucket {
	return ratelimit.Bucket{
		Key:         fmt.Sprintf("otp-failed-attempt:%s", target),
		Size:        5,
		ResetPeriod: duration.UserInteraction,
	}
}

func (s *Service) getCode(target string) (*Code, error) {
	return s.CodeStore.Get(target)
}

func (s *Service) createCode(target string, codeModel *Code) (*Code, error) {
	if codeModel == nil {
		codeModel = &Code{}
	}
	codeModel.Target = target
	codeModel.Code = secretcode.OOBOTPSecretCode.Generate()
	codeModel.ExpireAt = s.Clock.NowUTC().Add(duration.UserInteraction)

	err := s.CodeStore.Create(target, codeModel)
	if err != nil {
		return nil, err
	}

	// Reset failed attempt count
	err = s.RateLimiter.ClearBucket(TrackFailedAttemptBucket(target))
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
	err := s.RateLimiter.TakeToken(TrackFailedAttemptBucket(target))
	if err != nil {
		return err
	}

	pass, _, err := s.RateLimiter.CheckToken(TrackFailedAttemptBucket(target))
	if err != nil {
		return err
	} else if !pass {
		// Maximum number of failed attempt exceeded
		s.deleteCode(target)
	}
	return ErrInvalidCode
}

func (s *Service) GenerateCode(target string) (*Code, error) {
	return s.createCode(target, nil)
}

func (s *Service) GenerateWhatsappCode(target string, appID string, webSessionID string) (*Code, error) {
	return s.createCode(target, &Code{
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
