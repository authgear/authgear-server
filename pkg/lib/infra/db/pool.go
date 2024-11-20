package db

import (
	"context"
	"database/sql"
	"errors"
	"sync"

	"github.com/jmoiron/sqlx"
)

type PoolDB struct {
	db *sqlx.DB

	closeMutex sync.RWMutex
	stmtLock   sync.RWMutex
	stmts      map[string]*sqlx.Stmt
}

func (d *PoolDB) Close() error {
	d.closeMutex.Lock()
	defer d.closeMutex.Unlock()

	if d.db == nil {
		return nil
	}

	for _, stmt := range d.stmts {
		_ = stmt.Close()
	}
	clear(d.stmts)

	if err := d.db.Close(); err != nil {
		return err
	}

	d.db = nil
	return nil
}

func (d *PoolDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error) {
	d.closeMutex.RLock()
	defer d.closeMutex.RUnlock()

	if d.db == nil {
		return nil, errors.New("db: db is closed")
	}

	return d.db.BeginTxx(ctx, opts)
}

func (d *PoolDB) Prepare(ctx context.Context, query string) (stmt *sqlx.Stmt, err error) {
	d.closeMutex.RLock()
	defer d.closeMutex.RUnlock()

	if d.db == nil {
		return nil, errors.New("db: db is closed")
	}

	d.stmtLock.RLock()
	stmt, exists := d.stmts[query]
	d.stmtLock.RUnlock()

	if !exists {
		d.stmtLock.Lock()
		stmt, exists = d.stmts[query]
		if !exists {
			stmt, err = d.db.PreparexContext(ctx, query)
			if err == nil {
				d.stmts[query] = stmt
			}
		}
		d.stmtLock.Unlock()
	}

	return
}

type Pool struct {
	closed     bool
	closeMutex sync.RWMutex

	cache      map[string]*PoolDB
	cacheMutex sync.RWMutex
}

func NewPool() *Pool {
	return &Pool{cache: map[string]*PoolDB{}}
}

func (p *Pool) Open(opts ConnectionOptions) (db *PoolDB, err error) {
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
	if err == nil {
		clear(p.cache)
	}

	return
}

func (p *Pool) openPostgresDB(opts ConnectionOptions) (*PoolDB, error) {
	pgdb, err := sqlx.Open("postgres", opts.DatabaseURL)
	if err != nil {
		return nil, err
	}

	pgdb.SetMaxOpenConns(opts.MaxOpenConnection)
	pgdb.SetMaxIdleConns(opts.MaxIdleConnection)
	pgdb.SetConnMaxLifetime(opts.MaxConnectionLifetime)
	pgdb.SetConnMaxIdleTime(opts.IdleConnectionTimeout)

	return &PoolDB{
		db:    pgdb,
		stmts: make(map[string]*sqlx.Stmt),
	}, nil
}
