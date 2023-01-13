package workflow

import (
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
)

const savepointName = "workflow"

type SavePointImpl struct {
	SQLExecutor *appdb.SQLExecutor
}

func (p SavePointImpl) Begin() (err error) {
	_, err = p.SQLExecutor.ExecWith(savePointNew(savepointName))
	if err != nil {
		return
	}
	return
}

func (p SavePointImpl) Rollback() (err error) {
	_, err = p.SQLExecutor.ExecWith(savePointRollback(savepointName))
	if err != nil {
		return
	}
	return
}

func (p SavePointImpl) Commit() (err error) {
	_, err = p.SQLExecutor.ExecWith(savePointRelease(savepointName))
	if err != nil {
		return
	}
	return
}

type savePointNew string

// nolint:golint
func (s savePointNew) ToSql() (string, []interface{}, error) {
	return "SAVEPOINT " + string(s), nil, nil
}

type savePointRelease string

// nolint:golint
func (s savePointRelease) ToSql() (string, []interface{}, error) {
	return "RELEASE SAVEPOINT " + string(s), nil, nil
}

type savePointRollback string

// nolint:golint
func (s savePointRollback) ToSql() (string, []interface{}, error) {
	return "ROLLBACK TO SAVEPOINT " + string(s), nil, nil
}
