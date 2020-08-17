package interaction

type savePoint string

func (s savePoint) New() savePointNew           { return savePointNew(s) }
func (s savePoint) Release() savePointRelease   { return savePointRelease(s) }
func (s savePoint) Rollback() savePointRollback { return savePointRollback(s) }

type savePointNew savePoint

// nolint:golint
func (s savePointNew) ToSql() (string, []interface{}, error) {
	return "SAVEPOINT " + string(s), nil, nil
}

type savePointRelease savePoint

// nolint:golint
func (s savePointRelease) ToSql() (string, []interface{}, error) {
	return "RELEASE SAVEPOINT " + string(s), nil, nil
}

type savePointRollback savePoint

// nolint:golint
func (s savePointRollback) ToSql() (string, []interface{}, error) {
	return "ROLLBACK TO SAVEPOINT " + string(s), nil, nil
}
