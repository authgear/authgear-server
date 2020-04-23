package forgotpassword

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	goredis "github.com/gomodule/redigo/redis"

	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/core/redis"
)

type StoreImpl struct {
	Context context.Context
}

var _ Store = &StoreImpl{}

func (s *StoreImpl) Create(code *Code) (err error) {
	bytes, err := json.Marshal(code)
	if err != nil {
		return
	}

	conn := redis.GetConn(s.Context)
	key := codeKey(code.CodeHash)
	_, err = goredis.String(conn.Do("SET", key, bytes, "PX", codeExpire(code), "NX"))
	if errors.Is(err, goredis.ErrNil) {
		err = errors.Newf("duplicated forgot password code: %w", err)
		return
	}

	return
}

func (s *StoreImpl) Update(code *Code) (err error) {
	bytes, err := json.Marshal(code)
	if err != nil {
		return
	}

	conn := redis.GetConn(s.Context)
	key := codeKey(code.CodeHash)
	_, err = goredis.String(conn.Do("SET", key, bytes, "PX", codeExpire(code), "XX"))
	if errors.Is(err, goredis.ErrNil) {
		err = errors.Newf("non-existent forgot password code: %w", err)
		return
	}

	return
}

func (s *StoreImpl) Get(codeHash string) (code *Code, err error) {
	conn := redis.GetConn(s.Context)
	key := codeKey(codeHash)

	data, err := goredis.Bytes(conn.Do("GET", key))
	if errors.Is(err, goredis.ErrNil) {
		err = ErrInvalidCode
		return
	} else if err != nil {
		return
	}

	err = json.Unmarshal(data, &code)
	return
}

func codeKey(codeHash string) string {
	return fmt.Sprintf("forgotpassword-code:%s", codeHash)
}

func codeExpire(code *Code) int64 {
	d := code.ExpireAt.Sub(code.CreatedAt)
	return int64(d / time.Millisecond)
}
