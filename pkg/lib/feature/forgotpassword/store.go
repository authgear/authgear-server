package forgotpassword

import (
	"encoding/json"
	"fmt"
	"time"

	redigo "github.com/gomodule/redigo/redis"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

type Store struct {
	AppID config.AppID
	Redis *redis.Handle
}

func (s *Store) Create(code *Code) (err error) {
	bytes, err := json.Marshal(code)
	if err != nil {
		return
	}

	key := codeKey(s.AppID, code.CodeHash)

	err = s.Redis.WithConn(func(conn redis.Conn) error {
		_, err := redigo.String(conn.Do("SET", key, bytes, "PX", codeExpire(code), "NX"))
		if errorutil.Is(err, redigo.ErrNil) {
			err = fmt.Errorf("duplicated forgot password code: %w", err)
		}
		return err
	})

	return
}

func (s *Store) Update(code *Code) (err error) {
	bytes, err := json.Marshal(code)
	if err != nil {
		return
	}

	key := codeKey(s.AppID, code.CodeHash)

	err = s.Redis.WithConn(func(conn redis.Conn) error {
		_, err := redigo.String(conn.Do("SET", key, bytes, "PX", codeExpire(code), "XX"))
		if errorutil.Is(err, redigo.ErrNil) {
			err = fmt.Errorf("non-existent forgot password code: %w", err)
		}
		return err
	})

	return
}

func (s *Store) Get(codeHash string) (code *Code, err error) {
	key := codeKey(s.AppID, codeHash)

	err = s.Redis.WithConn(func(conn redis.Conn) error {
		data, err := redigo.Bytes(conn.Do("GET", key))
		if errorutil.Is(err, redigo.ErrNil) {
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

func codeExpire(code *Code) int64 {
	d := code.ExpireAt.Sub(code.CreatedAt)
	return int64(d / time.Millisecond)
}
