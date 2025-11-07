package redisqueue

import (
	"time"
)

type QueueName string

const (
	QueueUserImport  QueueName = "user-import"
	QueueUserExport  QueueName = "user-export"
	QueueUserReindex QueueName = "user-reindex"
)

func (q QueueName) GetTTLForEnqueue() time.Duration {
	switch q {
	case QueueUserReindex:
		return 20 * time.Minute
	default:
		return 24 * time.Hour
	}
}

func (q QueueName) GetTTLForRetention() time.Duration {
	switch q {
	case QueueUserReindex:
		return 5 * time.Minute
	default:
		return 24 * time.Hour
	}
}
