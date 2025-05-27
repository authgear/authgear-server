package webappoauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	goredis "github.com/redis/go-redis/v9"

	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/globalredis"
	"github.com/authgear/authgear-server/pkg/util/base32"
	"github.com/authgear/authgear-server/pkg/util/crypto"
	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/rand"
)

type Store struct {
	Redis *globalredis.Handle
}

func NewStateToken() (stateToken string, stateTokenHash string) {
	// Some provider has a hard-limit on the length of the state.
	// Here we use 32 which is observed to be short enough.
	stateToken = rand.StringWithAlphabet(32, base32.Alphabet, rand.SecureRand)
	stateTokenHash = crypto.SHA256String(stateToken)
	return
}

func (s *Store) GenerateState(ctx context.Context, state *WebappOAuthState) (stateToken string, err error) {
	data, err := json.Marshal(state)
	if err != nil {
		return
	}

	ttl := duration.UserInteraction

	stateToken, stateTokenHash := NewStateToken()
	key := stateKey(stateTokenHash)

	err = s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		_, err := conn.SetNX(ctx, key, data, ttl).Result()
		if errors.Is(err, goredis.Nil) {
			err = fmt.Errorf("state string already exist: %w", err)
			return err
		} else if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return
	}

	return
}
func (s *Store) readState(ctx context.Context, stateToken string, deleteKey bool) (state *WebappOAuthState, err error) {
	stateTokenHash := crypto.SHA256String(stateToken)
	key := stateKey(stateTokenHash)

	var data []byte
	err = s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		var err error
		data, err = conn.Get(ctx, key).Bytes()
		if errors.Is(err, goredis.Nil) {
			err = ErrOAuthStateInvalid
			return err
		} else if err != nil {
			return err
		}

		if deleteKey {
			_, err = conn.Del(ctx, key).Result()
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return
	}

	var stateStruct WebappOAuthState
	err = json.Unmarshal(data, &stateStruct)
	if err != nil {
		return
	}

	state = &stateStruct
	return
}

func (s *Store) PopAndRecoverState(ctx context.Context, stateToken string) (state *WebappOAuthState, err error) {
	return s.readState(ctx, stateToken, true)
}

func (s *Store) RecoverState(ctx context.Context, stateToken string) (state *WebappOAuthState, err error) {
	return s.readState(ctx, stateToken, false)
}

func stateKey(stateTokenHash string) string {
	return fmt.Sprintf("oauthrelyingparty-state:%s", stateTokenHash)
}
