package db

import (
	"sync"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

type Pool struct {
	Config *config.DatabaseEnvironmentConfig
	db     *sqlx.DB
	mutex  sync.RWMutex
}

func NewPool(cfg *config.DatabaseEnvironmentConfig) *Pool {
	return &Pool{Config: cfg}
}

func (p *Pool) Open() (*sqlx.DB, error) {
	db := p.loadDB()
	if db != nil {
		return db, nil
	}
	return p.doOpen()
}

func (p *Pool) loadDB() *sqlx.DB {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return p.db
}

func (p *Pool) doOpen() (*sqlx.DB, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.db != nil {
		return p.db, nil
	}

	db, err := sqlx.Open("postgres", p.Config.DatabaseURL)
	if err != nil {
		return nil, errorutil.HandledWithMessage(err, "failed to connect to database")
	}

	db.SetMaxOpenConns(p.Config.MaxOpenConn)
	db.SetMaxIdleConns(p.Config.MaxIdleConn)
	db.SetConnMaxLifetime(time.Second * time.Duration(p.Config.ConnMaxLifetimeSeconds))
	db.SetConnMaxIdleTime(time.Second * time.Duration(p.Config.ConnMaxIdleTimeSeconds))

	p.db = db
	return db, nil
}
