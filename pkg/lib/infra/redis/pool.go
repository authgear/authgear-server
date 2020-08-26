package redis

import (
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/authgear/authgear-server/pkg/lib/config"
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
	dialFunc := func() (conn redis.Conn, err error) {
		conn, err = redis.DialURL(credentials.RedisURL)
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
