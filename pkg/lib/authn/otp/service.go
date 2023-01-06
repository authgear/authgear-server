package otp

import (
	"errors"

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

	CodeStore CodeStore
	Logger    Logger
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

	return codeModel, nil
}

func (s *Service) deleteCode(target string) {
	if err := s.CodeStore.Delete(target); err != nil {
		s.Logger.WithError(err).Error("failed to delete code after validation")
	}
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

func (s *Service) VerifyCode(target string, code string) error {
	codeModel, err := s.getCode(target)
	if errors.Is(err, ErrCodeNotFound) {
		return ErrInvalidCode
	} else if err != nil {
		return err
	}

	if !secretcode.OOBOTPSecretCode.Compare(code, codeModel.Code) {
		return ErrInvalidCode
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
		return ErrInvalidCode
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
