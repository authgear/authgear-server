package redis

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/FZambia/sentinel"
	"github.com/gomodule/redigo/redis"

	"github.com/authgear/authgear-server/pkg/auth/config"
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

func (p *Pool) Open(cfg *config.RedisConfig, credentials *config.RedisCredentials) *redis.Pool {
	p.closeMutex.RLock()
	defer func() { p.closeMutex.RUnlock() }()
	if p.closed {
		panic("redis: pool is closed")
	}
	connKey := credentials.ConnKey()

	p.cacheMutex.RLock()
	pool, exists := p.cache[connKey]
	p.cacheMutex.RUnlock()
	if exists {
		return pool
	}

	p.cacheMutex.Lock()
	pool, exists = p.cache[connKey]
	if !exists {
		pool = p.openRedis(cfg, credentials)
		p.cache[connKey] = pool
	}
	p.cacheMutex.Unlock()

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

func (p *Pool) openRedis(cfg *config.RedisConfig, credentials *config.RedisCredentials) *redis.Pool {
	if credentials.Sentinel.Enabled {
		return p.newSentinelPool(cfg, credentials)
	}

	hostPort := fmt.Sprintf("%s:%d", credentials.Host, credentials.Port)
	dialFunc := func() (conn redis.Conn, err error) {
		conn, err = redis.Dial(
			"tcp",
			hostPort,
			redis.DialDatabase(credentials.DB),
			redis.DialPassword(credentials.Password),
		)
		return
	}
	testOnBorrowFunc := func(conn redis.Conn, t time.Time) (err error) {
		_, err = conn.Do("PING")
		return
	}

	return p.newPool(
		cfg,
		dialFunc,
		testOnBorrowFunc,
	)
}

func (p *Pool) newSentinelPool(cfg *config.RedisConfig, c *config.RedisCredentials) *redis.Pool {
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

	return p.newPool(
		cfg,
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

func (p *Pool) newPool(
	cfg *config.RedisConfig,
	dialFunc func() (conn redis.Conn, err error),
	testOnBorrowFunc func(conn redis.Conn, t time.Time) error,
) *redis.Pool {
	return &redis.Pool{
		MaxActive:       *cfg.MaxOpenConnection,
		MaxIdle:         *cfg.MaxIdleConnection,
		IdleTimeout:     cfg.IdleConnectionTimeout.Duration(),
		MaxConnLifetime: cfg.MaxConnectionLifetime.Duration(),
		Wait:            true,
		Dial:            dialFunc,
		TestOnBorrow:    testOnBorrowFunc,
	}
}
