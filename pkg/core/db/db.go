package db

import "github.com/skygeario/skygear-server/pkg/core/config"

// DBProvider is the interface to get db with configuration
type DBProvider interface {
	GetDB(config.TenantConfiguration) IDB
}

type RealDBProvider struct{}

func (p RealDBProvider) GetDB(tConfig config.TenantConfiguration) IDB {
	return &DB{tConfig.DBConnectionStr}
}

type GetDB func() IDB

type IDB interface {
	GetRecord(string) string
}

type DB struct {
	ConnectionStr string
}

func (db DB) GetRecord(recordID string) string {
	return db.ConnectionStr + ":" + recordID
}
