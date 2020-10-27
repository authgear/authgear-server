package ratelimit

import "time"

type Storage interface {
	WithConn(func(StorageConn) error) error
}

type StorageConn interface {
	TakeToken(bucket Bucket, now time.Time) (int, error)
	GetResetTime(bucket Bucket, now time.Time) (time.Time, error)
	Reset(bucket Bucket, now time.Time) error
}
