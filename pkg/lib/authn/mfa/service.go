package mfa

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/secretcode"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type StoreDeviceToken interface {
	Get(ctx context.Context, userID string, token string) (*DeviceToken, error)
	Create(ctx context.Context, token *DeviceToken) error
	DeleteAll(ctx context.Context, userID string) error
	HasTokens(ctx context.Context, userID string) (bool, error)
	Count(ctx context.Context, userID string) (int, error)
}

type StoreRecoveryCode interface {
	List(ctx context.Context, userID string) ([]*RecoveryCode, error)
	Get(ctx context.Context, userID string, code string) (*RecoveryCode, error)
	DeleteAll(ctx context.Context, userID string) error
	CreateAll(ctx context.Context, codes []*RecoveryCode) error
	UpdateConsumed(ctx context.Context, code *RecoveryCode) error
}

type RateLimiter interface {
	Reserve(ctx context.Context, spec ratelimit.BucketSpec) (*ratelimit.Reservation, *ratelimit.FailedReservation, error)
	Cancel(ctx context.Context, r *ratelimit.Reservation)
}

type Service struct {
	ReadOnlyService
	IP            httputil.RemoteIP
	DeviceTokens  StoreDeviceToken
	RecoveryCodes StoreRecoveryCode
	Clock         clock.Clock
	Config        *config.AppConfig
	FeatureConfig *config.FeatureConfig
	EnvConfig     *config.RateLimitsEnvironmentConfig
	RateLimiter   RateLimiter
	Lockout       Lockout
}

func (s *Service) GenerateDeviceToken(ctx context.Context) string {
	return GenerateDeviceToken()
}

func (s *Service) reserveRateLimit(
	ctx context.Context,
	rl ratelimit.RateLimitGroup,
	userID string,
) (reservations []*ratelimit.Reservation, err error) {
	specs := rl.ResolveBucketSpecs(s.Config, s.FeatureConfig, s.EnvConfig, &ratelimit.ResolveBucketSpecOptions{
		IPAddress: string(s.IP),
		UserID:    userID,
	})
	for _, spec := range specs {
		spec := *spec
		resv, failedResv, resvErr := s.RateLimiter.Reserve(ctx, spec)
		err = resvErr
		if err != nil {
			return
		}
		if ratelimitErr := failedResv.Error(); ratelimitErr != nil {
			err = ratelimitErr
			return
		}
		reservations = append(reservations, resv)
	}

	return
}

func (s *Service) CreateDeviceToken(ctx context.Context, userID string, token string) (*DeviceToken, error) {
	t := &DeviceToken{
		UserID:    userID,
		Token:     token,
		CreatedAt: s.Clock.NowUTC(),
		ExpireAt:  s.Clock.NowUTC().Add(s.Config.Authentication.DeviceToken.ExpireIn.Duration()),
	}

	if err := s.DeviceTokens.Create(ctx, t); err != nil {
		return nil, err
	}

	return t, nil
}

func (s *Service) VerifyDeviceToken(ctx context.Context, userID string, token string) error {
	resvs, err := s.reserveRateLimit(
		ctx,
		ratelimit.RateLimitGroupAuthenticationDeviceToken,
		userID,
	)
	if err != nil {
		return err
	}

	defer func() {
		for _, resv := range resvs {
			s.RateLimiter.Cancel(ctx, resv)
		}
	}()

	_, err = s.DeviceTokens.Get(ctx, userID, token)
	if errors.Is(err, ErrDeviceTokenNotFound) {
		for _, resv := range resvs {
			resv.PreventCancel()
		}
	}
	return err
}

func (s *Service) InvalidateAllDeviceTokens(ctx context.Context, userID string) error {
	return s.DeviceTokens.DeleteAll(ctx, userID)
}

func (s *Service) HasDeviceTokens(ctx context.Context, userID string) (bool, error) {
	return s.DeviceTokens.HasTokens(ctx, userID)
}

func (s *Service) CountDeviceTokens(ctx context.Context, userID string) (int, error) {
	return s.DeviceTokens.Count(ctx, userID)
}

func (s *Service) GenerateRecoveryCodes(ctx context.Context) []string {
	codes := make([]string, s.Config.Authentication.RecoveryCode.Count)
	for i := range codes {
		codes[i] = secretcode.RecoveryCode.Generate()
	}
	return codes
}

func (s *Service) InvalidateAllRecoveryCode(ctx context.Context, userID string) error {
	return s.RecoveryCodes.DeleteAll(ctx, userID)
}

func (s *Service) ReplaceRecoveryCodes(ctx context.Context, userID string, codes []string) ([]*RecoveryCode, error) {
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

	if err := s.RecoveryCodes.DeleteAll(ctx, userID); err != nil {
		return nil, err
	}
	if err := s.RecoveryCodes.CreateAll(ctx, codeModels); err != nil {
		return nil, err
	}

	return codeModels, nil
}

func (s *Service) VerifyRecoveryCode(ctx context.Context, userID string, code string) (*RecoveryCode, error) {
	resvs, err := s.reserveRateLimit(
		ctx,
		ratelimit.RateLimitGroupAuthenticationRecoveryCode,
		userID,
	)
	if err != nil {
		return nil, err
	}

	defer func() {
		for _, resv := range resvs {
			s.RateLimiter.Cancel(ctx, resv)
		}
	}()

	err = s.Lockout.Check(ctx, userID)
	if err != nil {
		return nil, err
	}

	code, err = secretcode.RecoveryCode.FormatForComparison(code)
	if err != nil {
		return nil, ErrRecoveryCodeNotFound
	}

	rc, err := s.RecoveryCodes.Get(ctx, userID, code)
	if errors.Is(err, ErrRecoveryCodeNotFound) {
		for _, resv := range resvs {
			resv.PreventCancel()
		}
		aerr := s.Lockout.MakeRecoveryCodeAttempt(ctx, userID, 1)
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

func (s *Service) ConsumeRecoveryCode(ctx context.Context, rc *RecoveryCode) error {
	rc.Consumed = true
	rc.UpdatedAt = s.Clock.NowUTC()

	if err := s.RecoveryCodes.UpdateConsumed(ctx, rc); err != nil {
		return err
	}

	return nil
}
