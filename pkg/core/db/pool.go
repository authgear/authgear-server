package db

import (
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/skygeario/skygear-server/pkg/core/errors"
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

func (p *Pool) Open(source string) (db *sqlx.DB, err error) {
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
			db, err = openPostgresDB(source)
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

func openPostgresDB(url string) (db *sqlx.DB, err error) {
	db, err = sqlx.Open("postgres", url)
	if err != nil {
		return
	}

	// TODO(pool): configurable / profile for good value?
	db.SetMaxOpenConns(30)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	return
}
