package redis

import (
	"sync"

	"github.com/go-redis/redis/v8"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type Pool struct {
	closed     bool
	closeMutex sync.RWMutex

	cachedClient      map[string]*redis.Client
	cachedClientMutex sync.RWMutex
}

func NewPool() *Pool {
	p := &Pool{
		cachedClient: make(map[string]*redis.Client),
	}
	return p
}

func (p *Pool) Client(cfg *config.RedisConfig, credentials *config.RedisCredentials) *redis.Client {
	p.closeMutex.RLock()
	defer func() { p.closeMutex.RUnlock() }()
	if p.closed {
		panic("redis: pool is closed")
	}
	connKey := credentials.ConnKey()

	p.cachedClientMutex.RLock()
	pool, exists := p.cachedClient[connKey]
	p.cachedClientMutex.RUnlock()
	if exists {
		return pool
	}

	p.cachedClientMutex.Lock()
	pool, exists = p.cachedClient[connKey]
	if !exists {
		pool = p.openRedis(cfg, credentials)
		p.cachedClient[connKey] = pool
	}
	p.cachedClientMutex.Unlock()

	return pool
}

func (p *Pool) Close() (err error) {
	p.closeMutex.Lock()
	defer func() { p.closeMutex.Unlock() }()

	p.closed = true
	for _, db := range p.cachedClient {
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
