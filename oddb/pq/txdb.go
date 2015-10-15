package pq

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/oursky/skygear/oddb"
)

func (db *database) Begin() (err error) {
	if db.txDone {
		return oddb.ErrDatabaseTxDone
	}

	switch dbx := db.Db.(type) {
	default:
		panic(fmt.Sprintf("got unrecgonized type(db.Db) = %T", db.Db))
	case *sqlx.DB:
		var tx *sqlx.Tx
		tx, err = dbx.Beginx()
		if err == nil {
			db.Db = tx
		}
	case *sqlx.Tx:
		err = oddb.ErrDatabaseTxDidBegin
	}

	return
}

func (db *database) Commit() (err error) {
	if db.txDone {
		return oddb.ErrDatabaseTxDone
	}

	switch dbx := db.Db.(type) {
	default:
		panic(fmt.Sprintf("got unrecgonized type(db.Db) = %T", db.Db))
	case *sqlx.DB:
		err = oddb.ErrDatabaseTxDidNotBegin
	case *sqlx.Tx:
		err = dbx.Commit()
		if err == nil {
			db.txDone = true
		}
	}

	return
}

func (db *database) Rollback() (err error) {
	if db.txDone {
		return oddb.ErrDatabaseTxDone
	}

	switch dbx := db.Db.(type) {
	default:
		panic(fmt.Sprintf("got unrecgonized type(db.Db) = %T", db.Db))
	case *sqlx.DB:
		err = oddb.ErrDatabaseTxDidNotBegin
	case *sqlx.Tx:
		err = dbx.Rollback()
		if err == nil {
			db.txDone = true
		}
	}

	return
}

var _ oddb.TxDatabase = &database{}
