package redis

import (
	goredis "github.com/go-redis/redis/v8"
)

func NewClient(redisURL string) (*goredis.Client, error) {
	opts, err := goredis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	redisClient := goredis.NewClient(opts)
	return redisClient, nil
}
