package db

import (
	"time"
)

type ConnectionOptions struct {
	DatabaseURL           string
	MaxOpenConnection     int
	MaxIdleConnection     int
	MaxConnectionLifetime time.Duration
	IdleConnectionTimeout time.Duration

	UsePreparedStatements bool
}
