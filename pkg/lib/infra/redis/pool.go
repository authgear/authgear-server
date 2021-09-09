package redis

import (
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	redsyncgoredis "github.com/go-redsync/redsync/v4/redis/goredis/v8"
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

func (p *Pool) instance(connectionOptions *ConnectionOptions) *redisInstance {
	p.closeMutex.RLock()
	defer func() { p.closeMutex.RUnlock() }()
	if p.closed {
		panic("redis: pool is closed")
	}
	connKey := connectionOptions.ConnKey()

	p.cachedInstanceMutex.RLock()
	instance, exists := p.cachedInstance[connKey]
	p.cachedInstanceMutex.RUnlock()
	if exists {
		return instance
	}

	p.cachedInstanceMutex.Lock()
	instance, exists = p.cachedInstance[connKey]
	if !exists {
		instance = p.openInstance(connectionOptions)
		p.cachedInstance[connKey] = instance
	}
	p.cachedInstanceMutex.Unlock()

	return instance
}

func (p *Pool) Client(connectionOptions *ConnectionOptions) *redis.Client {
	return p.instance(connectionOptions).Client
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

func (p *Pool) openInstance(connectionOptions *ConnectionOptions) *redisInstance {
	opts, err := redis.ParseURL(connectionOptions.RedisURL)
	if err != nil {
		panic(err)
	}
	// FIXME(redis): MaxIdleConnection is not supported.
	opts.PoolSize = *connectionOptions.MaxOpenConnection
	opts.IdleTimeout = connectionOptions.IdleConnectionTimeout.Duration()
	opts.MaxConnAge = connectionOptions.MaxConnectionLifetime.Duration()
	client := redis.NewClient(opts)
	redsyncPool := redsyncgoredis.NewPool(client)
	redsyncInstance := redsync.New(redsyncPool)
	return &redisInstance{
		Client:  client,
		Redsync: redsyncInstance,
	}
}
