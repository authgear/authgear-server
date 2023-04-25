package ratelimit

import "time"

type Storage interface {
	WithConn(func(StorageConn) error) error
}

type StorageConn interface {
	TakeToken(spec BucketSpec, now time.Time, delta int) (int, error)
	GetResetTime(spec BucketSpec, now time.Time) (time.Time, error)
	Reset(spec BucketSpec, now time.Time) error
}
