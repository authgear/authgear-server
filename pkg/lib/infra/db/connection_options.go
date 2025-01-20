package db

import (
	"time"
)

type ConnectionPurpose string

const (
	ConnectionPurposeGlobal         = "global"
	ConnectionPurposeApp            = "app"
	ConnectionPurposeAuditReadOnly  = "audit_read_only"
	ConnectionPurposeAuditReadWrite = "audit_read_write"
	ConnectionPurposeSearch         = "search"
)

type ConnectionInfo struct {
	Purpose     ConnectionPurpose
	DatabaseURL string
}

type ConnectionOptions struct {
	MaxOpenConnection     int
	MaxIdleConnection     int
	MaxConnectionLifetime time.Duration
	IdleConnectionTimeout time.Duration
}
