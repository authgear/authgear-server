package accountmanagement

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	goredis "github.com/go-redis/redis/v8"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

type RedisStore struct {
	Context context.Context
	AppID   config.AppID
	Redis   *appredis.Handle
	Clock   clock.Clock
}

type GenerateTokenOptions struct {
	// OAuth
	UserID      string
	Alias       string
	MaybeState  string
	RedirectURI string

	// Phone
	PhoneNumber string

	// Email
	Email string
}

func (s *RedisStore) GenerateToken(options GenerateTokenOptions) (string, error) {
	tokenString := GenerateToken()
	tokenHash := HashToken(tokenString)

	now := s.Clock.NowUTC()
	ttl := duration.UserInteraction
	expireAt := now.Add(ttl)

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

		// Phone
		PhoneNumber: options.PhoneNumber,

		// Email
		Email: options.Email,
	}

	tokenBytes, err := json.Marshal(token)
	if err != nil {
		return "", err
	}

	tokenKey := tokenKey(token.AppID, token.TokenHash)

	err = s.Redis.WithConnContext(s.Context, func(conn redis.Redis_6_0_Cmdable) error {
		_, err = conn.SetNX(s.Context, tokenKey, tokenBytes, ttl).Result()
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

func (s *RedisStore) GetToken(tokenStr string) (*Token, error) {
	tokenHash := HashToken(tokenStr)

	tokenKey := tokenKey(string(s.AppID), tokenHash)

	var tokenBytes []byte
	err := s.Redis.WithConnContext(s.Context, func(conn *goredis.Conn) error {
		var err error
		tokenBytes, err = conn.Get(s.Context, tokenKey).Bytes()
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

func (s *RedisStore) ConsumeToken(tokenStr string) (*Token, error) {
	tokenHash := HashToken(tokenStr)

	tokenKey := tokenKey(string(s.AppID), tokenHash)

	var tokenBytes []byte
	err := s.Redis.WithConnContext(s.Context, func(conn redis.Redis_6_0_Cmdable) error {
		var err error
		tokenBytes, err = conn.Get(s.Context, tokenKey).Bytes()
		if errors.Is(err, goredis.Nil) {
			// Token Invalid
			return ErrOAuthTokenInvalid
		} else if err != nil {
			return err
		}

		_, err = conn.Del(s.Context, tokenKey).Result()
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

func tokenKey(appID string, tokenHash string) string {
	return fmt.Sprintf("app:%s:account-management-token:%s", appID, tokenHash)
}
