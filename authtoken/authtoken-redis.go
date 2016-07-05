package authtoken

import (
	"time"

	"github.com/garyburd/redigo/redis"
)

// RedisStore implements TokenStore by saving users' token
// in a redis server
type RedisStore struct {
	pool   *redis.Pool
	prefix string
}

// NewRedisStore creates a redis token store.
//
// address is url to the redis server
//
// prefix is a string prepending to access token key in redis
//   For example if the token is `cf4bdc65-3fe6-4d40-b7fd-58f00b82c506`
//   and the prefix is `myApp`, the key in redis should be
//   `myApp:cf4bdc65-3fe6-4d40-b7fd-58f00b82c506`.
func NewRedisStore(address string, prefix string) *RedisStore {
	store := RedisStore{}

	if prefix != "" {
		store.prefix = prefix + ":"
	}

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
	var expireAt int64
	if !t.ExpiredAt.IsZero() {
		expireAt = t.ExpiredAt.UnixNano()
	}
	return &RedisToken{
		t.AccessToken,
		expireAt,
		t.AppName,
		t.UserInfoID,
	}
}

// ToToken converts a RedisToken to auth token
func (r RedisToken) ToToken() *Token {
	expireAt := time.Time{}
	if r.ExpiredAt != 0 {
		expireAt = time.Unix(0, r.ExpiredAt).UTC()
	}
	return &Token{
		r.AccessToken,
		expireAt,
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

	accessTokenWithPrefix := r.prefix + accessToken

	v, err := redis.Values(c.Do("HGETALL", accessTokenWithPrefix))
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
	accessTokenWithPrefix := r.prefix + redisToken.AccessToken
	tokenArgs := redis.Args{}.Add(accessTokenWithPrefix).AddFlat(redisToken)

	c.Send("MULTI")
	c.Send("HMSET", tokenArgs...)
	if !token.ExpiredAt.IsZero() {
		c.Send("EXPIREAT", token.AccessToken, token.ExpiredAt.Unix())
	}
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

	accessTokenWithPrefix := r.prefix + accessToken
	_, err := c.Do("DEL", accessTokenWithPrefix)
	if err != nil {
		return err
	}

	return nil
}
