package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

type Handle struct {
	Context context.Context
	Config  *config.DatabaseConfig
	Logger  Logger

	database *sqlx.DB `wire:"-"`
	tx       *sqlx.Tx `wire:"-"`
}

func (h *Handle) Conn() (sqlx.ExtContext, error) {
	tx := h.tx
	if tx == nil {
		return h.openDB()
	}
	return tx, nil
}

func (h *Handle) HasTx() bool {
	return h.tx != nil
}

// WithTx commits if do finishes without error and rolls back otherwise.
func (h *Handle) WithTx(do func() error) (err error) {
	if err = h.beginTx(); err != nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			_ = h.rollbackTx()
			panic(r)
		} else if err != nil {
			if rbErr := h.rollbackTx(); rbErr != nil {
				h.Logger.WithError(rbErr).Error("failed to rollback tx")
			}
		} else {
			err = h.commitTx()
		}
	}()

	err = do()
	return
}

// ReadOnly runs do in a transaction and rolls back always.
func (h *Handle) ReadOnly(do func() error) (err error) {
	if err = h.beginTx(); err != nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			_ = h.rollbackTx()
			panic(r)
		} else if err != nil {
			if rbErr := h.rollbackTx(); rbErr != nil {
				h.Logger.WithError(rbErr).Error("failed to rollback tx")
			}
		} else {
			err = h.rollbackTx()
		}
	}()

	err = do()
	return
}

func (h *Handle) beginTx() error {
	if h.tx != nil {
		panic("db: a transaction has already begun")
	}

	db, err := h.openDB()
	if err != nil {
		return err
	}
	tx, err := db.BeginTxx(h.Context, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
	})
	if err != nil {
		return errorutil.HandledWithMessage(err, "failed to begin transaction")
	}

	h.tx = tx

	return nil
}

func (h *Handle) commitTx() error {
	if h.tx == nil {
		panic("db: a transaction has not begun")
	}

	err := h.tx.Commit()
	if err != nil {
		return errorutil.HandledWithMessage(err, "failed to commit transaction")
	}
	h.tx = nil

	return nil
}

func (h *Handle) rollbackTx() error {
	if h.tx == nil {
		panic("db: a transaction has not begun")
	}

	err := h.tx.Rollback()
	if err != nil {
		return errorutil.HandledWithMessage(err, "failed to rollback transaction")
	}

	h.tx = nil
	return nil
}

func (h *Handle) openDB() (*sqlx.DB, error) {
	if h.database == nil {
		db, err := sqlx.Open("postgres", h.Config.DatabaseURL)
		if err != nil {
			return nil, errorutil.HandledWithMessage(err, "failed to connect to database")
		}

		db.SetMaxOpenConns(h.Config.MaxOpenConn)
		db.SetMaxIdleConns(h.Config.MaxIdleConn)
		db.SetConnMaxLifetime(time.Second * time.Duration(h.Config.ConnMaxLifetimeSeconds))
		db.SetConnMaxIdleTime(time.Second * time.Duration(h.Config.ConnMaxIdleTimeSeconds))

		h.database = db
	}
	return h.database, nil
}
