package mfa

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/secretcode"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

const (
	VerifyDeviceTokenPerUserPerIP  ratelimit.BucketName = "VerifyDeviceTokenPerUserPerIP"
	VerifyDeviceTokenPerIP         ratelimit.BucketName = "VerifyDeviceTokenPerIP"
	VerifyRecoveryCodePerUserPerIP ratelimit.BucketName = "VerifyRecoveryCodePerUserPerIP"
	VerifyRecoveryCodePerIP        ratelimit.BucketName = "VerifyRecoveryCodePerIP"
)

type StoreDeviceToken interface {
	Get(userID string, token string) (*DeviceToken, error)
	Create(token *DeviceToken) error
	DeleteAll(userID string) error
	HasTokens(userID string) (bool, error)
	Count(userID string) (int, error)
}

type StoreRecoveryCode interface {
	List(userID string) ([]*RecoveryCode, error)
	Get(userID string, code string) (*RecoveryCode, error)
	DeleteAll(userID string) error
	CreateAll(codes []*RecoveryCode) error
	UpdateConsumed(code *RecoveryCode) error
}

type RateLimiter interface {
	Reserve(spec ratelimit.BucketSpec) *ratelimit.Reservation
	Cancel(r *ratelimit.Reservation)
}

type Service struct {
	IP            httputil.RemoteIP
	DeviceTokens  StoreDeviceToken
	RecoveryCodes StoreRecoveryCode
	Clock         clock.Clock
	Config        *config.AuthenticationConfig
	RateLimiter   RateLimiter
	Lockout       Lockout
}

func (s *Service) GenerateDeviceToken() string {
	return GenerateDeviceToken()
}

func (s *Service) reserveRateLimit(
	namePerUserPerIP ratelimit.BucketName,
	perUserPerIP *config.RateLimitConfig,
	namePerIP ratelimit.BucketName,
	perIP *config.RateLimitConfig,
	userID string,
) (rPerUserPerIP *ratelimit.Reservation, rPerIP *ratelimit.Reservation, err error) {
	if perUserPerIP.Enabled == nil {
		perUserPerIP = s.Config.RateLimits.General.PerUserPerIP
	}
	if perIP.Enabled == nil {
		perIP = s.Config.RateLimits.General.PerIP
	}

	rPerUserPerIP = s.RateLimiter.Reserve(ratelimit.NewBucketSpec(
		perUserPerIP, namePerUserPerIP,
		userID, string(s.IP),
	))
	if err = rPerUserPerIP.Error(); err != nil {
		return
	}

	rPerIP = s.RateLimiter.Reserve(ratelimit.NewBucketSpec(
		perIP, namePerIP,
		string(s.IP),
	))
	if err = rPerIP.Error(); err != nil {
		return
	}

	return
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
	perUserPerIP, perIP, err := s.reserveRateLimit(
		VerifyDeviceTokenPerUserPerIP,
		s.Config.RateLimits.DeviceToken.PerUserPerIP,
		VerifyDeviceTokenPerIP,
		s.Config.RateLimits.DeviceToken.PerIP,
		userID,
	)
	defer s.RateLimiter.Cancel(perUserPerIP)
	defer s.RateLimiter.Cancel(perIP)

	if err != nil {
		return err
	}

	_, err = s.DeviceTokens.Get(userID, token)
	if errors.Is(err, ErrDeviceTokenNotFound) {
		perUserPerIP.Consume()
		perIP.Consume()
	}
	return err
}

func (s *Service) InvalidateAllDeviceTokens(userID string) error {
	return s.DeviceTokens.DeleteAll(userID)
}

func (s *Service) HasDeviceTokens(userID string) (bool, error) {
	return s.DeviceTokens.HasTokens(userID)
}

func (s *Service) CountDeviceTokens(userID string) (int, error) {
	return s.DeviceTokens.Count(userID)
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
	perUserPerIP, perIP, err := s.reserveRateLimit(
		VerifyRecoveryCodePerUserPerIP,
		s.Config.RateLimits.RecoveryCode.PerUserPerIP,
		VerifyRecoveryCodePerIP,
		s.Config.RateLimits.RecoveryCode.PerIP,
		userID,
	)
	defer s.RateLimiter.Cancel(perUserPerIP)
	defer s.RateLimiter.Cancel(perIP)

	if err != nil {
		return nil, err
	}

	err = s.Lockout.Check(userID)
	if err != nil {
		return nil, err
	}

	code, err = secretcode.RecoveryCode.FormatForComparison(code)
	if err != nil {
		return nil, ErrRecoveryCodeNotFound
	}

	rc, err := s.RecoveryCodes.Get(userID, code)
	if errors.Is(err, ErrRecoveryCodeNotFound) {
		perUserPerIP.Consume()
		perIP.Consume()
		aerr := s.Lockout.MakeRecoveryCodeAttempt(userID, 1)
		if aerr != nil {
			return nil, aerr
		}
		return nil, err
	} else if err != nil {
		return nil, err
	}

	if rc.Consumed {
		// Do not consume the rate limit tokens and record the attempt here,
		// since trying a used recovery code is rare and usually mistakes of real user.
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
