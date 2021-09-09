package redis

import (
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	redsyncgoredis "github.com/go-redsync/redsync/v4/redis/goredis/v8"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type redisInstance struct {
	Client  *redis.Client
	Redsync *redsync.Redsync
}

type Pool struct {
	closed     bool
	closeMutex sync.RWMutex

	cachedInstance      map[string]*redisInstance
	cachedInstanceMutex sync.RWMutex
}

func NewPool() *Pool {
	p := &Pool{
		cachedInstance: make(map[string]*redisInstance),
	}
	return p
}

func (p *Pool) Instance(cfg *config.RedisConfig, credentials *config.RedisCredentials) *redisInstance {
	p.closeMutex.RLock()
	defer func() { p.closeMutex.RUnlock() }()
	if p.closed {
		panic("redis: pool is closed")
	}
	connKey := credentials.ConnKey()

	p.cachedInstanceMutex.RLock()
	instance, exists := p.cachedInstance[connKey]
	p.cachedInstanceMutex.RUnlock()
	if exists {
		return instance
	}

	p.cachedInstanceMutex.Lock()
	instance, exists = p.cachedInstance[connKey]
	if !exists {
		instance = p.openInstance(cfg, credentials)
		p.cachedInstance[connKey] = instance
	}
	p.cachedInstanceMutex.Unlock()

	return instance
}

func (p *Pool) Client(cfg *config.RedisConfig, credentials *config.RedisCredentials) *redis.Client {
	return p.Instance(cfg, credentials).Client
}

func (p *Pool) Close() (err error) {
	p.closeMutex.Lock()
	defer func() { p.closeMutex.Unlock() }()

	p.closed = true
	for _, instance := range p.cachedInstance {
		if closeErr := instance.Client.Close(); closeErr != nil {
			err = closeErr
		}
	}

	return
}

func (p *Pool) openInstance(cfg *config.RedisConfig, credentials *config.RedisCredentials) *redisInstance {
	opts, err := redis.ParseURL(credentials.RedisURL)
	if err != nil {
		panic(err)
	}
	// FIXME(redis): MaxIdleConnection is not supported.
	opts.PoolSize = *cfg.MaxOpenConnection
	opts.IdleTimeout = cfg.IdleConnectionTimeout.Duration()
	opts.MaxConnAge = cfg.MaxConnectionLifetime.Duration()
	client := redis.NewClient(opts)
	redsyncPool := redsyncgoredis.NewPool(client)
	redsyncInstance := redsync.New(redsyncPool)
	return &redisInstance{
		Client:  client,
		Redsync: redsyncInstance,
	}
}
