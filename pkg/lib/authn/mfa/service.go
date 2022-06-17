package mfa

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/secretcode"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type StoreDeviceToken interface {
	Get(userID string, token string) (*DeviceToken, error)
	Create(token *DeviceToken) error
	DeleteAll(userID string) error
	HasTokens(userID string) (bool, error)
}

type StoreRecoveryCode interface {
	List(userID string) ([]*RecoveryCode, error)
	Get(userID string, code string) (*RecoveryCode, error)
	DeleteAll(userID string) error
	CreateAll(codes []*RecoveryCode) error
	UpdateConsumed(code *RecoveryCode) error
}

type RateLimiter interface {
	TakeToken(bucket ratelimit.Bucket) error
}

type Service struct {
	DeviceTokens  StoreDeviceToken
	RecoveryCodes StoreRecoveryCode
	Clock         clock.Clock
	Config        *config.AuthenticationConfig
	RateLimiter   RateLimiter
}

func (s *Service) GenerateDeviceToken() string {
	return GenerateDeviceToken()
}

func (s *Service) CreateDeviceToken(userID string, token string) (*DeviceToken, error) {
	t := &DeviceToken{
		UserID:    userID,
		Token:     token,
		CreatedAt: s.Clock.NowUTC(),
		ExpireAt:  s.Clock.NowUTC().Add(s.Config.DeviceToken.ExpireIn.Duration()),
	}

	if err := s.DeviceTokens.Create(t); err != nil {
		return nil, err
	}

	return t, nil
}

func (s *Service) VerifyDeviceToken(userID string, token string) error {
	err := s.RateLimiter.TakeToken(AntiBruteForceDeviceTokenBucket(userID))
	if err != nil {
		return err
	}

	_, err = s.DeviceTokens.Get(userID, token)
	return err
}

func (s *Service) InvalidateAllDeviceTokens(userID string) error {
	return s.DeviceTokens.DeleteAll(userID)
}

func (s *Service) HasDeviceTokens(userID string) (bool, error) {
	return s.DeviceTokens.HasTokens(userID)
}

func (s *Service) GenerateRecoveryCodes() []string {
	codes := make([]string, s.Config.RecoveryCode.Count)
	for i := range codes {
		codes[i] = secretcode.RecoveryCode.Generate()
	}
	return codes
}

func (s *Service) InvalidateAllRecoveryCode(userID string) error {
	return s.RecoveryCodes.DeleteAll(userID)
}

func (s *Service) ReplaceRecoveryCodes(userID string, codes []string) ([]*RecoveryCode, error) {
	codeModels := make([]*RecoveryCode, len(codes))
	now := s.Clock.NowUTC()
	for i, code := range codes {
		codeModels[i] = &RecoveryCode{
			ID:        uuid.New(),
			UserID:    userID,
			Code:      code,
			CreatedAt: now,
			UpdatedAt: now,
			Consumed:  false,
		}
	}

	if err := s.RecoveryCodes.DeleteAll(userID); err != nil {
		return nil, err
	}
	if err := s.RecoveryCodes.CreateAll(codeModels); err != nil {
		return nil, err
	}

	return codeModels, nil
}

func (s *Service) VerifyRecoveryCode(userID string, code string) (*RecoveryCode, error) {
	err := s.RateLimiter.TakeToken(AutiBruteForceRecoveryCodeBucket(userID))
	if err != nil {
		return nil, err
	}

	code, err = secretcode.RecoveryCode.FormatForComparison(code)
	if err != nil {
		err = ErrRecoveryCodeNotFound
		return nil, err
	}

	rc, err := s.RecoveryCodes.Get(userID, code)
	if err != nil {
		return nil, err
	}

	if rc.Consumed {
		return nil, ErrRecoveryCodeConsumed
	}

	return rc, nil
}

func (s *Service) ConsumeRecoveryCode(rc *RecoveryCode) error {
	rc.Consumed = true
	rc.UpdatedAt = s.Clock.NowUTC()

	if err := s.RecoveryCodes.UpdateConsumed(rc); err != nil {
		return err
	}

	return nil
}

func (s *Service) ListRecoveryCodes(userID string) ([]*RecoveryCode, error) {
	return s.RecoveryCodes.List(userID)
}
