package webappoauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	goredis "github.com/go-redis/redis/v8"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/base32"
	"github.com/authgear/authgear-server/pkg/util/crypto"
	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/rand"
)

type Store struct {
	Context context.Context
	Redis   *appredis.Handle
	AppID   config.AppID
}

func NewStateToken() (stateToken string, stateTokenHash string) {
	// Some provider has a hard-limit on the length of the state.
	// Here we use 32 which is observed to be short enough.
	stateToken = rand.StringWithAlphabet(32, base32.Alphabet, rand.SecureRand)
	stateTokenHash = crypto.SHA256String(stateToken)
	return
}

func (s *Store) GenerateState(state *WebappOAuthState) (stateToken string, err error) {
	data, err := json.Marshal(state)
	if err != nil {
		return
	}

	ttl := duration.UserInteraction

	stateToken, stateTokenHash := NewStateToken()
	key := stateKey(string(s.AppID), stateTokenHash)

	err = s.Redis.WithConn(func(conn *goredis.Conn) error {
		_, err := conn.SetNX(s.Context, key, data, ttl).Result()
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

func (s *Store) PopAndRecoverState(stateToken string) (state *WebappOAuthState, err error) {
	stateTokenHash := crypto.SHA256String(stateToken)
	key := stateKey(string(s.AppID), stateTokenHash)

	var data []byte
	err = s.Redis.WithConn(func(conn *goredis.Conn) error {
		var err error
		data, err = conn.GetDel(s.Context, key).Bytes()
		if errors.Is(err, goredis.Nil) {
			err = ErrOAuthStateInvalid
			return err
		} else if err != nil {
			return err
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

func stateKey(appID string, stateTokenHash string) string {
	return fmt.Sprintf("app:%s:oauthrelyingparty-state:%s", appID, stateTokenHash)
}
