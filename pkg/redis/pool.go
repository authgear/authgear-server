package redis

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/FZambia/sentinel"
	"github.com/gomodule/redigo/redis"

	"github.com/skygeario/skygear-server/pkg/auth/config"
)

type Pool struct {
	closed     bool
	closeMutex sync.RWMutex

	cache      map[string]*redis.Pool
	cacheMutex sync.RWMutex
}

func NewPool() *Pool {
	p := &Pool{cache: map[string]*redis.Pool{}}
	return p
}

func (p *Pool) Open(c *config.RedisCredentials) *redis.Pool {
	p.closeMutex.RLock()
	defer func() { p.closeMutex.RUnlock() }()
	if p.closed {
		panic("redis: pool is closed")
	}
	connKey := c.ConnKey()

	p.cacheMutex.RLock()
	pool, exists := p.cache[connKey]
	p.cacheMutex.RUnlock()

	if !exists {
		p.cacheMutex.Lock()
		pool, exists = p.cache[connKey]
		if !exists {
			p.cache[connKey] = openRedis(c)
		}
		p.cacheMutex.Unlock()
	}

	return pool
}

func (p *Pool) Close() (err error) {
	p.closeMutex.Lock()
	defer func() { p.closeMutex.Unlock() }()

	p.closed = true
	for _, db := range p.cache {
		if closeErr := db.Close(); closeErr != nil {
			err = closeErr
		}
	}

	return
}

func openRedis(c *config.RedisCredentials) *redis.Pool {
	if c.Sentinel.Enabled {
		return newSentinelPool(c)
	}
	hostPort := fmt.Sprintf("%s:%d", c.Host, c.Port)
	return newPool(
		func() (conn redis.Conn, err error) {
			conn, err = redis.Dial(
				"tcp",
				hostPort,
				redis.DialDatabase(c.DB),
				redis.DialPassword(c.Password),
			)
			return
		},
		func(conn redis.Conn, t time.Time) (err error) {
			_, err = conn.Do("PING")
			return
		},
	)
}

func newSentinelPool(c *config.RedisCredentials) *redis.Pool {
	s := &sentinel.Sentinel{
		Addrs:      c.Sentinel.Addrs,
		MasterName: c.Sentinel.MasterName,
		Dial: func(addr string) (redis.Conn, error) {
			dialConnectTimeout := 3 * time.Second
			timeout := 500 * time.Millisecond
			c, err := redis.Dial(
				"tcp",
				addr,
				redis.DialConnectTimeout(dialConnectTimeout),
				redis.DialReadTimeout(timeout),
				redis.DialWriteTimeout(timeout),
			)
			if err != nil {
				return nil, err
			}
			return c, nil
		},
	}
	return newPool(
		func() (redis.Conn, error) {
			masterAddr, err := s.MasterAddr()
			if err != nil {
				return nil, err
			}
			c, err := redis.Dial(
				"tcp",
				masterAddr,
				redis.DialDatabase(c.DB),
				redis.DialPassword(c.Password),
			)
			if err != nil {
				return nil, err
			}
			return c, nil
		},
		func(c redis.Conn, t time.Time) error {
			if !sentinel.TestRole(c, "master") {
				return errors.New("Role check failed")
			}
			return nil
		},
	)
}

func newPool(
	dialFunc func() (conn redis.Conn, err error),
	testOnBorrowFunc func(conn redis.Conn, t time.Time) error,
) *redis.Pool {
	// TODO(pool): configurable / profile for good value?
	return &redis.Pool{
		MaxIdle:      30,
		IdleTimeout:  5 * time.Minute,
		Dial:         dialFunc,
		TestOnBorrow: testOnBorrowFunc,
	}
}
