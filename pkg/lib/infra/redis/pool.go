package redis

import (
	"sync"

	"github.com/go-redis/redis/v8"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type Pool struct {
	closed     bool
	closeMutex sync.RWMutex

	cache      map[string]*redis.Client
	cacheMutex sync.RWMutex
}

func NewPool() *Pool {
	p := &Pool{cache: map[string]*redis.Client{}}
	return p
}

func (p *Pool) Open(cfg *config.RedisConfig, credentials *config.RedisCredentials) *redis.Client {
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

func (p *Pool) openRedis(cfg *config.RedisConfig, credentials *config.RedisCredentials) *redis.Client {
	opts, err := redis.ParseURL(credentials.RedisURL)
	if err != nil {
		panic(err)
	}
	// FIXME(redis): MaxIdleConnection is not supported.
	opts.PoolSize = *cfg.MaxOpenConnection
	opts.IdleTimeout = cfg.IdleConnectionTimeout.Duration()
	opts.MaxConnAge = cfg.MaxConnectionLifetime.Duration()
	return redis.NewClient(opts)
}
