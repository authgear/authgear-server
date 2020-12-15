package forgotpassword

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	goredis "github.com/go-redis/redis/v8"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
)

type Store struct {
	AppID config.AppID
	Redis *redis.Handle
}

func (s *Store) Create(code *Code) (err error) {
	ctx := context.Background()
	bytes, err := json.Marshal(code)
	if err != nil {
		return
	}

	key := codeKey(s.AppID, code.CodeHash)
	err = s.Redis.WithConn(func(conn *goredis.Conn) error {
		_, err := conn.SetNX(ctx, key, bytes, codeExpire(code)).Result()
		if errors.Is(err, goredis.Nil) {
			err = fmt.Errorf("duplicated forgot password code: %w", err)
		}
		return err
	})

	return
}

func (s *Store) Update(code *Code) (err error) {
	ctx := context.Background()
	bytes, err := json.Marshal(code)
	if err != nil {
		return
	}

	key := codeKey(s.AppID, code.CodeHash)
	err = s.Redis.WithConn(func(conn *goredis.Conn) error {
		_, err := conn.SetXX(ctx, key, bytes, codeExpire(code)).Result()
		if errors.Is(err, goredis.Nil) {
			err = fmt.Errorf("non-existent forgot password code: %w", err)
		}
		return err
	})

	return
}

func (s *Store) Get(codeHash string) (code *Code, err error) {
	ctx := context.Background()
	key := codeKey(s.AppID, codeHash)

	err = s.Redis.WithConn(func(conn *goredis.Conn) error {
		data, err := conn.Get(ctx, key).Bytes()
		if errors.Is(err, goredis.Nil) {
			err = ErrInvalidCode
			return err
		} else if err != nil {
			return err
		}

		return json.Unmarshal(data, &code)
	})

	return
}

func codeKey(appID config.AppID, codeHash string) string {
	return fmt.Sprintf("app:%s:forgotpassword-code:%s", appID, codeHash)
}

func codeExpire(code *Code) time.Duration {
	d := code.ExpireAt.Sub(code.CreatedAt)
	return d
}
