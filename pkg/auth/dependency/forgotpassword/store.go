package forgotpassword

import (
	"encoding/json"
	"fmt"
	"time"

	redigo "github.com/gomodule/redigo/redis"

	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/redis"
)

type Store struct {
	Redis *redis.Context
}

func (s *Store) Create(code *Code) (err error) {
	bytes, err := json.Marshal(code)
	if err != nil {
		return
	}

	conn := s.Redis.Conn()
	key := codeKey(code.CodeHash)
	_, err = redigo.String(conn.Do("SET", key, bytes, "PX", codeExpire(code), "NX"))
	if errors.Is(err, redigo.ErrNil) {
		err = errors.Newf("duplicated forgot password code: %w", err)
		return
	}

	return
}

func (s *Store) Update(code *Code) (err error) {
	bytes, err := json.Marshal(code)
	if err != nil {
		return
	}

	conn := s.Redis.Conn()
	key := codeKey(code.CodeHash)
	_, err = redigo.String(conn.Do("SET", key, bytes, "PX", codeExpire(code), "XX"))
	if errors.Is(err, redigo.ErrNil) {
		err = errors.Newf("non-existent forgot password code: %w", err)
		return
	}

	return
}

func (s *Store) Get(codeHash string) (code *Code, err error) {
	conn := s.Redis.Conn()
	key := codeKey(codeHash)

	data, err := redigo.Bytes(conn.Do("GET", key))
	if errors.Is(err, redigo.ErrNil) {
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
