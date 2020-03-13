package redis

import (
	"errors"
	"fmt"
	"time"

	"github.com/FZambia/sentinel"
	"github.com/gomodule/redigo/redis"
)

// When sentinel is enabled, host and port will be ignored
type Configuration struct {
	Host     string         `envconfig:"HOST"`
	Port     int            `envconfig:"PORT" default:"6379"`
	Password string         `envconfig:"PASSWORD"`
	DB       int            `envconfig:"DB" default:"0"`
	Sentinel SentinelConfig `envconfig:"SENTINEL"`
}

type SentinelConfig struct {
	Enabled    bool     `envconfig:"ENABLED"`
	Addrs      []string `envconfig:"ADDRS"`
	MasterName string   `envconfig:"MASTER_NAME"`
}

func NewPool(config Configuration) (*redis.Pool, error) {
	if config.Sentinel.Enabled {
		if len(config.Sentinel.Addrs) == 0 {
			return nil, errors.New("redis sentinel addrs are not provided")
		}
	} else {
		if config.Host == "" {
			return nil, errors.New("redis host is not provided")
		}
	}
	if config.Sentinel.Enabled {
		return newSentinelPool(config), nil
	}
	hostPort := fmt.Sprintf("%s:%d", config.Host, config.Port)
	return newPool(
		func() (conn redis.Conn, err error) {
			conn, err = redis.Dial(
				"tcp",
				hostPort,
				redis.DialDatabase(config.DB),
				redis.DialPassword(config.Password),
			)
			return
		},
		func(conn redis.Conn, t time.Time) (err error) {
			_, err = conn.Do("PING")
			return
		},
	), nil
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

func newSentinelPool(config Configuration) *redis.Pool {
	sntnl := &sentinel.Sentinel{
		Addrs:      config.Sentinel.Addrs,
		MasterName: config.Sentinel.MasterName,
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
			masterAddr, err := sntnl.MasterAddr()
			if err != nil {
				return nil, err
			}
			c, err := redis.Dial(
				"tcp",
				masterAddr,
				redis.DialDatabase(config.DB),
				redis.DialPassword(config.Password),
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
