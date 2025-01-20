package db

import (
	"database/sql"
	"errors"
	"sync"

	"github.com/authgear/authgear-server/pkg/util/otelutil"
)

var actualPoolOpener = openPostgresDB

type Pool struct {
	closed     bool
	closeMutex sync.RWMutex

	cache      map[ConnectionInfo]*sql.DB
	cacheMutex sync.RWMutex
}

func NewPool() *Pool {
	return &Pool{cache: map[ConnectionInfo]*sql.DB{}}
}

func (p *Pool) Open(info ConnectionInfo, opts ConnectionOptions) (db *sql.DB, err error) {
	p.closeMutex.RLock()
	defer func() { p.closeMutex.RUnlock() }()
	if p.closed {
		return nil, errors.New("db: pool is closed")
	}

	p.cacheMutex.RLock()
	db, exists := p.cache[info]
	p.cacheMutex.RUnlock()

	if !exists {
		p.cacheMutex.Lock()
		db, exists = p.cache[info]
		if !exists {
			db, err = actualPoolOpener(info, opts)
			if err == nil {
				p.cache[info] = db
			}
		}
		p.cacheMutex.Unlock()
	}

	return
}

func (p *Pool) Close() (err error) {
	p.closeMutex.Lock()
	defer func() { p.closeMutex.Unlock() }()

	p.closed = true
	for _, db := range p.cache {
		if closeErr := db.Close(); closeErr != nil {
			err = closeErr
		}
	}
	if err == nil {
		clear(p.cache)
	}

	return
}

func openPostgresDB(info ConnectionInfo, opts ConnectionOptions) (*sql.DB, error) {
	pgdb, err := otelutil.OTelSQLOpenPostgres(info.DatabaseURL)
	if err != nil {
		return nil, err
	}

	pgdb.SetMaxOpenConns(opts.MaxOpenConnection)
	pgdb.SetMaxIdleConns(opts.MaxIdleConnection)
	pgdb.SetConnMaxLifetime(opts.MaxConnectionLifetime)
	pgdb.SetConnMaxIdleTime(opts.IdleConnectionTimeout)

	return pgdb, nil
}
