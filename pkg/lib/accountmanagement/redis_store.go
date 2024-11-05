package accountmanagement

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	goredis "github.com/redis/go-redis/v9"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

type RedisStore struct {
	AppID config.AppID
	Redis *appredis.Handle
	Clock clock.Clock
}

type GenerateTokenOptions struct {
	UserID string

	// OAuth
	Alias       string
	MaybeState  string
	RedirectURI string

	// Phone
	IdentityChannel     model.AuthenticatorOOBChannel
	IdentityPhoneNumber string
	// Email
	IdentityEmail string
	// IdentityID for updating identity
	IdentityID string

	// AuthenticatorID for updating authenticator
	AuthenticatorID                   string
	AuthenticatorRecoveryCodes        []string
	AuthenticatorRecoveryCodesCreated bool
	AuthenticatorType                 model.AuthenticatorType

	// TOTP
	AuthenticatorTOTPIssuer           string
	AuthenticatorTOTPEndUserAccountID string
	AuthenticatorTOTPDisplayName      string
	AuthenticatorTOTPSecret           string
	AuthenticatorTOTPVerified         bool

	// OOB OTP
	AuthenticatorOOBOTPChannel  model.AuthenticatorOOBChannel
	AuthenticatorOOBOTPTarget   string
	AuthenticatorOOBOTPVerified bool
}

func (s *RedisStore) GenerateToken(ctx context.Context, options GenerateTokenOptions) (string, error) {
	tokenString := GenerateToken()
	tokenHash := HashToken(tokenString)

	now := s.Clock.NowUTC()
	ttl := duration.UserInteraction
	expireAt := now.Add(ttl)

	var tokenIdentity *TokenIdentity
	if options.IdentityID != "" || options.IdentityPhoneNumber != "" || options.IdentityEmail != "" {
		tokenIdentity = &TokenIdentity{
			IdentityID:  options.IdentityID,
			Channel:     string(options.IdentityChannel),
			PhoneNumber: options.IdentityPhoneNumber,
			Email:       options.IdentityEmail,
		}
	}

	var tokenAuthenticator *TokenAuthenticator
	if options.AuthenticatorID != "" || len(options.AuthenticatorRecoveryCodes) > 0 || options.AuthenticatorTOTPSecret != "" || options.AuthenticatorTOTPVerified || options.AuthenticatorOOBOTPChannel != "" || options.AuthenticatorOOBOTPTarget != "" || options.AuthenticatorOOBOTPVerified {
		tokenAuthenticator = &TokenAuthenticator{
			AuthenticatorID:      options.AuthenticatorID,
			AuthenticatorType:    string(options.AuthenticatorType),
			RecoveryCodes:        options.AuthenticatorRecoveryCodes,
			RecoveryCodesCreated: options.AuthenticatorRecoveryCodesCreated,
			TOTPIssuer:           options.AuthenticatorTOTPIssuer,
			TOTPDisplayName:      options.AuthenticatorTOTPDisplayName,
			TOTPEndUserAccountID: options.AuthenticatorTOTPEndUserAccountID,
			TOTPSecret:           options.AuthenticatorTOTPSecret,
			TOTPVerified:         options.AuthenticatorTOTPVerified,
			OOBOTPChannel:        options.AuthenticatorOOBOTPChannel,
			OOBOTPTarget:         options.AuthenticatorOOBOTPTarget,
			OOBOTPVerified:       options.AuthenticatorOOBOTPVerified,
		}
	}

	token := &Token{
		AppID:     string(s.AppID),
		UserID:    options.UserID,
		TokenHash: tokenHash,
		CreatedAt: &now,
		ExpireAt:  &expireAt,

		// OAuth
		Alias:       options.Alias,
		State:       options.MaybeState,
		RedirectURI: options.RedirectURI,

		// Identity
		Identity: tokenIdentity,

		Authenticator: tokenAuthenticator,
	}

	tokenBytes, err := json.Marshal(token)
	if err != nil {
		return "", err
	}

	tokenKey := tokenKey(token.AppID, token.TokenHash)

	err = s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		_, err = conn.SetNX(ctx, tokenKey, tokenBytes, ttl).Result()
		if errors.Is(err, goredis.Nil) {
			return errors.New("account management token collision")
		} else if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *RedisStore) GetToken(ctx context.Context, tokenStr string) (*Token, error) {
	tokenHash := HashToken(tokenStr)

	tokenKey := tokenKey(string(s.AppID), tokenHash)

	var tokenBytes []byte
	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		var err error
		tokenBytes, err = conn.Get(ctx, tokenKey).Bytes()
		if errors.Is(err, goredis.Nil) {
			// Token Invalid
			return ErrAccountManagementTokenInvalid
		} else if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var token Token
	err = json.Unmarshal(tokenBytes, &token)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (s *RedisStore) ConsumeToken(ctx context.Context, tokenStr string) (*Token, error) {
	tokenHash := HashToken(tokenStr)

	tokenKey := tokenKey(string(s.AppID), tokenHash)

	var tokenBytes []byte
	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		var err error
		tokenBytes, err = conn.Get(ctx, tokenKey).Bytes()
		if errors.Is(err, goredis.Nil) {
			// Token Invalid
			return ErrAccountManagementTokenInvalid
		} else if err != nil {
			return err
		}

		_, err = conn.Del(ctx, tokenKey).Result()
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	var token Token
	err = json.Unmarshal(tokenBytes, &token)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (s *RedisStore) ConsumeToken_OAuth(ctx context.Context, tokenStr string) (*Token, error) {
	token, err := s.ConsumeToken(ctx, tokenStr)
	if errors.Is(err, ErrAccountManagementTokenInvalid) {
		return token, ErrOAuthTokenInvalid
	}
	return token, err
}

func tokenKey(appID string, tokenHash string) string {
	return fmt.Sprintf("app:%s:account-management-token:%s", appID, tokenHash)
}
