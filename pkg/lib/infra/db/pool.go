package db

import (
	"errors"
	"sync"

	"github.com/jmoiron/sqlx"
)

type Pool struct {
	closed     bool
	closeMutex sync.RWMutex

	cache      map[string]*sqlx.DB
	cacheMutex sync.RWMutex
}

func NewPool() *Pool {
	return &Pool{cache: map[string]*sqlx.DB{}}
}

func (p *Pool) Open(opts ConnectionOptions) (db *sqlx.DB, err error) {
	source := opts.DatabaseURL

	p.closeMutex.RLock()
	defer func() { p.closeMutex.RUnlock() }()
	if p.closed {
		return nil, errors.New("db: pool is closed")
	}

	p.cacheMutex.RLock()
	db, exists := p.cache[source]
	p.cacheMutex.RUnlock()

	if !exists {
		p.cacheMutex.Lock()
		db, exists = p.cache[source]
		if !exists {
			db, err = p.openPostgresDB(opts)
			if err == nil {
				p.cache[source] = db
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

	return
}

func (p *Pool) openPostgresDB(opts ConnectionOptions) (db *sqlx.DB, err error) {
	db, err = sqlx.Open("postgres", opts.DatabaseURL)
	if err != nil {
		return
	}

	db.SetMaxOpenConns(opts.MaxOpenConnection)
	db.SetMaxIdleConns(opts.MaxIdleConnection)
	db.SetConnMaxLifetime(opts.MaxConnectionLifetime)
	db.SetConnMaxIdleTime(opts.IdleConnectionTimeout)
	return
}
