package sso

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	goredis "github.com/gomodule/redigo/redis"

	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/core/redis"
)

var ErrCodeNotFound = errors.New("code not found")

type SkygearAuthorizationCodeStore interface {
	Get(codeHash string) (*SkygearAuthorizationCode, error)
	Set(code *SkygearAuthorizationCode) error
	Delete(codeHash string) error
}

func NewSkygearAuthorizationCodeStore(ctx context.Context) SkygearAuthorizationCodeStore {
	return &storeImpl{
		Context: ctx,
	}
}

type storeImpl struct {
	Context context.Context
}

var _ SkygearAuthorizationCodeStore = &storeImpl{}

func (s *storeImpl) Get(codeHash string) (code *SkygearAuthorizationCode, err error) {
	conn := redis.GetConn(s.Context)
	key := codeKey(codeHash)
	data, err := goredis.Bytes(conn.Do("GET", key))
	if errors.Is(err, goredis.ErrNil) {
		err = ErrCodeNotFound
		return
	} else if err != nil {
		return
	}
	err = json.Unmarshal(data, &code)
	return
}

func (s *storeImpl) Set(code *SkygearAuthorizationCode) (err error) {
	bytes, err := json.Marshal(code)
	if err != nil {
		return
	}

	conn := redis.GetConn(s.Context)
	key := codeKey(code.CodeHash)
	_, err = goredis.String(conn.Do("SET", key, bytes, "PX", toMilliseconds(5*time.Minute), "NX"))
	if errors.Is(err, goredis.ErrNil) {
		err = errors.Newf("duplicated authorization code: %w", err)
		return
	}
	return
}

func (s *storeImpl) Delete(codeHash string) error {
	conn := redis.GetConn(s.Context)
	key := codeKey(codeHash)
	_, err := conn.Do("DEL", key)
	return err
}

func toMilliseconds(d time.Duration) int64 {
	return int64(d / time.Millisecond)
}

func codeKey(codeHash string) string {
	return fmt.Sprintf("authorization-code:%s", codeHash)
}
