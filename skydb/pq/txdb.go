package pq

import (
	"github.com/oursky/skygear/skydb"
)

func (db *database) Begin() (err error) {
	return db.c.Begin()
}

func (db *database) Commit() (err error) {
	return db.c.Commit()
}

func (db *database) Rollback() (err error) {
	return db.c.Rollback()
}

var _ skydb.TxDatabase = &database{}
