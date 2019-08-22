package redis

import (
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

type Configuration struct {
	Host     string `envconfig:"HOST"`
	Port     int    `envconfig:"PORT" default:"6379"`
	Password string `envconfig:"PASSWORD"`
	DB       int    `envconfig:"DB" default:"0"`
}

func NewPool(config Configuration) *redis.Pool {
	hostPort := fmt.Sprintf("%s:%d", config.Host, config.Port)
	// TODO(pool): configurable / profile for good value?
	return &redis.Pool{
		MaxIdle:     5,
		IdleTimeout: 5 * time.Minute,
		Dial: func() (conn redis.Conn, err error) {
			conn, err = redis.Dial(
				"tcp",
				hostPort,
				redis.DialDatabase(config.DB),
				redis.DialPassword(config.Password),
			)
			return
		},
		TestOnBorrow: func(conn redis.Conn, t time.Time) (err error) {
			_, err = conn.Do("PING")
			return
		},
	}
}
