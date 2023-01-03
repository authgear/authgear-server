package otp

import (
	"errors"
	"time"

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

func (s *Service) createCode(target string, expireAt time.Time) (*Code, error) {
	code := secretcode.OOBOTPSecretCode.Generate()

	codeModel := &Code{
		Code:     code,
		ExpireAt: expireAt,
	}

	err := s.CodeStore.Create(target, codeModel)
	if err != nil {
		return nil, err
	}

	return codeModel, nil
}

func (s *Service) GenerateCode(target string, expireAt time.Time) (*Code, error) {
	code, err := s.getCode(target)
	if errors.Is(err, ErrCodeNotFound) {
		code = nil
	} else if err != nil {
		return nil, err
	}

	if code == nil || s.Clock.NowUTC().After(code.ExpireAt) {
		code, err = s.createCode(target, expireAt)
		if err != nil {
			return nil, err
		}
	}

	return code, nil
}

func (s *Service) VerifyCode(target string, code string) error {
	codeModel, err := s.CodeStore.Get(target)
	if errors.Is(err, ErrCodeNotFound) {
		return ErrInvalidCode
	} else if err != nil {
		return err
	}

	if !secretcode.OOBOTPSecretCode.Compare(code, codeModel.Code) {
		return ErrInvalidCode
	}

	if err = s.CodeStore.Delete(target); err != nil {
		s.Logger.WithError(err).Error("failed to delete code after validation")
	}

	return nil
}

func (s *Service) SetUserInputtedCode(target string, userInputtedCode string) (*Code, error) {
	codeModel, err := s.CodeStore.Get(target)
	if err != nil {
		return nil, err
	}

	codeModel.UserInputtedCode = userInputtedCode
	if err := s.CodeStore.Update(target, codeModel); err != nil {
		return nil, err
	}

	return codeModel, nil
}
