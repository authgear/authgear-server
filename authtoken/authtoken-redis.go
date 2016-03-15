package authtoken

import (
	"time"

	"github.com/garyburd/redigo/redis"
)

// RedisStore implements TokenStore by saving users' token
// in a redis server
type RedisStore struct {
	pool *redis.Pool
}

// NewRedisStore creates a redis token store.
func NewRedisStore(address string) *RedisStore {
	store := RedisStore{}

	store.pool = &redis.Pool{
		MaxIdle: 50, // NOTE: May make it configurable
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialURL(address)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	return &store
}

// RedisToken stores a Token with UnixNano timestamp
type RedisToken struct {
	AccessToken string `redis:"accessToken"`
	ExpiredAt   int64  `redis:"expiredAt"`
	AppName     string `redis:"appName"`
	UserInfoID  string `redis:"userInfoID"`
}

// ToRedisToken converts an auth token to RedisToken
func (t Token) ToRedisToken() *RedisToken {
	return &RedisToken{
		t.AccessToken,
		t.ExpiredAt.UnixNano(),
		t.AppName,
		t.UserInfoID,
	}
}

// ToToken converts a RedisToken to auth token
func (r RedisToken) ToToken() *Token {
	return &Token{
		r.AccessToken,
		time.Unix(0, r.ExpiredAt).UTC(),
		r.AppName,
		r.UserInfoID,
	}
}

// Get tries to read the specified access token from redis store and
// writes to the supplied Token.
func (r *RedisStore) Get(accessToken string, token *Token) error {
	c := r.pool.Get()
	if err := c.Err(); err != nil {
		return err
	}
	defer c.Close()

	v, err := redis.Values(c.Do("HGETALL", accessToken))
	if err != nil {
		return err
	}
	// Check if the result is empty
	if len(v) == 0 {
		return &NotFoundError{accessToken, err}
	}

	var redisToken RedisToken
	err = redis.ScanStruct(v, &redisToken)
	if err != nil {
		return err
	}
	*token = *redisToken.ToToken()

	return nil
}

// Put writes the specified token into redis store and overwrites existing
// Token if any.
func (r *RedisStore) Put(token *Token) error {
	c := r.pool.Get()
	if err := c.Err(); err != nil {
		return err
	}
	defer c.Close()

	redisToken := token.ToRedisToken()
	tokenArgs := redis.Args{}.Add(redisToken.AccessToken).AddFlat(redisToken)

	c.Send("MULTI")
	c.Send("HMSET", tokenArgs...)
	c.Send("EXPIREAT", token.AccessToken, token.ExpiredAt.Unix())
	_, err := c.Do("EXEC")
	if err != nil {
		return err
	}

	return nil
}

// Delete removes the access token from redis store.
func (r *RedisStore) Delete(accessToken string) error {
	c := r.pool.Get()
	if err := c.Err(); err != nil {
		return err
	}
	defer c.Close()

	_, err := c.Do("DEL", accessToken)
	if err != nil {
		return err
	}

	return nil
}
