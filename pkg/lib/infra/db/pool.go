package db

import (
	"sync"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/authgear/authgear-server/pkg/util/errors"
)

type Pool struct {
	closed     bool
	closeMutex sync.RWMutex

	cache      map[string]*sqlx.DB
	cacheMutex sync.RWMutex
}

type OpenOptions struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func NewPool() *Pool {
	return &Pool{cache: map[string]*sqlx.DB{}}
}

func (p *Pool) Open(opts OpenOptions) (db *sqlx.DB, err error) {
	source := opts.URL

	p.closeMutex.RLock()
	defer func() { p.closeMutex.RUnlock() }()
	if p.closed {
		return nil, errors.New("skydb: pool is closed")
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

func (p *Pool) openPostgresDB(opts OpenOptions) (db *sqlx.DB, err error) {
	db, err = sqlx.Open("postgres", opts.URL)
	if err != nil {
		return
	}

	db.SetMaxOpenConns(opts.MaxOpenConns)
	db.SetMaxIdleConns(opts.MaxIdleConns)
	db.SetConnMaxLifetime(opts.ConnMaxLifetime)
	return
}
