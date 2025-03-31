package databasesqlwrapper

import (
	context "context"
	driver "database/sql/driver"
)

type ConnUnwrapper interface {
	UnwrapConn() driver.Conn
}

func UnwrapConn(w driver.Conn) driver.Conn {
	if ww, ok := w.(ConnUnwrapper); ok {
		return UnwrapConn(ww.UnwrapConn())
	} else {
		return w
	}
}

type Conn_Begin func() (driver.Tx, error)
type Conn_Close func() error
type Conn_Prepare func(string) (driver.Stmt, error)
type Conn_BeginTx func(context.Context, driver.TxOptions) (driver.Tx, error)
type Conn_PrepareContext func(context.Context, string) (driver.Stmt, error)
type Conn_Exec func(string, []driver.Value) (driver.Result, error)
type Conn_ExecContext func(context.Context, string, []driver.NamedValue) (driver.Result, error)
type Conn_CheckNamedValue func(*driver.NamedValue) error
type Conn_Ping func(context.Context) error
type Conn_Query func(string, []driver.Value) (driver.Rows, error)
type Conn_QueryContext func(context.Context, string, []driver.NamedValue) (driver.Rows, error)
type Conn_ResetSession func(context.Context) error
type Conn_IsValid func() bool

type ConnInterceptor struct {
	Begin           func(Conn_Begin) Conn_Begin
	Close           func(Conn_Close) Conn_Close
	Prepare         func(Conn_Prepare) Conn_Prepare
	BeginTx         func(Conn_BeginTx) Conn_BeginTx
	PrepareContext  func(Conn_PrepareContext) Conn_PrepareContext
	Exec            func(Conn_Exec) Conn_Exec
	ExecContext     func(Conn_ExecContext) Conn_ExecContext
	CheckNamedValue func(Conn_CheckNamedValue) Conn_CheckNamedValue
	Ping            func(Conn_Ping) Conn_Ping
	Query           func(Conn_Query) Conn_Query
	QueryContext    func(Conn_QueryContext) Conn_QueryContext
	ResetSession    func(Conn_ResetSession) Conn_ResetSession
	IsValid         func(Conn_IsValid) Conn_IsValid
}

type ConnWrapper struct {
	wrapped     driver.Conn
	interceptor ConnInterceptor
}

func (w *ConnWrapper) UnwrapConn() driver.Conn {
	return w.wrapped
}

func (w *ConnWrapper) Begin() (driver.Tx, error) {
	f := w.wrapped.(driver.Conn).Begin
	if w.interceptor.Begin != nil {
		f = w.interceptor.Begin(f)
	}
	return f()
}

func (w *ConnWrapper) Close() error {
	f := w.wrapped.(driver.Conn).Close
	if w.interceptor.Close != nil {
		f = w.interceptor.Close(f)
	}
	return f()
}

func (w *ConnWrapper) Prepare(a0 string) (driver.Stmt, error) {
	f := w.wrapped.(driver.Conn).Prepare
	if w.interceptor.Prepare != nil {
		f = w.interceptor.Prepare(f)
	}
	return f(a0)
}

func (w *ConnWrapper) BeginTx(a0 context.Context, a1 driver.TxOptions) (driver.Tx, error) {
	f := w.wrapped.(driver.ConnBeginTx).BeginTx
	if w.interceptor.BeginTx != nil {
		f = w.interceptor.BeginTx(f)
	}
	return f(a0, a1)
}

func (w *ConnWrapper) PrepareContext(a0 context.Context, a1 string) (driver.Stmt, error) {
	f := w.wrapped.(driver.ConnPrepareContext).PrepareContext
	if w.interceptor.PrepareContext != nil {
		f = w.interceptor.PrepareContext(f)
	}
	return f(a0, a1)
}

func (w *ConnWrapper) Exec(a0 string, a1 []driver.Value) (driver.Result, error) {
	f := w.wrapped.(driver.Execer).Exec
	if w.interceptor.Exec != nil {
		f = w.interceptor.Exec(f)
	}
	return f(a0, a1)
}

func (w *ConnWrapper) ExecContext(a0 context.Context, a1 string, a2 []driver.NamedValue) (driver.Result, error) {
	f := w.wrapped.(driver.ExecerContext).ExecContext
	if w.interceptor.ExecContext != nil {
		f = w.interceptor.ExecContext(f)
	}
	return f(a0, a1, a2)
}

func (w *ConnWrapper) CheckNamedValue(a0 *driver.NamedValue) error {
	f := w.wrapped.(driver.NamedValueChecker).CheckNamedValue
	if w.interceptor.CheckNamedValue != nil {
		f = w.interceptor.CheckNamedValue(f)
	}
	return f(a0)
}

func (w *ConnWrapper) Ping(a0 context.Context) error {
	f := w.wrapped.(driver.Pinger).Ping
	if w.interceptor.Ping != nil {
		f = w.interceptor.Ping(f)
	}
	return f(a0)
}

func (w *ConnWrapper) Query(a0 string, a1 []driver.Value) (driver.Rows, error) {
	f := w.wrapped.(driver.Queryer).Query
	if w.interceptor.Query != nil {
		f = w.interceptor.Query(f)
	}
	return f(a0, a1)
}

func (w *ConnWrapper) QueryContext(a0 context.Context, a1 string, a2 []driver.NamedValue) (driver.Rows, error) {
	f := w.wrapped.(driver.QueryerContext).QueryContext
	if w.interceptor.QueryContext != nil {
		f = w.interceptor.QueryContext(f)
	}
	return f(a0, a1, a2)
}

func (w *ConnWrapper) ResetSession(a0 context.Context) error {
	f := w.wrapped.(driver.SessionResetter).ResetSession
	if w.interceptor.ResetSession != nil {
		f = w.interceptor.ResetSession(f)
	}
	return f(a0)
}

func (w *ConnWrapper) IsValid() bool {
	f := w.wrapped.(driver.Validator).IsValid
	if w.interceptor.IsValid != nil {
		f = w.interceptor.IsValid(f)
	}
	return f()
}

func WrapConn(wrapped driver.Conn, interceptor ConnInterceptor) driver.Conn {
	w := &ConnWrapper{wrapped: wrapped, interceptor: interceptor}
	_, ok0 := wrapped.(driver.ConnBeginTx)
	_, ok1 := wrapped.(driver.ConnPrepareContext)
	_, ok2 := wrapped.(driver.Execer)
	_, ok3 := wrapped.(driver.ExecerContext)
	_, ok4 := wrapped.(driver.NamedValueChecker)
	_, ok5 := wrapped.(driver.Pinger)
	_, ok6 := wrapped.(driver.Queryer)
	_, ok7 := wrapped.(driver.QueryerContext)
	_, ok8 := wrapped.(driver.SessionResetter)
	_, ok9 := wrapped.(driver.Validator)
	switch {
	// combination 1/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && !ok4 && !ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
		}{w, w}
	// combination 2/1024
	case ok0 && !ok1 && !ok2 && !ok3 && !ok4 && !ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
		}{w, w, w}
	// combination 3/1024
	case !ok0 && ok1 && !ok2 && !ok3 && !ok4 && !ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
		}{w, w, w}
	// combination 4/1024
	case ok0 && ok1 && !ok2 && !ok3 && !ok4 && !ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
		}{w, w, w, w}
	// combination 5/1024
	case !ok0 && !ok1 && ok2 && !ok3 && !ok4 && !ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
		}{w, w, w}
	// combination 6/1024
	case ok0 && !ok1 && ok2 && !ok3 && !ok4 && !ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
		}{w, w, w, w}
	// combination 7/1024
	case !ok0 && ok1 && ok2 && !ok3 && !ok4 && !ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
		}{w, w, w, w}
	// combination 8/1024
	case ok0 && ok1 && ok2 && !ok3 && !ok4 && !ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
		}{w, w, w, w, w}
	// combination 9/1024
	case !ok0 && !ok1 && !ok2 && ok3 && !ok4 && !ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
		}{w, w, w}
	// combination 10/1024
	case ok0 && !ok1 && !ok2 && ok3 && !ok4 && !ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
		}{w, w, w, w}
	// combination 11/1024
	case !ok0 && ok1 && !ok2 && ok3 && !ok4 && !ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
		}{w, w, w, w}
	// combination 12/1024
	case ok0 && ok1 && !ok2 && ok3 && !ok4 && !ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
		}{w, w, w, w, w}
	// combination 13/1024
	case !ok0 && !ok1 && ok2 && ok3 && !ok4 && !ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
		}{w, w, w, w}
	// combination 14/1024
	case ok0 && !ok1 && ok2 && ok3 && !ok4 && !ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
		}{w, w, w, w, w}
	// combination 15/1024
	case !ok0 && ok1 && ok2 && ok3 && !ok4 && !ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
		}{w, w, w, w, w}
	// combination 16/1024
	case ok0 && ok1 && ok2 && ok3 && !ok4 && !ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
		}{w, w, w, w, w, w}
	// combination 17/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && ok4 && !ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.NamedValueChecker
		}{w, w, w}
	// combination 18/1024
	case ok0 && !ok1 && !ok2 && !ok3 && ok4 && !ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.NamedValueChecker
		}{w, w, w, w}
	// combination 19/1024
	case !ok0 && ok1 && !ok2 && !ok3 && ok4 && !ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.NamedValueChecker
		}{w, w, w, w}
	// combination 20/1024
	case ok0 && ok1 && !ok2 && !ok3 && ok4 && !ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.NamedValueChecker
		}{w, w, w, w, w}
	// combination 21/1024
	case !ok0 && !ok1 && ok2 && !ok3 && ok4 && !ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.NamedValueChecker
		}{w, w, w, w}
	// combination 22/1024
	case ok0 && !ok1 && ok2 && !ok3 && ok4 && !ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.NamedValueChecker
		}{w, w, w, w, w}
	// combination 23/1024
	case !ok0 && ok1 && ok2 && !ok3 && ok4 && !ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
		}{w, w, w, w, w}
	// combination 24/1024
	case ok0 && ok1 && ok2 && !ok3 && ok4 && !ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
		}{w, w, w, w, w, w}
	// combination 25/1024
	case !ok0 && !ok1 && !ok2 && ok3 && ok4 && !ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.NamedValueChecker
		}{w, w, w, w}
	// combination 26/1024
	case ok0 && !ok1 && !ok2 && ok3 && ok4 && !ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.NamedValueChecker
		}{w, w, w, w, w}
	// combination 27/1024
	case !ok0 && ok1 && !ok2 && ok3 && ok4 && !ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
		}{w, w, w, w, w}
	// combination 28/1024
	case ok0 && ok1 && !ok2 && ok3 && ok4 && !ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
		}{w, w, w, w, w, w}
	// combination 29/1024
	case !ok0 && !ok1 && ok2 && ok3 && ok4 && !ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
		}{w, w, w, w, w}
	// combination 30/1024
	case ok0 && !ok1 && ok2 && ok3 && ok4 && !ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
		}{w, w, w, w, w, w}
	// combination 31/1024
	case !ok0 && ok1 && ok2 && ok3 && ok4 && !ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
		}{w, w, w, w, w, w}
	// combination 32/1024
	case ok0 && ok1 && ok2 && ok3 && ok4 && !ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
		}{w, w, w, w, w, w, w}
	// combination 33/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && !ok4 && ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Pinger
		}{w, w, w}
	// combination 34/1024
	case ok0 && !ok1 && !ok2 && !ok3 && !ok4 && ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Pinger
		}{w, w, w, w}
	// combination 35/1024
	case !ok0 && ok1 && !ok2 && !ok3 && !ok4 && ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Pinger
		}{w, w, w, w}
	// combination 36/1024
	case ok0 && ok1 && !ok2 && !ok3 && !ok4 && ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Pinger
		}{w, w, w, w, w}
	// combination 37/1024
	case !ok0 && !ok1 && ok2 && !ok3 && !ok4 && ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.Pinger
		}{w, w, w, w}
	// combination 38/1024
	case ok0 && !ok1 && ok2 && !ok3 && !ok4 && ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.Pinger
		}{w, w, w, w, w}
	// combination 39/1024
	case !ok0 && ok1 && ok2 && !ok3 && !ok4 && ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.Pinger
		}{w, w, w, w, w}
	// combination 40/1024
	case ok0 && ok1 && ok2 && !ok3 && !ok4 && ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.Pinger
		}{w, w, w, w, w, w}
	// combination 41/1024
	case !ok0 && !ok1 && !ok2 && ok3 && !ok4 && ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.Pinger
		}{w, w, w, w}
	// combination 42/1024
	case ok0 && !ok1 && !ok2 && ok3 && !ok4 && ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.Pinger
		}{w, w, w, w, w}
	// combination 43/1024
	case !ok0 && ok1 && !ok2 && ok3 && !ok4 && ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Pinger
		}{w, w, w, w, w}
	// combination 44/1024
	case ok0 && ok1 && !ok2 && ok3 && !ok4 && ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Pinger
		}{w, w, w, w, w, w}
	// combination 45/1024
	case !ok0 && !ok1 && ok2 && ok3 && !ok4 && ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.Pinger
		}{w, w, w, w, w}
	// combination 46/1024
	case ok0 && !ok1 && ok2 && ok3 && !ok4 && ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.Pinger
		}{w, w, w, w, w, w}
	// combination 47/1024
	case !ok0 && ok1 && ok2 && ok3 && !ok4 && ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Pinger
		}{w, w, w, w, w, w}
	// combination 48/1024
	case ok0 && ok1 && ok2 && ok3 && !ok4 && ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Pinger
		}{w, w, w, w, w, w, w}
	// combination 49/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && ok4 && ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.NamedValueChecker
			driver.Pinger
		}{w, w, w, w}
	// combination 50/1024
	case ok0 && !ok1 && !ok2 && !ok3 && ok4 && ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.NamedValueChecker
			driver.Pinger
		}{w, w, w, w, w}
	// combination 51/1024
	case !ok0 && ok1 && !ok2 && !ok3 && ok4 && ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Pinger
		}{w, w, w, w, w}
	// combination 52/1024
	case ok0 && ok1 && !ok2 && !ok3 && ok4 && ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Pinger
		}{w, w, w, w, w, w}
	// combination 53/1024
	case !ok0 && !ok1 && ok2 && !ok3 && ok4 && ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
		}{w, w, w, w, w}
	// combination 54/1024
	case ok0 && !ok1 && ok2 && !ok3 && ok4 && ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
		}{w, w, w, w, w, w}
	// combination 55/1024
	case !ok0 && ok1 && ok2 && !ok3 && ok4 && ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
		}{w, w, w, w, w, w}
	// combination 56/1024
	case ok0 && ok1 && ok2 && !ok3 && ok4 && ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
		}{w, w, w, w, w, w, w}
	// combination 57/1024
	case !ok0 && !ok1 && !ok2 && ok3 && ok4 && ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
		}{w, w, w, w, w}
	// combination 58/1024
	case ok0 && !ok1 && !ok2 && ok3 && ok4 && ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
		}{w, w, w, w, w, w}
	// combination 59/1024
	case !ok0 && ok1 && !ok2 && ok3 && ok4 && ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
		}{w, w, w, w, w, w}
	// combination 60/1024
	case ok0 && ok1 && !ok2 && ok3 && ok4 && ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
		}{w, w, w, w, w, w, w}
	// combination 61/1024
	case !ok0 && !ok1 && ok2 && ok3 && ok4 && ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
		}{w, w, w, w, w, w}
	// combination 62/1024
	case ok0 && !ok1 && ok2 && ok3 && ok4 && ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
		}{w, w, w, w, w, w, w}
	// combination 63/1024
	case !ok0 && ok1 && ok2 && ok3 && ok4 && ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
		}{w, w, w, w, w, w, w}
	// combination 64/1024
	case ok0 && ok1 && ok2 && ok3 && ok4 && ok5 && !ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
		}{w, w, w, w, w, w, w, w}
	// combination 65/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && !ok4 && !ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Queryer
		}{w, w, w}
	// combination 66/1024
	case ok0 && !ok1 && !ok2 && !ok3 && !ok4 && !ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Queryer
		}{w, w, w, w}
	// combination 67/1024
	case !ok0 && ok1 && !ok2 && !ok3 && !ok4 && !ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Queryer
		}{w, w, w, w}
	// combination 68/1024
	case ok0 && ok1 && !ok2 && !ok3 && !ok4 && !ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Queryer
		}{w, w, w, w, w}
	// combination 69/1024
	case !ok0 && !ok1 && ok2 && !ok3 && !ok4 && !ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.Queryer
		}{w, w, w, w}
	// combination 70/1024
	case ok0 && !ok1 && ok2 && !ok3 && !ok4 && !ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.Queryer
		}{w, w, w, w, w}
	// combination 71/1024
	case !ok0 && ok1 && ok2 && !ok3 && !ok4 && !ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.Queryer
		}{w, w, w, w, w}
	// combination 72/1024
	case ok0 && ok1 && ok2 && !ok3 && !ok4 && !ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.Queryer
		}{w, w, w, w, w, w}
	// combination 73/1024
	case !ok0 && !ok1 && !ok2 && ok3 && !ok4 && !ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.Queryer
		}{w, w, w, w}
	// combination 74/1024
	case ok0 && !ok1 && !ok2 && ok3 && !ok4 && !ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.Queryer
		}{w, w, w, w, w}
	// combination 75/1024
	case !ok0 && ok1 && !ok2 && ok3 && !ok4 && !ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Queryer
		}{w, w, w, w, w}
	// combination 76/1024
	case ok0 && ok1 && !ok2 && ok3 && !ok4 && !ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Queryer
		}{w, w, w, w, w, w}
	// combination 77/1024
	case !ok0 && !ok1 && ok2 && ok3 && !ok4 && !ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.Queryer
		}{w, w, w, w, w}
	// combination 78/1024
	case ok0 && !ok1 && ok2 && ok3 && !ok4 && !ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.Queryer
		}{w, w, w, w, w, w}
	// combination 79/1024
	case !ok0 && ok1 && ok2 && ok3 && !ok4 && !ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Queryer
		}{w, w, w, w, w, w}
	// combination 80/1024
	case ok0 && ok1 && ok2 && ok3 && !ok4 && !ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Queryer
		}{w, w, w, w, w, w, w}
	// combination 81/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && ok4 && !ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.NamedValueChecker
			driver.Queryer
		}{w, w, w, w}
	// combination 82/1024
	case ok0 && !ok1 && !ok2 && !ok3 && ok4 && !ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.NamedValueChecker
			driver.Queryer
		}{w, w, w, w, w}
	// combination 83/1024
	case !ok0 && ok1 && !ok2 && !ok3 && ok4 && !ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Queryer
		}{w, w, w, w, w}
	// combination 84/1024
	case ok0 && ok1 && !ok2 && !ok3 && ok4 && !ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Queryer
		}{w, w, w, w, w, w}
	// combination 85/1024
	case !ok0 && !ok1 && ok2 && !ok3 && ok4 && !ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.NamedValueChecker
			driver.Queryer
		}{w, w, w, w, w}
	// combination 86/1024
	case ok0 && !ok1 && ok2 && !ok3 && ok4 && !ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.NamedValueChecker
			driver.Queryer
		}{w, w, w, w, w, w}
	// combination 87/1024
	case !ok0 && ok1 && ok2 && !ok3 && ok4 && !ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Queryer
		}{w, w, w, w, w, w}
	// combination 88/1024
	case ok0 && ok1 && ok2 && !ok3 && ok4 && !ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Queryer
		}{w, w, w, w, w, w, w}
	// combination 89/1024
	case !ok0 && !ok1 && !ok2 && ok3 && ok4 && !ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
		}{w, w, w, w, w}
	// combination 90/1024
	case ok0 && !ok1 && !ok2 && ok3 && ok4 && !ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
		}{w, w, w, w, w, w}
	// combination 91/1024
	case !ok0 && ok1 && !ok2 && ok3 && ok4 && !ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
		}{w, w, w, w, w, w}
	// combination 92/1024
	case ok0 && ok1 && !ok2 && ok3 && ok4 && !ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
		}{w, w, w, w, w, w, w}
	// combination 93/1024
	case !ok0 && !ok1 && ok2 && ok3 && ok4 && !ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
		}{w, w, w, w, w, w}
	// combination 94/1024
	case ok0 && !ok1 && ok2 && ok3 && ok4 && !ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
		}{w, w, w, w, w, w, w}
	// combination 95/1024
	case !ok0 && ok1 && ok2 && ok3 && ok4 && !ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
		}{w, w, w, w, w, w, w}
	// combination 96/1024
	case ok0 && ok1 && ok2 && ok3 && ok4 && !ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
		}{w, w, w, w, w, w, w, w}
	// combination 97/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && !ok4 && ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Pinger
			driver.Queryer
		}{w, w, w, w}
	// combination 98/1024
	case ok0 && !ok1 && !ok2 && !ok3 && !ok4 && ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Pinger
			driver.Queryer
		}{w, w, w, w, w}
	// combination 99/1024
	case !ok0 && ok1 && !ok2 && !ok3 && !ok4 && ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Pinger
			driver.Queryer
		}{w, w, w, w, w}
	// combination 100/1024
	case ok0 && ok1 && !ok2 && !ok3 && !ok4 && ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Pinger
			driver.Queryer
		}{w, w, w, w, w, w}
	// combination 101/1024
	case !ok0 && !ok1 && ok2 && !ok3 && !ok4 && ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.Pinger
			driver.Queryer
		}{w, w, w, w, w}
	// combination 102/1024
	case ok0 && !ok1 && ok2 && !ok3 && !ok4 && ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.Pinger
			driver.Queryer
		}{w, w, w, w, w, w}
	// combination 103/1024
	case !ok0 && ok1 && ok2 && !ok3 && !ok4 && ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.Pinger
			driver.Queryer
		}{w, w, w, w, w, w}
	// combination 104/1024
	case ok0 && ok1 && ok2 && !ok3 && !ok4 && ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.Pinger
			driver.Queryer
		}{w, w, w, w, w, w, w}
	// combination 105/1024
	case !ok0 && !ok1 && !ok2 && ok3 && !ok4 && ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
		}{w, w, w, w, w}
	// combination 106/1024
	case ok0 && !ok1 && !ok2 && ok3 && !ok4 && ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
		}{w, w, w, w, w, w}
	// combination 107/1024
	case !ok0 && ok1 && !ok2 && ok3 && !ok4 && ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
		}{w, w, w, w, w, w}
	// combination 108/1024
	case ok0 && ok1 && !ok2 && ok3 && !ok4 && ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
		}{w, w, w, w, w, w, w}
	// combination 109/1024
	case !ok0 && !ok1 && ok2 && ok3 && !ok4 && ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
		}{w, w, w, w, w, w}
	// combination 110/1024
	case ok0 && !ok1 && ok2 && ok3 && !ok4 && ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
		}{w, w, w, w, w, w, w}
	// combination 111/1024
	case !ok0 && ok1 && ok2 && ok3 && !ok4 && ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
		}{w, w, w, w, w, w, w}
	// combination 112/1024
	case ok0 && ok1 && ok2 && ok3 && !ok4 && ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
		}{w, w, w, w, w, w, w, w}
	// combination 113/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && ok4 && ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
		}{w, w, w, w, w}
	// combination 114/1024
	case ok0 && !ok1 && !ok2 && !ok3 && ok4 && ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
		}{w, w, w, w, w, w}
	// combination 115/1024
	case !ok0 && ok1 && !ok2 && !ok3 && ok4 && ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
		}{w, w, w, w, w, w}
	// combination 116/1024
	case ok0 && ok1 && !ok2 && !ok3 && ok4 && ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
		}{w, w, w, w, w, w, w}
	// combination 117/1024
	case !ok0 && !ok1 && ok2 && !ok3 && ok4 && ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
		}{w, w, w, w, w, w}
	// combination 118/1024
	case ok0 && !ok1 && ok2 && !ok3 && ok4 && ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
		}{w, w, w, w, w, w, w}
	// combination 119/1024
	case !ok0 && ok1 && ok2 && !ok3 && ok4 && ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
		}{w, w, w, w, w, w, w}
	// combination 120/1024
	case ok0 && ok1 && ok2 && !ok3 && ok4 && ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
		}{w, w, w, w, w, w, w, w}
	// combination 121/1024
	case !ok0 && !ok1 && !ok2 && ok3 && ok4 && ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
		}{w, w, w, w, w, w}
	// combination 122/1024
	case ok0 && !ok1 && !ok2 && ok3 && ok4 && ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
		}{w, w, w, w, w, w, w}
	// combination 123/1024
	case !ok0 && ok1 && !ok2 && ok3 && ok4 && ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
		}{w, w, w, w, w, w, w}
	// combination 124/1024
	case ok0 && ok1 && !ok2 && ok3 && ok4 && ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
		}{w, w, w, w, w, w, w, w}
	// combination 125/1024
	case !ok0 && !ok1 && ok2 && ok3 && ok4 && ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
		}{w, w, w, w, w, w, w}
	// combination 126/1024
	case ok0 && !ok1 && ok2 && ok3 && ok4 && ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
		}{w, w, w, w, w, w, w, w}
	// combination 127/1024
	case !ok0 && ok1 && ok2 && ok3 && ok4 && ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
		}{w, w, w, w, w, w, w, w}
	// combination 128/1024
	case ok0 && ok1 && ok2 && ok3 && ok4 && ok5 && ok6 && !ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
		}{w, w, w, w, w, w, w, w, w}
	// combination 129/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && !ok4 && !ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.QueryerContext
		}{w, w, w}
	// combination 130/1024
	case ok0 && !ok1 && !ok2 && !ok3 && !ok4 && !ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.QueryerContext
		}{w, w, w, w}
	// combination 131/1024
	case !ok0 && ok1 && !ok2 && !ok3 && !ok4 && !ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.QueryerContext
		}{w, w, w, w}
	// combination 132/1024
	case ok0 && ok1 && !ok2 && !ok3 && !ok4 && !ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.QueryerContext
		}{w, w, w, w, w}
	// combination 133/1024
	case !ok0 && !ok1 && ok2 && !ok3 && !ok4 && !ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.QueryerContext
		}{w, w, w, w}
	// combination 134/1024
	case ok0 && !ok1 && ok2 && !ok3 && !ok4 && !ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.QueryerContext
		}{w, w, w, w, w}
	// combination 135/1024
	case !ok0 && ok1 && ok2 && !ok3 && !ok4 && !ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.QueryerContext
		}{w, w, w, w, w}
	// combination 136/1024
	case ok0 && ok1 && ok2 && !ok3 && !ok4 && !ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.QueryerContext
		}{w, w, w, w, w, w}
	// combination 137/1024
	case !ok0 && !ok1 && !ok2 && ok3 && !ok4 && !ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.QueryerContext
		}{w, w, w, w}
	// combination 138/1024
	case ok0 && !ok1 && !ok2 && ok3 && !ok4 && !ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.QueryerContext
		}{w, w, w, w, w}
	// combination 139/1024
	case !ok0 && ok1 && !ok2 && ok3 && !ok4 && !ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.QueryerContext
		}{w, w, w, w, w}
	// combination 140/1024
	case ok0 && ok1 && !ok2 && ok3 && !ok4 && !ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.QueryerContext
		}{w, w, w, w, w, w}
	// combination 141/1024
	case !ok0 && !ok1 && ok2 && ok3 && !ok4 && !ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.QueryerContext
		}{w, w, w, w, w}
	// combination 142/1024
	case ok0 && !ok1 && ok2 && ok3 && !ok4 && !ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.QueryerContext
		}{w, w, w, w, w, w}
	// combination 143/1024
	case !ok0 && ok1 && ok2 && ok3 && !ok4 && !ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.QueryerContext
		}{w, w, w, w, w, w}
	// combination 144/1024
	case ok0 && ok1 && ok2 && ok3 && !ok4 && !ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.QueryerContext
		}{w, w, w, w, w, w, w}
	// combination 145/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && ok4 && !ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.NamedValueChecker
			driver.QueryerContext
		}{w, w, w, w}
	// combination 146/1024
	case ok0 && !ok1 && !ok2 && !ok3 && ok4 && !ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.NamedValueChecker
			driver.QueryerContext
		}{w, w, w, w, w}
	// combination 147/1024
	case !ok0 && ok1 && !ok2 && !ok3 && ok4 && !ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.QueryerContext
		}{w, w, w, w, w}
	// combination 148/1024
	case ok0 && ok1 && !ok2 && !ok3 && ok4 && !ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.QueryerContext
		}{w, w, w, w, w, w}
	// combination 149/1024
	case !ok0 && !ok1 && ok2 && !ok3 && ok4 && !ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.NamedValueChecker
			driver.QueryerContext
		}{w, w, w, w, w}
	// combination 150/1024
	case ok0 && !ok1 && ok2 && !ok3 && ok4 && !ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.NamedValueChecker
			driver.QueryerContext
		}{w, w, w, w, w, w}
	// combination 151/1024
	case !ok0 && ok1 && ok2 && !ok3 && ok4 && !ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.QueryerContext
		}{w, w, w, w, w, w}
	// combination 152/1024
	case ok0 && ok1 && ok2 && !ok3 && ok4 && !ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.QueryerContext
		}{w, w, w, w, w, w, w}
	// combination 153/1024
	case !ok0 && !ok1 && !ok2 && ok3 && ok4 && !ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.NamedValueChecker
			driver.QueryerContext
		}{w, w, w, w, w}
	// combination 154/1024
	case ok0 && !ok1 && !ok2 && ok3 && ok4 && !ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.NamedValueChecker
			driver.QueryerContext
		}{w, w, w, w, w, w}
	// combination 155/1024
	case !ok0 && ok1 && !ok2 && ok3 && ok4 && !ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.QueryerContext
		}{w, w, w, w, w, w}
	// combination 156/1024
	case ok0 && ok1 && !ok2 && ok3 && ok4 && !ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.QueryerContext
		}{w, w, w, w, w, w, w}
	// combination 157/1024
	case !ok0 && !ok1 && ok2 && ok3 && ok4 && !ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.QueryerContext
		}{w, w, w, w, w, w}
	// combination 158/1024
	case ok0 && !ok1 && ok2 && ok3 && ok4 && !ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.QueryerContext
		}{w, w, w, w, w, w, w}
	// combination 159/1024
	case !ok0 && ok1 && ok2 && ok3 && ok4 && !ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.QueryerContext
		}{w, w, w, w, w, w, w}
	// combination 160/1024
	case ok0 && ok1 && ok2 && ok3 && ok4 && !ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.QueryerContext
		}{w, w, w, w, w, w, w, w}
	// combination 161/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && !ok4 && ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Pinger
			driver.QueryerContext
		}{w, w, w, w}
	// combination 162/1024
	case ok0 && !ok1 && !ok2 && !ok3 && !ok4 && ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Pinger
			driver.QueryerContext
		}{w, w, w, w, w}
	// combination 163/1024
	case !ok0 && ok1 && !ok2 && !ok3 && !ok4 && ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Pinger
			driver.QueryerContext
		}{w, w, w, w, w}
	// combination 164/1024
	case ok0 && ok1 && !ok2 && !ok3 && !ok4 && ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Pinger
			driver.QueryerContext
		}{w, w, w, w, w, w}
	// combination 165/1024
	case !ok0 && !ok1 && ok2 && !ok3 && !ok4 && ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.Pinger
			driver.QueryerContext
		}{w, w, w, w, w}
	// combination 166/1024
	case ok0 && !ok1 && ok2 && !ok3 && !ok4 && ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.Pinger
			driver.QueryerContext
		}{w, w, w, w, w, w}
	// combination 167/1024
	case !ok0 && ok1 && ok2 && !ok3 && !ok4 && ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.Pinger
			driver.QueryerContext
		}{w, w, w, w, w, w}
	// combination 168/1024
	case ok0 && ok1 && ok2 && !ok3 && !ok4 && ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.Pinger
			driver.QueryerContext
		}{w, w, w, w, w, w, w}
	// combination 169/1024
	case !ok0 && !ok1 && !ok2 && ok3 && !ok4 && ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.Pinger
			driver.QueryerContext
		}{w, w, w, w, w}
	// combination 170/1024
	case ok0 && !ok1 && !ok2 && ok3 && !ok4 && ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.Pinger
			driver.QueryerContext
		}{w, w, w, w, w, w}
	// combination 171/1024
	case !ok0 && ok1 && !ok2 && ok3 && !ok4 && ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Pinger
			driver.QueryerContext
		}{w, w, w, w, w, w}
	// combination 172/1024
	case ok0 && ok1 && !ok2 && ok3 && !ok4 && ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Pinger
			driver.QueryerContext
		}{w, w, w, w, w, w, w}
	// combination 173/1024
	case !ok0 && !ok1 && ok2 && ok3 && !ok4 && ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.QueryerContext
		}{w, w, w, w, w, w}
	// combination 174/1024
	case ok0 && !ok1 && ok2 && ok3 && !ok4 && ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.QueryerContext
		}{w, w, w, w, w, w, w}
	// combination 175/1024
	case !ok0 && ok1 && ok2 && ok3 && !ok4 && ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.QueryerContext
		}{w, w, w, w, w, w, w}
	// combination 176/1024
	case ok0 && ok1 && ok2 && ok3 && !ok4 && ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.QueryerContext
		}{w, w, w, w, w, w, w, w}
	// combination 177/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && ok4 && ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
		}{w, w, w, w, w}
	// combination 178/1024
	case ok0 && !ok1 && !ok2 && !ok3 && ok4 && ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
		}{w, w, w, w, w, w}
	// combination 179/1024
	case !ok0 && ok1 && !ok2 && !ok3 && ok4 && ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
		}{w, w, w, w, w, w}
	// combination 180/1024
	case ok0 && ok1 && !ok2 && !ok3 && ok4 && ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
		}{w, w, w, w, w, w, w}
	// combination 181/1024
	case !ok0 && !ok1 && ok2 && !ok3 && ok4 && ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
		}{w, w, w, w, w, w}
	// combination 182/1024
	case ok0 && !ok1 && ok2 && !ok3 && ok4 && ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
		}{w, w, w, w, w, w, w}
	// combination 183/1024
	case !ok0 && ok1 && ok2 && !ok3 && ok4 && ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
		}{w, w, w, w, w, w, w}
	// combination 184/1024
	case ok0 && ok1 && ok2 && !ok3 && ok4 && ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
		}{w, w, w, w, w, w, w, w}
	// combination 185/1024
	case !ok0 && !ok1 && !ok2 && ok3 && ok4 && ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
		}{w, w, w, w, w, w}
	// combination 186/1024
	case ok0 && !ok1 && !ok2 && ok3 && ok4 && ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
		}{w, w, w, w, w, w, w}
	// combination 187/1024
	case !ok0 && ok1 && !ok2 && ok3 && ok4 && ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
		}{w, w, w, w, w, w, w}
	// combination 188/1024
	case ok0 && ok1 && !ok2 && ok3 && ok4 && ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
		}{w, w, w, w, w, w, w, w}
	// combination 189/1024
	case !ok0 && !ok1 && ok2 && ok3 && ok4 && ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
		}{w, w, w, w, w, w, w}
	// combination 190/1024
	case ok0 && !ok1 && ok2 && ok3 && ok4 && ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
		}{w, w, w, w, w, w, w, w}
	// combination 191/1024
	case !ok0 && ok1 && ok2 && ok3 && ok4 && ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
		}{w, w, w, w, w, w, w, w}
	// combination 192/1024
	case ok0 && ok1 && ok2 && ok3 && ok4 && ok5 && !ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
		}{w, w, w, w, w, w, w, w, w}
	// combination 193/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && !ok4 && !ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w}
	// combination 194/1024
	case ok0 && !ok1 && !ok2 && !ok3 && !ok4 && !ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w}
	// combination 195/1024
	case !ok0 && ok1 && !ok2 && !ok3 && !ok4 && !ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w}
	// combination 196/1024
	case ok0 && ok1 && !ok2 && !ok3 && !ok4 && !ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w}
	// combination 197/1024
	case !ok0 && !ok1 && ok2 && !ok3 && !ok4 && !ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w}
	// combination 198/1024
	case ok0 && !ok1 && ok2 && !ok3 && !ok4 && !ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w}
	// combination 199/1024
	case !ok0 && ok1 && ok2 && !ok3 && !ok4 && !ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w}
	// combination 200/1024
	case ok0 && ok1 && ok2 && !ok3 && !ok4 && !ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w}
	// combination 201/1024
	case !ok0 && !ok1 && !ok2 && ok3 && !ok4 && !ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w}
	// combination 202/1024
	case ok0 && !ok1 && !ok2 && ok3 && !ok4 && !ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w}
	// combination 203/1024
	case !ok0 && ok1 && !ok2 && ok3 && !ok4 && !ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w}
	// combination 204/1024
	case ok0 && ok1 && !ok2 && ok3 && !ok4 && !ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w}
	// combination 205/1024
	case !ok0 && !ok1 && ok2 && ok3 && !ok4 && !ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w}
	// combination 206/1024
	case ok0 && !ok1 && ok2 && ok3 && !ok4 && !ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w}
	// combination 207/1024
	case !ok0 && ok1 && ok2 && ok3 && !ok4 && !ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w}
	// combination 208/1024
	case ok0 && ok1 && ok2 && ok3 && !ok4 && !ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w, w}
	// combination 209/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && ok4 && !ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w}
	// combination 210/1024
	case ok0 && !ok1 && !ok2 && !ok3 && ok4 && !ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w}
	// combination 211/1024
	case !ok0 && ok1 && !ok2 && !ok3 && ok4 && !ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w}
	// combination 212/1024
	case ok0 && ok1 && !ok2 && !ok3 && ok4 && !ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w}
	// combination 213/1024
	case !ok0 && !ok1 && ok2 && !ok3 && ok4 && !ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w}
	// combination 214/1024
	case ok0 && !ok1 && ok2 && !ok3 && ok4 && !ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w}
	// combination 215/1024
	case !ok0 && ok1 && ok2 && !ok3 && ok4 && !ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w}
	// combination 216/1024
	case ok0 && ok1 && ok2 && !ok3 && ok4 && !ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w, w}
	// combination 217/1024
	case !ok0 && !ok1 && !ok2 && ok3 && ok4 && !ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w}
	// combination 218/1024
	case ok0 && !ok1 && !ok2 && ok3 && ok4 && !ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w}
	// combination 219/1024
	case !ok0 && ok1 && !ok2 && ok3 && ok4 && !ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w}
	// combination 220/1024
	case ok0 && ok1 && !ok2 && ok3 && ok4 && !ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w, w}
	// combination 221/1024
	case !ok0 && !ok1 && ok2 && ok3 && ok4 && !ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w}
	// combination 222/1024
	case ok0 && !ok1 && ok2 && ok3 && ok4 && !ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w, w}
	// combination 223/1024
	case !ok0 && ok1 && ok2 && ok3 && ok4 && !ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w, w}
	// combination 224/1024
	case ok0 && ok1 && ok2 && ok3 && ok4 && !ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w, w, w}
	// combination 225/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && !ok4 && ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w}
	// combination 226/1024
	case ok0 && !ok1 && !ok2 && !ok3 && !ok4 && ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w}
	// combination 227/1024
	case !ok0 && ok1 && !ok2 && !ok3 && !ok4 && ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w}
	// combination 228/1024
	case ok0 && ok1 && !ok2 && !ok3 && !ok4 && ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w}
	// combination 229/1024
	case !ok0 && !ok1 && ok2 && !ok3 && !ok4 && ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w}
	// combination 230/1024
	case ok0 && !ok1 && ok2 && !ok3 && !ok4 && ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w}
	// combination 231/1024
	case !ok0 && ok1 && ok2 && !ok3 && !ok4 && ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w}
	// combination 232/1024
	case ok0 && ok1 && ok2 && !ok3 && !ok4 && ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w, w}
	// combination 233/1024
	case !ok0 && !ok1 && !ok2 && ok3 && !ok4 && ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w}
	// combination 234/1024
	case ok0 && !ok1 && !ok2 && ok3 && !ok4 && ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w}
	// combination 235/1024
	case !ok0 && ok1 && !ok2 && ok3 && !ok4 && ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w}
	// combination 236/1024
	case ok0 && ok1 && !ok2 && ok3 && !ok4 && ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w, w}
	// combination 237/1024
	case !ok0 && !ok1 && ok2 && ok3 && !ok4 && ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w}
	// combination 238/1024
	case ok0 && !ok1 && ok2 && ok3 && !ok4 && ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w, w}
	// combination 239/1024
	case !ok0 && ok1 && ok2 && ok3 && !ok4 && ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w, w}
	// combination 240/1024
	case ok0 && ok1 && ok2 && ok3 && !ok4 && ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w, w, w}
	// combination 241/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && ok4 && ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w}
	// combination 242/1024
	case ok0 && !ok1 && !ok2 && !ok3 && ok4 && ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w}
	// combination 243/1024
	case !ok0 && ok1 && !ok2 && !ok3 && ok4 && ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w}
	// combination 244/1024
	case ok0 && ok1 && !ok2 && !ok3 && ok4 && ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w, w}
	// combination 245/1024
	case !ok0 && !ok1 && ok2 && !ok3 && ok4 && ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w}
	// combination 246/1024
	case ok0 && !ok1 && ok2 && !ok3 && ok4 && ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w, w}
	// combination 247/1024
	case !ok0 && ok1 && ok2 && !ok3 && ok4 && ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w, w}
	// combination 248/1024
	case ok0 && ok1 && ok2 && !ok3 && ok4 && ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w, w, w}
	// combination 249/1024
	case !ok0 && !ok1 && !ok2 && ok3 && ok4 && ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w}
	// combination 250/1024
	case ok0 && !ok1 && !ok2 && ok3 && ok4 && ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w, w}
	// combination 251/1024
	case !ok0 && ok1 && !ok2 && ok3 && ok4 && ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w, w}
	// combination 252/1024
	case ok0 && ok1 && !ok2 && ok3 && ok4 && ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w, w, w}
	// combination 253/1024
	case !ok0 && !ok1 && ok2 && ok3 && ok4 && ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w, w}
	// combination 254/1024
	case ok0 && !ok1 && ok2 && ok3 && ok4 && ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w, w, w}
	// combination 255/1024
	case !ok0 && ok1 && ok2 && ok3 && ok4 && ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w, w, w}
	// combination 256/1024
	case ok0 && ok1 && ok2 && ok3 && ok4 && ok5 && ok6 && ok7 && !ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 257/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && !ok4 && !ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.SessionResetter
		}{w, w, w}
	// combination 258/1024
	case ok0 && !ok1 && !ok2 && !ok3 && !ok4 && !ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.SessionResetter
		}{w, w, w, w}
	// combination 259/1024
	case !ok0 && ok1 && !ok2 && !ok3 && !ok4 && !ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.SessionResetter
		}{w, w, w, w}
	// combination 260/1024
	case ok0 && ok1 && !ok2 && !ok3 && !ok4 && !ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.SessionResetter
		}{w, w, w, w, w}
	// combination 261/1024
	case !ok0 && !ok1 && ok2 && !ok3 && !ok4 && !ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.SessionResetter
		}{w, w, w, w}
	// combination 262/1024
	case ok0 && !ok1 && ok2 && !ok3 && !ok4 && !ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.SessionResetter
		}{w, w, w, w, w}
	// combination 263/1024
	case !ok0 && ok1 && ok2 && !ok3 && !ok4 && !ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.SessionResetter
		}{w, w, w, w, w}
	// combination 264/1024
	case ok0 && ok1 && ok2 && !ok3 && !ok4 && !ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 265/1024
	case !ok0 && !ok1 && !ok2 && ok3 && !ok4 && !ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.SessionResetter
		}{w, w, w, w}
	// combination 266/1024
	case ok0 && !ok1 && !ok2 && ok3 && !ok4 && !ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.SessionResetter
		}{w, w, w, w, w}
	// combination 267/1024
	case !ok0 && ok1 && !ok2 && ok3 && !ok4 && !ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.SessionResetter
		}{w, w, w, w, w}
	// combination 268/1024
	case ok0 && ok1 && !ok2 && ok3 && !ok4 && !ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 269/1024
	case !ok0 && !ok1 && ok2 && ok3 && !ok4 && !ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.SessionResetter
		}{w, w, w, w, w}
	// combination 270/1024
	case ok0 && !ok1 && ok2 && ok3 && !ok4 && !ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 271/1024
	case !ok0 && ok1 && ok2 && ok3 && !ok4 && !ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 272/1024
	case ok0 && ok1 && ok2 && ok3 && !ok4 && !ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 273/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && ok4 && !ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.NamedValueChecker
			driver.SessionResetter
		}{w, w, w, w}
	// combination 274/1024
	case ok0 && !ok1 && !ok2 && !ok3 && ok4 && !ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.NamedValueChecker
			driver.SessionResetter
		}{w, w, w, w, w}
	// combination 275/1024
	case !ok0 && ok1 && !ok2 && !ok3 && ok4 && !ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.SessionResetter
		}{w, w, w, w, w}
	// combination 276/1024
	case ok0 && ok1 && !ok2 && !ok3 && ok4 && !ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 277/1024
	case !ok0 && !ok1 && ok2 && !ok3 && ok4 && !ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.NamedValueChecker
			driver.SessionResetter
		}{w, w, w, w, w}
	// combination 278/1024
	case ok0 && !ok1 && ok2 && !ok3 && ok4 && !ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.NamedValueChecker
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 279/1024
	case !ok0 && ok1 && ok2 && !ok3 && ok4 && !ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 280/1024
	case ok0 && ok1 && ok2 && !ok3 && ok4 && !ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 281/1024
	case !ok0 && !ok1 && !ok2 && ok3 && ok4 && !ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.NamedValueChecker
			driver.SessionResetter
		}{w, w, w, w, w}
	// combination 282/1024
	case ok0 && !ok1 && !ok2 && ok3 && ok4 && !ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.NamedValueChecker
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 283/1024
	case !ok0 && ok1 && !ok2 && ok3 && ok4 && !ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 284/1024
	case ok0 && ok1 && !ok2 && ok3 && ok4 && !ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 285/1024
	case !ok0 && !ok1 && ok2 && ok3 && ok4 && !ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 286/1024
	case ok0 && !ok1 && ok2 && ok3 && ok4 && !ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 287/1024
	case !ok0 && ok1 && ok2 && ok3 && ok4 && !ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 288/1024
	case ok0 && ok1 && ok2 && ok3 && ok4 && !ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 289/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && !ok4 && ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Pinger
			driver.SessionResetter
		}{w, w, w, w}
	// combination 290/1024
	case ok0 && !ok1 && !ok2 && !ok3 && !ok4 && ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Pinger
			driver.SessionResetter
		}{w, w, w, w, w}
	// combination 291/1024
	case !ok0 && ok1 && !ok2 && !ok3 && !ok4 && ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Pinger
			driver.SessionResetter
		}{w, w, w, w, w}
	// combination 292/1024
	case ok0 && ok1 && !ok2 && !ok3 && !ok4 && ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Pinger
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 293/1024
	case !ok0 && !ok1 && ok2 && !ok3 && !ok4 && ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.Pinger
			driver.SessionResetter
		}{w, w, w, w, w}
	// combination 294/1024
	case ok0 && !ok1 && ok2 && !ok3 && !ok4 && ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.Pinger
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 295/1024
	case !ok0 && ok1 && ok2 && !ok3 && !ok4 && ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.Pinger
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 296/1024
	case ok0 && ok1 && ok2 && !ok3 && !ok4 && ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.Pinger
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 297/1024
	case !ok0 && !ok1 && !ok2 && ok3 && !ok4 && ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.Pinger
			driver.SessionResetter
		}{w, w, w, w, w}
	// combination 298/1024
	case ok0 && !ok1 && !ok2 && ok3 && !ok4 && ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.Pinger
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 299/1024
	case !ok0 && ok1 && !ok2 && ok3 && !ok4 && ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Pinger
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 300/1024
	case ok0 && ok1 && !ok2 && ok3 && !ok4 && ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Pinger
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 301/1024
	case !ok0 && !ok1 && ok2 && ok3 && !ok4 && ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 302/1024
	case ok0 && !ok1 && ok2 && ok3 && !ok4 && ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 303/1024
	case !ok0 && ok1 && ok2 && ok3 && !ok4 && ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 304/1024
	case ok0 && ok1 && ok2 && ok3 && !ok4 && ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 305/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && ok4 && ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.NamedValueChecker
			driver.Pinger
			driver.SessionResetter
		}{w, w, w, w, w}
	// combination 306/1024
	case ok0 && !ok1 && !ok2 && !ok3 && ok4 && ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.NamedValueChecker
			driver.Pinger
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 307/1024
	case !ok0 && ok1 && !ok2 && !ok3 && ok4 && ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Pinger
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 308/1024
	case ok0 && ok1 && !ok2 && !ok3 && ok4 && ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Pinger
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 309/1024
	case !ok0 && !ok1 && ok2 && !ok3 && ok4 && ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 310/1024
	case ok0 && !ok1 && ok2 && !ok3 && ok4 && ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 311/1024
	case !ok0 && ok1 && ok2 && !ok3 && ok4 && ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 312/1024
	case ok0 && ok1 && ok2 && !ok3 && ok4 && ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 313/1024
	case !ok0 && !ok1 && !ok2 && ok3 && ok4 && ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 314/1024
	case ok0 && !ok1 && !ok2 && ok3 && ok4 && ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 315/1024
	case !ok0 && ok1 && !ok2 && ok3 && ok4 && ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 316/1024
	case ok0 && ok1 && !ok2 && ok3 && ok4 && ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 317/1024
	case !ok0 && !ok1 && ok2 && ok3 && ok4 && ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 318/1024
	case ok0 && !ok1 && ok2 && ok3 && ok4 && ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 319/1024
	case !ok0 && ok1 && ok2 && ok3 && ok4 && ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 320/1024
	case ok0 && ok1 && ok2 && ok3 && ok4 && ok5 && !ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w, w}
	// combination 321/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && !ok4 && !ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w}
	// combination 322/1024
	case ok0 && !ok1 && !ok2 && !ok3 && !ok4 && !ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w}
	// combination 323/1024
	case !ok0 && ok1 && !ok2 && !ok3 && !ok4 && !ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w}
	// combination 324/1024
	case ok0 && ok1 && !ok2 && !ok3 && !ok4 && !ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 325/1024
	case !ok0 && !ok1 && ok2 && !ok3 && !ok4 && !ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w}
	// combination 326/1024
	case ok0 && !ok1 && ok2 && !ok3 && !ok4 && !ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 327/1024
	case !ok0 && ok1 && ok2 && !ok3 && !ok4 && !ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 328/1024
	case ok0 && ok1 && ok2 && !ok3 && !ok4 && !ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 329/1024
	case !ok0 && !ok1 && !ok2 && ok3 && !ok4 && !ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w}
	// combination 330/1024
	case ok0 && !ok1 && !ok2 && ok3 && !ok4 && !ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 331/1024
	case !ok0 && ok1 && !ok2 && ok3 && !ok4 && !ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 332/1024
	case ok0 && ok1 && !ok2 && ok3 && !ok4 && !ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 333/1024
	case !ok0 && !ok1 && ok2 && ok3 && !ok4 && !ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 334/1024
	case ok0 && !ok1 && ok2 && ok3 && !ok4 && !ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 335/1024
	case !ok0 && ok1 && ok2 && ok3 && !ok4 && !ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 336/1024
	case ok0 && ok1 && ok2 && ok3 && !ok4 && !ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 337/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && ok4 && !ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.NamedValueChecker
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w}
	// combination 338/1024
	case ok0 && !ok1 && !ok2 && !ok3 && ok4 && !ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.NamedValueChecker
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 339/1024
	case !ok0 && ok1 && !ok2 && !ok3 && ok4 && !ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 340/1024
	case ok0 && ok1 && !ok2 && !ok3 && ok4 && !ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 341/1024
	case !ok0 && !ok1 && ok2 && !ok3 && ok4 && !ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.NamedValueChecker
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 342/1024
	case ok0 && !ok1 && ok2 && !ok3 && ok4 && !ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.NamedValueChecker
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 343/1024
	case !ok0 && ok1 && ok2 && !ok3 && ok4 && !ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 344/1024
	case ok0 && ok1 && ok2 && !ok3 && ok4 && !ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 345/1024
	case !ok0 && !ok1 && !ok2 && ok3 && ok4 && !ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 346/1024
	case ok0 && !ok1 && !ok2 && ok3 && ok4 && !ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 347/1024
	case !ok0 && ok1 && !ok2 && ok3 && ok4 && !ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 348/1024
	case ok0 && ok1 && !ok2 && ok3 && ok4 && !ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 349/1024
	case !ok0 && !ok1 && ok2 && ok3 && ok4 && !ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 350/1024
	case ok0 && !ok1 && ok2 && ok3 && ok4 && !ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 351/1024
	case !ok0 && ok1 && ok2 && ok3 && ok4 && !ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 352/1024
	case ok0 && ok1 && ok2 && ok3 && ok4 && !ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w, w}
	// combination 353/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && !ok4 && ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w}
	// combination 354/1024
	case ok0 && !ok1 && !ok2 && !ok3 && !ok4 && ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 355/1024
	case !ok0 && ok1 && !ok2 && !ok3 && !ok4 && ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 356/1024
	case ok0 && ok1 && !ok2 && !ok3 && !ok4 && ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 357/1024
	case !ok0 && !ok1 && ok2 && !ok3 && !ok4 && ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 358/1024
	case ok0 && !ok1 && ok2 && !ok3 && !ok4 && ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 359/1024
	case !ok0 && ok1 && ok2 && !ok3 && !ok4 && ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 360/1024
	case ok0 && ok1 && ok2 && !ok3 && !ok4 && ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 361/1024
	case !ok0 && !ok1 && !ok2 && ok3 && !ok4 && ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 362/1024
	case ok0 && !ok1 && !ok2 && ok3 && !ok4 && ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 363/1024
	case !ok0 && ok1 && !ok2 && ok3 && !ok4 && ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 364/1024
	case ok0 && ok1 && !ok2 && ok3 && !ok4 && ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 365/1024
	case !ok0 && !ok1 && ok2 && ok3 && !ok4 && ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 366/1024
	case ok0 && !ok1 && ok2 && ok3 && !ok4 && ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 367/1024
	case !ok0 && ok1 && ok2 && ok3 && !ok4 && ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 368/1024
	case ok0 && ok1 && ok2 && ok3 && !ok4 && ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w, w}
	// combination 369/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && ok4 && ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 370/1024
	case ok0 && !ok1 && !ok2 && !ok3 && ok4 && ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 371/1024
	case !ok0 && ok1 && !ok2 && !ok3 && ok4 && ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 372/1024
	case ok0 && ok1 && !ok2 && !ok3 && ok4 && ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 373/1024
	case !ok0 && !ok1 && ok2 && !ok3 && ok4 && ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 374/1024
	case ok0 && !ok1 && ok2 && !ok3 && ok4 && ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 375/1024
	case !ok0 && ok1 && ok2 && !ok3 && ok4 && ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 376/1024
	case ok0 && ok1 && ok2 && !ok3 && ok4 && ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w, w}
	// combination 377/1024
	case !ok0 && !ok1 && !ok2 && ok3 && ok4 && ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 378/1024
	case ok0 && !ok1 && !ok2 && ok3 && ok4 && ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 379/1024
	case !ok0 && ok1 && !ok2 && ok3 && ok4 && ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 380/1024
	case ok0 && ok1 && !ok2 && ok3 && ok4 && ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w, w}
	// combination 381/1024
	case !ok0 && !ok1 && ok2 && ok3 && ok4 && ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 382/1024
	case ok0 && !ok1 && ok2 && ok3 && ok4 && ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w, w}
	// combination 383/1024
	case !ok0 && ok1 && ok2 && ok3 && ok4 && ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w, w}
	// combination 384/1024
	case ok0 && ok1 && ok2 && ok3 && ok4 && ok5 && ok6 && !ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 385/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && !ok4 && !ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w}
	// combination 386/1024
	case ok0 && !ok1 && !ok2 && !ok3 && !ok4 && !ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w}
	// combination 387/1024
	case !ok0 && ok1 && !ok2 && !ok3 && !ok4 && !ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w}
	// combination 388/1024
	case ok0 && ok1 && !ok2 && !ok3 && !ok4 && !ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 389/1024
	case !ok0 && !ok1 && ok2 && !ok3 && !ok4 && !ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w}
	// combination 390/1024
	case ok0 && !ok1 && ok2 && !ok3 && !ok4 && !ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 391/1024
	case !ok0 && ok1 && ok2 && !ok3 && !ok4 && !ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 392/1024
	case ok0 && ok1 && ok2 && !ok3 && !ok4 && !ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 393/1024
	case !ok0 && !ok1 && !ok2 && ok3 && !ok4 && !ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w}
	// combination 394/1024
	case ok0 && !ok1 && !ok2 && ok3 && !ok4 && !ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 395/1024
	case !ok0 && ok1 && !ok2 && ok3 && !ok4 && !ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 396/1024
	case ok0 && ok1 && !ok2 && ok3 && !ok4 && !ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 397/1024
	case !ok0 && !ok1 && ok2 && ok3 && !ok4 && !ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 398/1024
	case ok0 && !ok1 && ok2 && ok3 && !ok4 && !ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 399/1024
	case !ok0 && ok1 && ok2 && ok3 && !ok4 && !ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 400/1024
	case ok0 && ok1 && ok2 && ok3 && !ok4 && !ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 401/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && ok4 && !ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.NamedValueChecker
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w}
	// combination 402/1024
	case ok0 && !ok1 && !ok2 && !ok3 && ok4 && !ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.NamedValueChecker
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 403/1024
	case !ok0 && ok1 && !ok2 && !ok3 && ok4 && !ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 404/1024
	case ok0 && ok1 && !ok2 && !ok3 && ok4 && !ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 405/1024
	case !ok0 && !ok1 && ok2 && !ok3 && ok4 && !ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.NamedValueChecker
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 406/1024
	case ok0 && !ok1 && ok2 && !ok3 && ok4 && !ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.NamedValueChecker
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 407/1024
	case !ok0 && ok1 && ok2 && !ok3 && ok4 && !ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 408/1024
	case ok0 && ok1 && ok2 && !ok3 && ok4 && !ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 409/1024
	case !ok0 && !ok1 && !ok2 && ok3 && ok4 && !ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.NamedValueChecker
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 410/1024
	case ok0 && !ok1 && !ok2 && ok3 && ok4 && !ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.NamedValueChecker
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 411/1024
	case !ok0 && ok1 && !ok2 && ok3 && ok4 && !ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 412/1024
	case ok0 && ok1 && !ok2 && ok3 && ok4 && !ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 413/1024
	case !ok0 && !ok1 && ok2 && ok3 && ok4 && !ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 414/1024
	case ok0 && !ok1 && ok2 && ok3 && ok4 && !ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 415/1024
	case !ok0 && ok1 && ok2 && ok3 && ok4 && !ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 416/1024
	case ok0 && ok1 && ok2 && ok3 && ok4 && !ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w, w}
	// combination 417/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && !ok4 && ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w}
	// combination 418/1024
	case ok0 && !ok1 && !ok2 && !ok3 && !ok4 && ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 419/1024
	case !ok0 && ok1 && !ok2 && !ok3 && !ok4 && ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 420/1024
	case ok0 && ok1 && !ok2 && !ok3 && !ok4 && ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 421/1024
	case !ok0 && !ok1 && ok2 && !ok3 && !ok4 && ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 422/1024
	case ok0 && !ok1 && ok2 && !ok3 && !ok4 && ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 423/1024
	case !ok0 && ok1 && ok2 && !ok3 && !ok4 && ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 424/1024
	case ok0 && ok1 && ok2 && !ok3 && !ok4 && ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 425/1024
	case !ok0 && !ok1 && !ok2 && ok3 && !ok4 && ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 426/1024
	case ok0 && !ok1 && !ok2 && ok3 && !ok4 && ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 427/1024
	case !ok0 && ok1 && !ok2 && ok3 && !ok4 && ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 428/1024
	case ok0 && ok1 && !ok2 && ok3 && !ok4 && ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 429/1024
	case !ok0 && !ok1 && ok2 && ok3 && !ok4 && ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 430/1024
	case ok0 && !ok1 && ok2 && ok3 && !ok4 && ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 431/1024
	case !ok0 && ok1 && ok2 && ok3 && !ok4 && ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 432/1024
	case ok0 && ok1 && ok2 && ok3 && !ok4 && ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w, w}
	// combination 433/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && ok4 && ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 434/1024
	case ok0 && !ok1 && !ok2 && !ok3 && ok4 && ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 435/1024
	case !ok0 && ok1 && !ok2 && !ok3 && ok4 && ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 436/1024
	case ok0 && ok1 && !ok2 && !ok3 && ok4 && ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 437/1024
	case !ok0 && !ok1 && ok2 && !ok3 && ok4 && ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 438/1024
	case ok0 && !ok1 && ok2 && !ok3 && ok4 && ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 439/1024
	case !ok0 && ok1 && ok2 && !ok3 && ok4 && ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 440/1024
	case ok0 && ok1 && ok2 && !ok3 && ok4 && ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w, w}
	// combination 441/1024
	case !ok0 && !ok1 && !ok2 && ok3 && ok4 && ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 442/1024
	case ok0 && !ok1 && !ok2 && ok3 && ok4 && ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 443/1024
	case !ok0 && ok1 && !ok2 && ok3 && ok4 && ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 444/1024
	case ok0 && ok1 && !ok2 && ok3 && ok4 && ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w, w}
	// combination 445/1024
	case !ok0 && !ok1 && ok2 && ok3 && ok4 && ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 446/1024
	case ok0 && !ok1 && ok2 && ok3 && ok4 && ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w, w}
	// combination 447/1024
	case !ok0 && ok1 && ok2 && ok3 && ok4 && ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w, w}
	// combination 448/1024
	case ok0 && ok1 && ok2 && ok3 && ok4 && ok5 && !ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 449/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && !ok4 && !ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w}
	// combination 450/1024
	case ok0 && !ok1 && !ok2 && !ok3 && !ok4 && !ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 451/1024
	case !ok0 && ok1 && !ok2 && !ok3 && !ok4 && !ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 452/1024
	case ok0 && ok1 && !ok2 && !ok3 && !ok4 && !ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 453/1024
	case !ok0 && !ok1 && ok2 && !ok3 && !ok4 && !ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 454/1024
	case ok0 && !ok1 && ok2 && !ok3 && !ok4 && !ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 455/1024
	case !ok0 && ok1 && ok2 && !ok3 && !ok4 && !ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 456/1024
	case ok0 && ok1 && ok2 && !ok3 && !ok4 && !ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 457/1024
	case !ok0 && !ok1 && !ok2 && ok3 && !ok4 && !ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 458/1024
	case ok0 && !ok1 && !ok2 && ok3 && !ok4 && !ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 459/1024
	case !ok0 && ok1 && !ok2 && ok3 && !ok4 && !ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 460/1024
	case ok0 && ok1 && !ok2 && ok3 && !ok4 && !ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 461/1024
	case !ok0 && !ok1 && ok2 && ok3 && !ok4 && !ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 462/1024
	case ok0 && !ok1 && ok2 && ok3 && !ok4 && !ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 463/1024
	case !ok0 && ok1 && ok2 && ok3 && !ok4 && !ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 464/1024
	case ok0 && ok1 && ok2 && ok3 && !ok4 && !ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w, w}
	// combination 465/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && ok4 && !ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 466/1024
	case ok0 && !ok1 && !ok2 && !ok3 && ok4 && !ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 467/1024
	case !ok0 && ok1 && !ok2 && !ok3 && ok4 && !ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 468/1024
	case ok0 && ok1 && !ok2 && !ok3 && ok4 && !ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 469/1024
	case !ok0 && !ok1 && ok2 && !ok3 && ok4 && !ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 470/1024
	case ok0 && !ok1 && ok2 && !ok3 && ok4 && !ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 471/1024
	case !ok0 && ok1 && ok2 && !ok3 && ok4 && !ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 472/1024
	case ok0 && ok1 && ok2 && !ok3 && ok4 && !ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w, w}
	// combination 473/1024
	case !ok0 && !ok1 && !ok2 && ok3 && ok4 && !ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 474/1024
	case ok0 && !ok1 && !ok2 && ok3 && ok4 && !ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 475/1024
	case !ok0 && ok1 && !ok2 && ok3 && ok4 && !ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 476/1024
	case ok0 && ok1 && !ok2 && ok3 && ok4 && !ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w, w}
	// combination 477/1024
	case !ok0 && !ok1 && ok2 && ok3 && ok4 && !ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 478/1024
	case ok0 && !ok1 && ok2 && ok3 && ok4 && !ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w, w}
	// combination 479/1024
	case !ok0 && ok1 && ok2 && ok3 && ok4 && !ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w, w}
	// combination 480/1024
	case ok0 && ok1 && ok2 && ok3 && ok4 && !ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 481/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && !ok4 && ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w}
	// combination 482/1024
	case ok0 && !ok1 && !ok2 && !ok3 && !ok4 && ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 483/1024
	case !ok0 && ok1 && !ok2 && !ok3 && !ok4 && ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 484/1024
	case ok0 && ok1 && !ok2 && !ok3 && !ok4 && ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 485/1024
	case !ok0 && !ok1 && ok2 && !ok3 && !ok4 && ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 486/1024
	case ok0 && !ok1 && ok2 && !ok3 && !ok4 && ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 487/1024
	case !ok0 && ok1 && ok2 && !ok3 && !ok4 && ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 488/1024
	case ok0 && ok1 && ok2 && !ok3 && !ok4 && ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w, w}
	// combination 489/1024
	case !ok0 && !ok1 && !ok2 && ok3 && !ok4 && ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 490/1024
	case ok0 && !ok1 && !ok2 && ok3 && !ok4 && ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 491/1024
	case !ok0 && ok1 && !ok2 && ok3 && !ok4 && ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 492/1024
	case ok0 && ok1 && !ok2 && ok3 && !ok4 && ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w, w}
	// combination 493/1024
	case !ok0 && !ok1 && ok2 && ok3 && !ok4 && ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 494/1024
	case ok0 && !ok1 && ok2 && ok3 && !ok4 && ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w, w}
	// combination 495/1024
	case !ok0 && ok1 && ok2 && ok3 && !ok4 && ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w, w}
	// combination 496/1024
	case ok0 && ok1 && ok2 && ok3 && !ok4 && ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 497/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && ok4 && ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w}
	// combination 498/1024
	case ok0 && !ok1 && !ok2 && !ok3 && ok4 && ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 499/1024
	case !ok0 && ok1 && !ok2 && !ok3 && ok4 && ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 500/1024
	case ok0 && ok1 && !ok2 && !ok3 && ok4 && ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w, w}
	// combination 501/1024
	case !ok0 && !ok1 && ok2 && !ok3 && ok4 && ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 502/1024
	case ok0 && !ok1 && ok2 && !ok3 && ok4 && ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w, w}
	// combination 503/1024
	case !ok0 && ok1 && ok2 && !ok3 && ok4 && ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w, w}
	// combination 504/1024
	case ok0 && ok1 && ok2 && !ok3 && ok4 && ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 505/1024
	case !ok0 && !ok1 && !ok2 && ok3 && ok4 && ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w}
	// combination 506/1024
	case ok0 && !ok1 && !ok2 && ok3 && ok4 && ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w, w}
	// combination 507/1024
	case !ok0 && ok1 && !ok2 && ok3 && ok4 && ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w, w}
	// combination 508/1024
	case ok0 && ok1 && !ok2 && ok3 && ok4 && ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 509/1024
	case !ok0 && !ok1 && ok2 && ok3 && ok4 && ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w, w}
	// combination 510/1024
	case ok0 && !ok1 && ok2 && ok3 && ok4 && ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 511/1024
	case !ok0 && ok1 && ok2 && ok3 && ok4 && ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 512/1024
	case ok0 && ok1 && ok2 && ok3 && ok4 && ok5 && ok6 && ok7 && ok8 && !ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
		}{w, w, w, w, w, w, w, w, w, w, w}
	// combination 513/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && !ok4 && !ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Validator
		}{w, w, w}
	// combination 514/1024
	case ok0 && !ok1 && !ok2 && !ok3 && !ok4 && !ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Validator
		}{w, w, w, w}
	// combination 515/1024
	case !ok0 && ok1 && !ok2 && !ok3 && !ok4 && !ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Validator
		}{w, w, w, w}
	// combination 516/1024
	case ok0 && ok1 && !ok2 && !ok3 && !ok4 && !ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Validator
		}{w, w, w, w, w}
	// combination 517/1024
	case !ok0 && !ok1 && ok2 && !ok3 && !ok4 && !ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.Validator
		}{w, w, w, w}
	// combination 518/1024
	case ok0 && !ok1 && ok2 && !ok3 && !ok4 && !ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.Validator
		}{w, w, w, w, w}
	// combination 519/1024
	case !ok0 && ok1 && ok2 && !ok3 && !ok4 && !ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.Validator
		}{w, w, w, w, w}
	// combination 520/1024
	case ok0 && ok1 && ok2 && !ok3 && !ok4 && !ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 521/1024
	case !ok0 && !ok1 && !ok2 && ok3 && !ok4 && !ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.Validator
		}{w, w, w, w}
	// combination 522/1024
	case ok0 && !ok1 && !ok2 && ok3 && !ok4 && !ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.Validator
		}{w, w, w, w, w}
	// combination 523/1024
	case !ok0 && ok1 && !ok2 && ok3 && !ok4 && !ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Validator
		}{w, w, w, w, w}
	// combination 524/1024
	case ok0 && ok1 && !ok2 && ok3 && !ok4 && !ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 525/1024
	case !ok0 && !ok1 && ok2 && ok3 && !ok4 && !ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.Validator
		}{w, w, w, w, w}
	// combination 526/1024
	case ok0 && !ok1 && ok2 && ok3 && !ok4 && !ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 527/1024
	case !ok0 && ok1 && ok2 && ok3 && !ok4 && !ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 528/1024
	case ok0 && ok1 && ok2 && ok3 && !ok4 && !ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 529/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && ok4 && !ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.NamedValueChecker
			driver.Validator
		}{w, w, w, w}
	// combination 530/1024
	case ok0 && !ok1 && !ok2 && !ok3 && ok4 && !ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.NamedValueChecker
			driver.Validator
		}{w, w, w, w, w}
	// combination 531/1024
	case !ok0 && ok1 && !ok2 && !ok3 && ok4 && !ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Validator
		}{w, w, w, w, w}
	// combination 532/1024
	case ok0 && ok1 && !ok2 && !ok3 && ok4 && !ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 533/1024
	case !ok0 && !ok1 && ok2 && !ok3 && ok4 && !ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.NamedValueChecker
			driver.Validator
		}{w, w, w, w, w}
	// combination 534/1024
	case ok0 && !ok1 && ok2 && !ok3 && ok4 && !ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.NamedValueChecker
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 535/1024
	case !ok0 && ok1 && ok2 && !ok3 && ok4 && !ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 536/1024
	case ok0 && ok1 && ok2 && !ok3 && ok4 && !ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 537/1024
	case !ok0 && !ok1 && !ok2 && ok3 && ok4 && !ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Validator
		}{w, w, w, w, w}
	// combination 538/1024
	case ok0 && !ok1 && !ok2 && ok3 && ok4 && !ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 539/1024
	case !ok0 && ok1 && !ok2 && ok3 && ok4 && !ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 540/1024
	case ok0 && ok1 && !ok2 && ok3 && ok4 && !ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 541/1024
	case !ok0 && !ok1 && ok2 && ok3 && ok4 && !ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 542/1024
	case ok0 && !ok1 && ok2 && ok3 && ok4 && !ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 543/1024
	case !ok0 && ok1 && ok2 && ok3 && ok4 && !ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 544/1024
	case ok0 && ok1 && ok2 && ok3 && ok4 && !ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 545/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && !ok4 && ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Pinger
			driver.Validator
		}{w, w, w, w}
	// combination 546/1024
	case ok0 && !ok1 && !ok2 && !ok3 && !ok4 && ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Pinger
			driver.Validator
		}{w, w, w, w, w}
	// combination 547/1024
	case !ok0 && ok1 && !ok2 && !ok3 && !ok4 && ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Pinger
			driver.Validator
		}{w, w, w, w, w}
	// combination 548/1024
	case ok0 && ok1 && !ok2 && !ok3 && !ok4 && ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Pinger
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 549/1024
	case !ok0 && !ok1 && ok2 && !ok3 && !ok4 && ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.Pinger
			driver.Validator
		}{w, w, w, w, w}
	// combination 550/1024
	case ok0 && !ok1 && ok2 && !ok3 && !ok4 && ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.Pinger
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 551/1024
	case !ok0 && ok1 && ok2 && !ok3 && !ok4 && ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.Pinger
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 552/1024
	case ok0 && ok1 && ok2 && !ok3 && !ok4 && ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.Pinger
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 553/1024
	case !ok0 && !ok1 && !ok2 && ok3 && !ok4 && ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.Pinger
			driver.Validator
		}{w, w, w, w, w}
	// combination 554/1024
	case ok0 && !ok1 && !ok2 && ok3 && !ok4 && ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.Pinger
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 555/1024
	case !ok0 && ok1 && !ok2 && ok3 && !ok4 && ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Pinger
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 556/1024
	case ok0 && ok1 && !ok2 && ok3 && !ok4 && ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Pinger
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 557/1024
	case !ok0 && !ok1 && ok2 && ok3 && !ok4 && ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 558/1024
	case ok0 && !ok1 && ok2 && ok3 && !ok4 && ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 559/1024
	case !ok0 && ok1 && ok2 && ok3 && !ok4 && ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 560/1024
	case ok0 && ok1 && ok2 && ok3 && !ok4 && ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 561/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && ok4 && ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.NamedValueChecker
			driver.Pinger
			driver.Validator
		}{w, w, w, w, w}
	// combination 562/1024
	case ok0 && !ok1 && !ok2 && !ok3 && ok4 && ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.NamedValueChecker
			driver.Pinger
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 563/1024
	case !ok0 && ok1 && !ok2 && !ok3 && ok4 && ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 564/1024
	case ok0 && ok1 && !ok2 && !ok3 && ok4 && ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 565/1024
	case !ok0 && !ok1 && ok2 && !ok3 && ok4 && ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 566/1024
	case ok0 && !ok1 && ok2 && !ok3 && ok4 && ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 567/1024
	case !ok0 && ok1 && ok2 && !ok3 && ok4 && ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 568/1024
	case ok0 && ok1 && ok2 && !ok3 && ok4 && ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 569/1024
	case !ok0 && !ok1 && !ok2 && ok3 && ok4 && ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 570/1024
	case ok0 && !ok1 && !ok2 && ok3 && ok4 && ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 571/1024
	case !ok0 && ok1 && !ok2 && ok3 && ok4 && ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 572/1024
	case ok0 && ok1 && !ok2 && ok3 && ok4 && ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 573/1024
	case !ok0 && !ok1 && ok2 && ok3 && ok4 && ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 574/1024
	case ok0 && !ok1 && ok2 && ok3 && ok4 && ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 575/1024
	case !ok0 && ok1 && ok2 && ok3 && ok4 && ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 576/1024
	case ok0 && ok1 && ok2 && ok3 && ok4 && ok5 && !ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 577/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && !ok4 && !ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Queryer
			driver.Validator
		}{w, w, w, w}
	// combination 578/1024
	case ok0 && !ok1 && !ok2 && !ok3 && !ok4 && !ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w}
	// combination 579/1024
	case !ok0 && ok1 && !ok2 && !ok3 && !ok4 && !ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w}
	// combination 580/1024
	case ok0 && ok1 && !ok2 && !ok3 && !ok4 && !ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 581/1024
	case !ok0 && !ok1 && ok2 && !ok3 && !ok4 && !ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w}
	// combination 582/1024
	case ok0 && !ok1 && ok2 && !ok3 && !ok4 && !ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 583/1024
	case !ok0 && ok1 && ok2 && !ok3 && !ok4 && !ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 584/1024
	case ok0 && ok1 && ok2 && !ok3 && !ok4 && !ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 585/1024
	case !ok0 && !ok1 && !ok2 && ok3 && !ok4 && !ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w}
	// combination 586/1024
	case ok0 && !ok1 && !ok2 && ok3 && !ok4 && !ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 587/1024
	case !ok0 && ok1 && !ok2 && ok3 && !ok4 && !ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 588/1024
	case ok0 && ok1 && !ok2 && ok3 && !ok4 && !ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 589/1024
	case !ok0 && !ok1 && ok2 && ok3 && !ok4 && !ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 590/1024
	case ok0 && !ok1 && ok2 && ok3 && !ok4 && !ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 591/1024
	case !ok0 && ok1 && ok2 && ok3 && !ok4 && !ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 592/1024
	case ok0 && ok1 && ok2 && ok3 && !ok4 && !ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 593/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && ok4 && !ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.NamedValueChecker
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w}
	// combination 594/1024
	case ok0 && !ok1 && !ok2 && !ok3 && ok4 && !ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.NamedValueChecker
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 595/1024
	case !ok0 && ok1 && !ok2 && !ok3 && ok4 && !ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 596/1024
	case ok0 && ok1 && !ok2 && !ok3 && ok4 && !ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 597/1024
	case !ok0 && !ok1 && ok2 && !ok3 && ok4 && !ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.NamedValueChecker
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 598/1024
	case ok0 && !ok1 && ok2 && !ok3 && ok4 && !ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.NamedValueChecker
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 599/1024
	case !ok0 && ok1 && ok2 && !ok3 && ok4 && !ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 600/1024
	case ok0 && ok1 && ok2 && !ok3 && ok4 && !ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 601/1024
	case !ok0 && !ok1 && !ok2 && ok3 && ok4 && !ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 602/1024
	case ok0 && !ok1 && !ok2 && ok3 && ok4 && !ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 603/1024
	case !ok0 && ok1 && !ok2 && ok3 && ok4 && !ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 604/1024
	case ok0 && ok1 && !ok2 && ok3 && ok4 && !ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 605/1024
	case !ok0 && !ok1 && ok2 && ok3 && ok4 && !ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 606/1024
	case ok0 && !ok1 && ok2 && ok3 && ok4 && !ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 607/1024
	case !ok0 && ok1 && ok2 && ok3 && ok4 && !ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 608/1024
	case ok0 && ok1 && ok2 && ok3 && ok4 && !ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 609/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && !ok4 && ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Pinger
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w}
	// combination 610/1024
	case ok0 && !ok1 && !ok2 && !ok3 && !ok4 && ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Pinger
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 611/1024
	case !ok0 && ok1 && !ok2 && !ok3 && !ok4 && ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Pinger
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 612/1024
	case ok0 && ok1 && !ok2 && !ok3 && !ok4 && ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Pinger
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 613/1024
	case !ok0 && !ok1 && ok2 && !ok3 && !ok4 && ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.Pinger
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 614/1024
	case ok0 && !ok1 && ok2 && !ok3 && !ok4 && ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.Pinger
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 615/1024
	case !ok0 && ok1 && ok2 && !ok3 && !ok4 && ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.Pinger
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 616/1024
	case ok0 && ok1 && ok2 && !ok3 && !ok4 && ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.Pinger
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 617/1024
	case !ok0 && !ok1 && !ok2 && ok3 && !ok4 && ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 618/1024
	case ok0 && !ok1 && !ok2 && ok3 && !ok4 && ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 619/1024
	case !ok0 && ok1 && !ok2 && ok3 && !ok4 && ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 620/1024
	case ok0 && ok1 && !ok2 && ok3 && !ok4 && ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 621/1024
	case !ok0 && !ok1 && ok2 && ok3 && !ok4 && ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 622/1024
	case ok0 && !ok1 && ok2 && ok3 && !ok4 && ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 623/1024
	case !ok0 && ok1 && ok2 && ok3 && !ok4 && ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 624/1024
	case ok0 && ok1 && ok2 && ok3 && !ok4 && ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 625/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && ok4 && ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 626/1024
	case ok0 && !ok1 && !ok2 && !ok3 && ok4 && ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 627/1024
	case !ok0 && ok1 && !ok2 && !ok3 && ok4 && ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 628/1024
	case ok0 && ok1 && !ok2 && !ok3 && ok4 && ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 629/1024
	case !ok0 && !ok1 && ok2 && !ok3 && ok4 && ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 630/1024
	case ok0 && !ok1 && ok2 && !ok3 && ok4 && ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 631/1024
	case !ok0 && ok1 && ok2 && !ok3 && ok4 && ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 632/1024
	case ok0 && ok1 && ok2 && !ok3 && ok4 && ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 633/1024
	case !ok0 && !ok1 && !ok2 && ok3 && ok4 && ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 634/1024
	case ok0 && !ok1 && !ok2 && ok3 && ok4 && ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 635/1024
	case !ok0 && ok1 && !ok2 && ok3 && ok4 && ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 636/1024
	case ok0 && ok1 && !ok2 && ok3 && ok4 && ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 637/1024
	case !ok0 && !ok1 && ok2 && ok3 && ok4 && ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 638/1024
	case ok0 && !ok1 && ok2 && ok3 && ok4 && ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 639/1024
	case !ok0 && ok1 && ok2 && ok3 && ok4 && ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 640/1024
	case ok0 && ok1 && ok2 && ok3 && ok4 && ok5 && ok6 && !ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 641/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && !ok4 && !ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w}
	// combination 642/1024
	case ok0 && !ok1 && !ok2 && !ok3 && !ok4 && !ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w}
	// combination 643/1024
	case !ok0 && ok1 && !ok2 && !ok3 && !ok4 && !ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w}
	// combination 644/1024
	case ok0 && ok1 && !ok2 && !ok3 && !ok4 && !ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 645/1024
	case !ok0 && !ok1 && ok2 && !ok3 && !ok4 && !ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w}
	// combination 646/1024
	case ok0 && !ok1 && ok2 && !ok3 && !ok4 && !ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 647/1024
	case !ok0 && ok1 && ok2 && !ok3 && !ok4 && !ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 648/1024
	case ok0 && ok1 && ok2 && !ok3 && !ok4 && !ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 649/1024
	case !ok0 && !ok1 && !ok2 && ok3 && !ok4 && !ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w}
	// combination 650/1024
	case ok0 && !ok1 && !ok2 && ok3 && !ok4 && !ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 651/1024
	case !ok0 && ok1 && !ok2 && ok3 && !ok4 && !ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 652/1024
	case ok0 && ok1 && !ok2 && ok3 && !ok4 && !ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 653/1024
	case !ok0 && !ok1 && ok2 && ok3 && !ok4 && !ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 654/1024
	case ok0 && !ok1 && ok2 && ok3 && !ok4 && !ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 655/1024
	case !ok0 && ok1 && ok2 && ok3 && !ok4 && !ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 656/1024
	case ok0 && ok1 && ok2 && ok3 && !ok4 && !ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 657/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && ok4 && !ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.NamedValueChecker
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w}
	// combination 658/1024
	case ok0 && !ok1 && !ok2 && !ok3 && ok4 && !ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.NamedValueChecker
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 659/1024
	case !ok0 && ok1 && !ok2 && !ok3 && ok4 && !ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 660/1024
	case ok0 && ok1 && !ok2 && !ok3 && ok4 && !ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 661/1024
	case !ok0 && !ok1 && ok2 && !ok3 && ok4 && !ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.NamedValueChecker
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 662/1024
	case ok0 && !ok1 && ok2 && !ok3 && ok4 && !ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.NamedValueChecker
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 663/1024
	case !ok0 && ok1 && ok2 && !ok3 && ok4 && !ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 664/1024
	case ok0 && ok1 && ok2 && !ok3 && ok4 && !ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 665/1024
	case !ok0 && !ok1 && !ok2 && ok3 && ok4 && !ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.NamedValueChecker
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 666/1024
	case ok0 && !ok1 && !ok2 && ok3 && ok4 && !ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.NamedValueChecker
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 667/1024
	case !ok0 && ok1 && !ok2 && ok3 && ok4 && !ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 668/1024
	case ok0 && ok1 && !ok2 && ok3 && ok4 && !ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 669/1024
	case !ok0 && !ok1 && ok2 && ok3 && ok4 && !ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 670/1024
	case ok0 && !ok1 && ok2 && ok3 && ok4 && !ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 671/1024
	case !ok0 && ok1 && ok2 && ok3 && ok4 && !ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 672/1024
	case ok0 && ok1 && ok2 && ok3 && ok4 && !ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 673/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && !ok4 && ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Pinger
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w}
	// combination 674/1024
	case ok0 && !ok1 && !ok2 && !ok3 && !ok4 && ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Pinger
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 675/1024
	case !ok0 && ok1 && !ok2 && !ok3 && !ok4 && ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Pinger
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 676/1024
	case ok0 && ok1 && !ok2 && !ok3 && !ok4 && ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Pinger
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 677/1024
	case !ok0 && !ok1 && ok2 && !ok3 && !ok4 && ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.Pinger
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 678/1024
	case ok0 && !ok1 && ok2 && !ok3 && !ok4 && ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.Pinger
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 679/1024
	case !ok0 && ok1 && ok2 && !ok3 && !ok4 && ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.Pinger
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 680/1024
	case ok0 && ok1 && ok2 && !ok3 && !ok4 && ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.Pinger
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 681/1024
	case !ok0 && !ok1 && !ok2 && ok3 && !ok4 && ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.Pinger
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 682/1024
	case ok0 && !ok1 && !ok2 && ok3 && !ok4 && ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.Pinger
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 683/1024
	case !ok0 && ok1 && !ok2 && ok3 && !ok4 && ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Pinger
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 684/1024
	case ok0 && ok1 && !ok2 && ok3 && !ok4 && ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Pinger
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 685/1024
	case !ok0 && !ok1 && ok2 && ok3 && !ok4 && ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 686/1024
	case ok0 && !ok1 && ok2 && ok3 && !ok4 && ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 687/1024
	case !ok0 && ok1 && ok2 && ok3 && !ok4 && ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 688/1024
	case ok0 && ok1 && ok2 && ok3 && !ok4 && ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 689/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && ok4 && ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 690/1024
	case ok0 && !ok1 && !ok2 && !ok3 && ok4 && ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 691/1024
	case !ok0 && ok1 && !ok2 && !ok3 && ok4 && ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 692/1024
	case ok0 && ok1 && !ok2 && !ok3 && ok4 && ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 693/1024
	case !ok0 && !ok1 && ok2 && !ok3 && ok4 && ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 694/1024
	case ok0 && !ok1 && ok2 && !ok3 && ok4 && ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 695/1024
	case !ok0 && ok1 && ok2 && !ok3 && ok4 && ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 696/1024
	case ok0 && ok1 && ok2 && !ok3 && ok4 && ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 697/1024
	case !ok0 && !ok1 && !ok2 && ok3 && ok4 && ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 698/1024
	case ok0 && !ok1 && !ok2 && ok3 && ok4 && ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 699/1024
	case !ok0 && ok1 && !ok2 && ok3 && ok4 && ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 700/1024
	case ok0 && ok1 && !ok2 && ok3 && ok4 && ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 701/1024
	case !ok0 && !ok1 && ok2 && ok3 && ok4 && ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 702/1024
	case ok0 && !ok1 && ok2 && ok3 && ok4 && ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 703/1024
	case !ok0 && ok1 && ok2 && ok3 && ok4 && ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 704/1024
	case ok0 && ok1 && ok2 && ok3 && ok4 && ok5 && !ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 705/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && !ok4 && !ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w}
	// combination 706/1024
	case ok0 && !ok1 && !ok2 && !ok3 && !ok4 && !ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 707/1024
	case !ok0 && ok1 && !ok2 && !ok3 && !ok4 && !ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 708/1024
	case ok0 && ok1 && !ok2 && !ok3 && !ok4 && !ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 709/1024
	case !ok0 && !ok1 && ok2 && !ok3 && !ok4 && !ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 710/1024
	case ok0 && !ok1 && ok2 && !ok3 && !ok4 && !ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 711/1024
	case !ok0 && ok1 && ok2 && !ok3 && !ok4 && !ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 712/1024
	case ok0 && ok1 && ok2 && !ok3 && !ok4 && !ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 713/1024
	case !ok0 && !ok1 && !ok2 && ok3 && !ok4 && !ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 714/1024
	case ok0 && !ok1 && !ok2 && ok3 && !ok4 && !ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 715/1024
	case !ok0 && ok1 && !ok2 && ok3 && !ok4 && !ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 716/1024
	case ok0 && ok1 && !ok2 && ok3 && !ok4 && !ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 717/1024
	case !ok0 && !ok1 && ok2 && ok3 && !ok4 && !ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 718/1024
	case ok0 && !ok1 && ok2 && ok3 && !ok4 && !ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 719/1024
	case !ok0 && ok1 && ok2 && ok3 && !ok4 && !ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 720/1024
	case ok0 && ok1 && ok2 && ok3 && !ok4 && !ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 721/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && ok4 && !ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 722/1024
	case ok0 && !ok1 && !ok2 && !ok3 && ok4 && !ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 723/1024
	case !ok0 && ok1 && !ok2 && !ok3 && ok4 && !ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 724/1024
	case ok0 && ok1 && !ok2 && !ok3 && ok4 && !ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 725/1024
	case !ok0 && !ok1 && ok2 && !ok3 && ok4 && !ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 726/1024
	case ok0 && !ok1 && ok2 && !ok3 && ok4 && !ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 727/1024
	case !ok0 && ok1 && ok2 && !ok3 && ok4 && !ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 728/1024
	case ok0 && ok1 && ok2 && !ok3 && ok4 && !ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 729/1024
	case !ok0 && !ok1 && !ok2 && ok3 && ok4 && !ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 730/1024
	case ok0 && !ok1 && !ok2 && ok3 && ok4 && !ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 731/1024
	case !ok0 && ok1 && !ok2 && ok3 && ok4 && !ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 732/1024
	case ok0 && ok1 && !ok2 && ok3 && ok4 && !ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 733/1024
	case !ok0 && !ok1 && ok2 && ok3 && ok4 && !ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 734/1024
	case ok0 && !ok1 && ok2 && ok3 && ok4 && !ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 735/1024
	case !ok0 && ok1 && ok2 && ok3 && ok4 && !ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 736/1024
	case ok0 && ok1 && ok2 && ok3 && ok4 && !ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 737/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && !ok4 && ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 738/1024
	case ok0 && !ok1 && !ok2 && !ok3 && !ok4 && ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 739/1024
	case !ok0 && ok1 && !ok2 && !ok3 && !ok4 && ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 740/1024
	case ok0 && ok1 && !ok2 && !ok3 && !ok4 && ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 741/1024
	case !ok0 && !ok1 && ok2 && !ok3 && !ok4 && ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 742/1024
	case ok0 && !ok1 && ok2 && !ok3 && !ok4 && ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 743/1024
	case !ok0 && ok1 && ok2 && !ok3 && !ok4 && ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 744/1024
	case ok0 && ok1 && ok2 && !ok3 && !ok4 && ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 745/1024
	case !ok0 && !ok1 && !ok2 && ok3 && !ok4 && ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 746/1024
	case ok0 && !ok1 && !ok2 && ok3 && !ok4 && ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 747/1024
	case !ok0 && ok1 && !ok2 && ok3 && !ok4 && ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 748/1024
	case ok0 && ok1 && !ok2 && ok3 && !ok4 && ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 749/1024
	case !ok0 && !ok1 && ok2 && ok3 && !ok4 && ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 750/1024
	case ok0 && !ok1 && ok2 && ok3 && !ok4 && ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 751/1024
	case !ok0 && ok1 && ok2 && ok3 && !ok4 && ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 752/1024
	case ok0 && ok1 && ok2 && ok3 && !ok4 && ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 753/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && ok4 && ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 754/1024
	case ok0 && !ok1 && !ok2 && !ok3 && ok4 && ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 755/1024
	case !ok0 && ok1 && !ok2 && !ok3 && ok4 && ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 756/1024
	case ok0 && ok1 && !ok2 && !ok3 && ok4 && ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 757/1024
	case !ok0 && !ok1 && ok2 && !ok3 && ok4 && ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 758/1024
	case ok0 && !ok1 && ok2 && !ok3 && ok4 && ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 759/1024
	case !ok0 && ok1 && ok2 && !ok3 && ok4 && ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 760/1024
	case ok0 && ok1 && ok2 && !ok3 && ok4 && ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 761/1024
	case !ok0 && !ok1 && !ok2 && ok3 && ok4 && ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 762/1024
	case ok0 && !ok1 && !ok2 && ok3 && ok4 && ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 763/1024
	case !ok0 && ok1 && !ok2 && ok3 && ok4 && ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 764/1024
	case ok0 && ok1 && !ok2 && ok3 && ok4 && ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 765/1024
	case !ok0 && !ok1 && ok2 && ok3 && ok4 && ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 766/1024
	case ok0 && !ok1 && ok2 && ok3 && ok4 && ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 767/1024
	case !ok0 && ok1 && ok2 && ok3 && ok4 && ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 768/1024
	case ok0 && ok1 && ok2 && ok3 && ok4 && ok5 && ok6 && ok7 && !ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w, w}
	// combination 769/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && !ok4 && !ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w}
	// combination 770/1024
	case ok0 && !ok1 && !ok2 && !ok3 && !ok4 && !ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w}
	// combination 771/1024
	case !ok0 && ok1 && !ok2 && !ok3 && !ok4 && !ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w}
	// combination 772/1024
	case ok0 && ok1 && !ok2 && !ok3 && !ok4 && !ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 773/1024
	case !ok0 && !ok1 && ok2 && !ok3 && !ok4 && !ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w}
	// combination 774/1024
	case ok0 && !ok1 && ok2 && !ok3 && !ok4 && !ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 775/1024
	case !ok0 && ok1 && ok2 && !ok3 && !ok4 && !ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 776/1024
	case ok0 && ok1 && ok2 && !ok3 && !ok4 && !ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 777/1024
	case !ok0 && !ok1 && !ok2 && ok3 && !ok4 && !ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w}
	// combination 778/1024
	case ok0 && !ok1 && !ok2 && ok3 && !ok4 && !ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 779/1024
	case !ok0 && ok1 && !ok2 && ok3 && !ok4 && !ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 780/1024
	case ok0 && ok1 && !ok2 && ok3 && !ok4 && !ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 781/1024
	case !ok0 && !ok1 && ok2 && ok3 && !ok4 && !ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 782/1024
	case ok0 && !ok1 && ok2 && ok3 && !ok4 && !ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 783/1024
	case !ok0 && ok1 && ok2 && ok3 && !ok4 && !ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 784/1024
	case ok0 && ok1 && ok2 && ok3 && !ok4 && !ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 785/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && ok4 && !ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.NamedValueChecker
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w}
	// combination 786/1024
	case ok0 && !ok1 && !ok2 && !ok3 && ok4 && !ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.NamedValueChecker
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 787/1024
	case !ok0 && ok1 && !ok2 && !ok3 && ok4 && !ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 788/1024
	case ok0 && ok1 && !ok2 && !ok3 && ok4 && !ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 789/1024
	case !ok0 && !ok1 && ok2 && !ok3 && ok4 && !ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.NamedValueChecker
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 790/1024
	case ok0 && !ok1 && ok2 && !ok3 && ok4 && !ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.NamedValueChecker
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 791/1024
	case !ok0 && ok1 && ok2 && !ok3 && ok4 && !ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 792/1024
	case ok0 && ok1 && ok2 && !ok3 && ok4 && !ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 793/1024
	case !ok0 && !ok1 && !ok2 && ok3 && ok4 && !ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.NamedValueChecker
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 794/1024
	case ok0 && !ok1 && !ok2 && ok3 && ok4 && !ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.NamedValueChecker
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 795/1024
	case !ok0 && ok1 && !ok2 && ok3 && ok4 && !ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 796/1024
	case ok0 && ok1 && !ok2 && ok3 && ok4 && !ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 797/1024
	case !ok0 && !ok1 && ok2 && ok3 && ok4 && !ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 798/1024
	case ok0 && !ok1 && ok2 && ok3 && ok4 && !ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 799/1024
	case !ok0 && ok1 && ok2 && ok3 && ok4 && !ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 800/1024
	case ok0 && ok1 && ok2 && ok3 && ok4 && !ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 801/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && !ok4 && ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Pinger
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w}
	// combination 802/1024
	case ok0 && !ok1 && !ok2 && !ok3 && !ok4 && ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Pinger
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 803/1024
	case !ok0 && ok1 && !ok2 && !ok3 && !ok4 && ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Pinger
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 804/1024
	case ok0 && ok1 && !ok2 && !ok3 && !ok4 && ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Pinger
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 805/1024
	case !ok0 && !ok1 && ok2 && !ok3 && !ok4 && ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.Pinger
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 806/1024
	case ok0 && !ok1 && ok2 && !ok3 && !ok4 && ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.Pinger
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 807/1024
	case !ok0 && ok1 && ok2 && !ok3 && !ok4 && ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.Pinger
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 808/1024
	case ok0 && ok1 && ok2 && !ok3 && !ok4 && ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.Pinger
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 809/1024
	case !ok0 && !ok1 && !ok2 && ok3 && !ok4 && ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.Pinger
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 810/1024
	case ok0 && !ok1 && !ok2 && ok3 && !ok4 && ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.Pinger
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 811/1024
	case !ok0 && ok1 && !ok2 && ok3 && !ok4 && ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Pinger
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 812/1024
	case ok0 && ok1 && !ok2 && ok3 && !ok4 && ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Pinger
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 813/1024
	case !ok0 && !ok1 && ok2 && ok3 && !ok4 && ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 814/1024
	case ok0 && !ok1 && ok2 && ok3 && !ok4 && ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 815/1024
	case !ok0 && ok1 && ok2 && ok3 && !ok4 && ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 816/1024
	case ok0 && ok1 && ok2 && ok3 && !ok4 && ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 817/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && ok4 && ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.NamedValueChecker
			driver.Pinger
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 818/1024
	case ok0 && !ok1 && !ok2 && !ok3 && ok4 && ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.NamedValueChecker
			driver.Pinger
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 819/1024
	case !ok0 && ok1 && !ok2 && !ok3 && ok4 && ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Pinger
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 820/1024
	case ok0 && ok1 && !ok2 && !ok3 && ok4 && ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Pinger
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 821/1024
	case !ok0 && !ok1 && ok2 && !ok3 && ok4 && ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 822/1024
	case ok0 && !ok1 && ok2 && !ok3 && ok4 && ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 823/1024
	case !ok0 && ok1 && ok2 && !ok3 && ok4 && ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 824/1024
	case ok0 && ok1 && ok2 && !ok3 && ok4 && ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 825/1024
	case !ok0 && !ok1 && !ok2 && ok3 && ok4 && ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 826/1024
	case ok0 && !ok1 && !ok2 && ok3 && ok4 && ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 827/1024
	case !ok0 && ok1 && !ok2 && ok3 && ok4 && ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 828/1024
	case ok0 && ok1 && !ok2 && ok3 && ok4 && ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 829/1024
	case !ok0 && !ok1 && ok2 && ok3 && ok4 && ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 830/1024
	case ok0 && !ok1 && ok2 && ok3 && ok4 && ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 831/1024
	case !ok0 && ok1 && ok2 && ok3 && ok4 && ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 832/1024
	case ok0 && ok1 && ok2 && ok3 && ok4 && ok5 && !ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 833/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && !ok4 && !ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w}
	// combination 834/1024
	case ok0 && !ok1 && !ok2 && !ok3 && !ok4 && !ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 835/1024
	case !ok0 && ok1 && !ok2 && !ok3 && !ok4 && !ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 836/1024
	case ok0 && ok1 && !ok2 && !ok3 && !ok4 && !ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 837/1024
	case !ok0 && !ok1 && ok2 && !ok3 && !ok4 && !ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 838/1024
	case ok0 && !ok1 && ok2 && !ok3 && !ok4 && !ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 839/1024
	case !ok0 && ok1 && ok2 && !ok3 && !ok4 && !ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 840/1024
	case ok0 && ok1 && ok2 && !ok3 && !ok4 && !ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 841/1024
	case !ok0 && !ok1 && !ok2 && ok3 && !ok4 && !ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 842/1024
	case ok0 && !ok1 && !ok2 && ok3 && !ok4 && !ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 843/1024
	case !ok0 && ok1 && !ok2 && ok3 && !ok4 && !ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 844/1024
	case ok0 && ok1 && !ok2 && ok3 && !ok4 && !ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 845/1024
	case !ok0 && !ok1 && ok2 && ok3 && !ok4 && !ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 846/1024
	case ok0 && !ok1 && ok2 && ok3 && !ok4 && !ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 847/1024
	case !ok0 && ok1 && ok2 && ok3 && !ok4 && !ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 848/1024
	case ok0 && ok1 && ok2 && ok3 && !ok4 && !ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 849/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && ok4 && !ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.NamedValueChecker
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 850/1024
	case ok0 && !ok1 && !ok2 && !ok3 && ok4 && !ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.NamedValueChecker
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 851/1024
	case !ok0 && ok1 && !ok2 && !ok3 && ok4 && !ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 852/1024
	case ok0 && ok1 && !ok2 && !ok3 && ok4 && !ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 853/1024
	case !ok0 && !ok1 && ok2 && !ok3 && ok4 && !ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.NamedValueChecker
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 854/1024
	case ok0 && !ok1 && ok2 && !ok3 && ok4 && !ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.NamedValueChecker
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 855/1024
	case !ok0 && ok1 && ok2 && !ok3 && ok4 && !ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 856/1024
	case ok0 && ok1 && ok2 && !ok3 && ok4 && !ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 857/1024
	case !ok0 && !ok1 && !ok2 && ok3 && ok4 && !ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 858/1024
	case ok0 && !ok1 && !ok2 && ok3 && ok4 && !ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 859/1024
	case !ok0 && ok1 && !ok2 && ok3 && ok4 && !ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 860/1024
	case ok0 && ok1 && !ok2 && ok3 && ok4 && !ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 861/1024
	case !ok0 && !ok1 && ok2 && ok3 && ok4 && !ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 862/1024
	case ok0 && !ok1 && ok2 && ok3 && ok4 && !ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 863/1024
	case !ok0 && ok1 && ok2 && ok3 && ok4 && !ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 864/1024
	case ok0 && ok1 && ok2 && ok3 && ok4 && !ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 865/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && !ok4 && ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 866/1024
	case ok0 && !ok1 && !ok2 && !ok3 && !ok4 && ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 867/1024
	case !ok0 && ok1 && !ok2 && !ok3 && !ok4 && ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 868/1024
	case ok0 && ok1 && !ok2 && !ok3 && !ok4 && ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 869/1024
	case !ok0 && !ok1 && ok2 && !ok3 && !ok4 && ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 870/1024
	case ok0 && !ok1 && ok2 && !ok3 && !ok4 && ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 871/1024
	case !ok0 && ok1 && ok2 && !ok3 && !ok4 && ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 872/1024
	case ok0 && ok1 && ok2 && !ok3 && !ok4 && ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 873/1024
	case !ok0 && !ok1 && !ok2 && ok3 && !ok4 && ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 874/1024
	case ok0 && !ok1 && !ok2 && ok3 && !ok4 && ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 875/1024
	case !ok0 && ok1 && !ok2 && ok3 && !ok4 && ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 876/1024
	case ok0 && ok1 && !ok2 && ok3 && !ok4 && ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 877/1024
	case !ok0 && !ok1 && ok2 && ok3 && !ok4 && ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 878/1024
	case ok0 && !ok1 && ok2 && ok3 && !ok4 && ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 879/1024
	case !ok0 && ok1 && ok2 && ok3 && !ok4 && ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 880/1024
	case ok0 && ok1 && ok2 && ok3 && !ok4 && ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 881/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && ok4 && ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 882/1024
	case ok0 && !ok1 && !ok2 && !ok3 && ok4 && ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 883/1024
	case !ok0 && ok1 && !ok2 && !ok3 && ok4 && ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 884/1024
	case ok0 && ok1 && !ok2 && !ok3 && ok4 && ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 885/1024
	case !ok0 && !ok1 && ok2 && !ok3 && ok4 && ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 886/1024
	case ok0 && !ok1 && ok2 && !ok3 && ok4 && ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 887/1024
	case !ok0 && ok1 && ok2 && !ok3 && ok4 && ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 888/1024
	case ok0 && ok1 && ok2 && !ok3 && ok4 && ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 889/1024
	case !ok0 && !ok1 && !ok2 && ok3 && ok4 && ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 890/1024
	case ok0 && !ok1 && !ok2 && ok3 && ok4 && ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 891/1024
	case !ok0 && ok1 && !ok2 && ok3 && ok4 && ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 892/1024
	case ok0 && ok1 && !ok2 && ok3 && ok4 && ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 893/1024
	case !ok0 && !ok1 && ok2 && ok3 && ok4 && ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 894/1024
	case ok0 && !ok1 && ok2 && ok3 && ok4 && ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 895/1024
	case !ok0 && ok1 && ok2 && ok3 && ok4 && ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 896/1024
	case ok0 && ok1 && ok2 && ok3 && ok4 && ok5 && ok6 && !ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w, w}
	// combination 897/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && !ok4 && !ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w}
	// combination 898/1024
	case ok0 && !ok1 && !ok2 && !ok3 && !ok4 && !ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 899/1024
	case !ok0 && ok1 && !ok2 && !ok3 && !ok4 && !ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 900/1024
	case ok0 && ok1 && !ok2 && !ok3 && !ok4 && !ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 901/1024
	case !ok0 && !ok1 && ok2 && !ok3 && !ok4 && !ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 902/1024
	case ok0 && !ok1 && ok2 && !ok3 && !ok4 && !ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 903/1024
	case !ok0 && ok1 && ok2 && !ok3 && !ok4 && !ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 904/1024
	case ok0 && ok1 && ok2 && !ok3 && !ok4 && !ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 905/1024
	case !ok0 && !ok1 && !ok2 && ok3 && !ok4 && !ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 906/1024
	case ok0 && !ok1 && !ok2 && ok3 && !ok4 && !ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 907/1024
	case !ok0 && ok1 && !ok2 && ok3 && !ok4 && !ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 908/1024
	case ok0 && ok1 && !ok2 && ok3 && !ok4 && !ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 909/1024
	case !ok0 && !ok1 && ok2 && ok3 && !ok4 && !ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 910/1024
	case ok0 && !ok1 && ok2 && ok3 && !ok4 && !ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 911/1024
	case !ok0 && ok1 && ok2 && ok3 && !ok4 && !ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 912/1024
	case ok0 && ok1 && ok2 && ok3 && !ok4 && !ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 913/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && ok4 && !ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.NamedValueChecker
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 914/1024
	case ok0 && !ok1 && !ok2 && !ok3 && ok4 && !ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.NamedValueChecker
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 915/1024
	case !ok0 && ok1 && !ok2 && !ok3 && ok4 && !ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 916/1024
	case ok0 && ok1 && !ok2 && !ok3 && ok4 && !ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 917/1024
	case !ok0 && !ok1 && ok2 && !ok3 && ok4 && !ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.NamedValueChecker
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 918/1024
	case ok0 && !ok1 && ok2 && !ok3 && ok4 && !ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.NamedValueChecker
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 919/1024
	case !ok0 && ok1 && ok2 && !ok3 && ok4 && !ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 920/1024
	case ok0 && ok1 && ok2 && !ok3 && ok4 && !ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 921/1024
	case !ok0 && !ok1 && !ok2 && ok3 && ok4 && !ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.NamedValueChecker
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 922/1024
	case ok0 && !ok1 && !ok2 && ok3 && ok4 && !ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.NamedValueChecker
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 923/1024
	case !ok0 && ok1 && !ok2 && ok3 && ok4 && !ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 924/1024
	case ok0 && ok1 && !ok2 && ok3 && ok4 && !ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 925/1024
	case !ok0 && !ok1 && ok2 && ok3 && ok4 && !ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 926/1024
	case ok0 && !ok1 && ok2 && ok3 && ok4 && !ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 927/1024
	case !ok0 && ok1 && ok2 && ok3 && ok4 && !ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 928/1024
	case ok0 && ok1 && ok2 && ok3 && ok4 && !ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 929/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && !ok4 && ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 930/1024
	case ok0 && !ok1 && !ok2 && !ok3 && !ok4 && ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 931/1024
	case !ok0 && ok1 && !ok2 && !ok3 && !ok4 && ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 932/1024
	case ok0 && ok1 && !ok2 && !ok3 && !ok4 && ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 933/1024
	case !ok0 && !ok1 && ok2 && !ok3 && !ok4 && ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 934/1024
	case ok0 && !ok1 && ok2 && !ok3 && !ok4 && ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 935/1024
	case !ok0 && ok1 && ok2 && !ok3 && !ok4 && ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 936/1024
	case ok0 && ok1 && ok2 && !ok3 && !ok4 && ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 937/1024
	case !ok0 && !ok1 && !ok2 && ok3 && !ok4 && ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 938/1024
	case ok0 && !ok1 && !ok2 && ok3 && !ok4 && ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 939/1024
	case !ok0 && ok1 && !ok2 && ok3 && !ok4 && ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 940/1024
	case ok0 && ok1 && !ok2 && ok3 && !ok4 && ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 941/1024
	case !ok0 && !ok1 && ok2 && ok3 && !ok4 && ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 942/1024
	case ok0 && !ok1 && ok2 && ok3 && !ok4 && ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 943/1024
	case !ok0 && ok1 && ok2 && ok3 && !ok4 && ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 944/1024
	case ok0 && ok1 && ok2 && ok3 && !ok4 && ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 945/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && ok4 && ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 946/1024
	case ok0 && !ok1 && !ok2 && !ok3 && ok4 && ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 947/1024
	case !ok0 && ok1 && !ok2 && !ok3 && ok4 && ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 948/1024
	case ok0 && ok1 && !ok2 && !ok3 && ok4 && ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 949/1024
	case !ok0 && !ok1 && ok2 && !ok3 && ok4 && ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 950/1024
	case ok0 && !ok1 && ok2 && !ok3 && ok4 && ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 951/1024
	case !ok0 && ok1 && ok2 && !ok3 && ok4 && ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 952/1024
	case ok0 && ok1 && ok2 && !ok3 && ok4 && ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 953/1024
	case !ok0 && !ok1 && !ok2 && ok3 && ok4 && ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 954/1024
	case ok0 && !ok1 && !ok2 && ok3 && ok4 && ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 955/1024
	case !ok0 && ok1 && !ok2 && ok3 && ok4 && ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 956/1024
	case ok0 && ok1 && !ok2 && ok3 && ok4 && ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 957/1024
	case !ok0 && !ok1 && ok2 && ok3 && ok4 && ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 958/1024
	case ok0 && !ok1 && ok2 && ok3 && ok4 && ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 959/1024
	case !ok0 && ok1 && ok2 && ok3 && ok4 && ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 960/1024
	case ok0 && ok1 && ok2 && ok3 && ok4 && ok5 && !ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w, w}
	// combination 961/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && !ok4 && !ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w}
	// combination 962/1024
	case ok0 && !ok1 && !ok2 && !ok3 && !ok4 && !ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 963/1024
	case !ok0 && ok1 && !ok2 && !ok3 && !ok4 && !ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 964/1024
	case ok0 && ok1 && !ok2 && !ok3 && !ok4 && !ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 965/1024
	case !ok0 && !ok1 && ok2 && !ok3 && !ok4 && !ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 966/1024
	case ok0 && !ok1 && ok2 && !ok3 && !ok4 && !ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 967/1024
	case !ok0 && ok1 && ok2 && !ok3 && !ok4 && !ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 968/1024
	case ok0 && ok1 && ok2 && !ok3 && !ok4 && !ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 969/1024
	case !ok0 && !ok1 && !ok2 && ok3 && !ok4 && !ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 970/1024
	case ok0 && !ok1 && !ok2 && ok3 && !ok4 && !ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 971/1024
	case !ok0 && ok1 && !ok2 && ok3 && !ok4 && !ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 972/1024
	case ok0 && ok1 && !ok2 && ok3 && !ok4 && !ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 973/1024
	case !ok0 && !ok1 && ok2 && ok3 && !ok4 && !ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 974/1024
	case ok0 && !ok1 && ok2 && ok3 && !ok4 && !ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 975/1024
	case !ok0 && ok1 && ok2 && ok3 && !ok4 && !ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 976/1024
	case ok0 && ok1 && ok2 && ok3 && !ok4 && !ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 977/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && ok4 && !ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 978/1024
	case ok0 && !ok1 && !ok2 && !ok3 && ok4 && !ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 979/1024
	case !ok0 && ok1 && !ok2 && !ok3 && ok4 && !ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 980/1024
	case ok0 && ok1 && !ok2 && !ok3 && ok4 && !ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 981/1024
	case !ok0 && !ok1 && ok2 && !ok3 && ok4 && !ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 982/1024
	case ok0 && !ok1 && ok2 && !ok3 && ok4 && !ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 983/1024
	case !ok0 && ok1 && ok2 && !ok3 && ok4 && !ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 984/1024
	case ok0 && ok1 && ok2 && !ok3 && ok4 && !ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 985/1024
	case !ok0 && !ok1 && !ok2 && ok3 && ok4 && !ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 986/1024
	case ok0 && !ok1 && !ok2 && ok3 && ok4 && !ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 987/1024
	case !ok0 && ok1 && !ok2 && ok3 && ok4 && !ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 988/1024
	case ok0 && ok1 && !ok2 && ok3 && ok4 && !ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 989/1024
	case !ok0 && !ok1 && ok2 && ok3 && ok4 && !ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 990/1024
	case ok0 && !ok1 && ok2 && ok3 && ok4 && !ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 991/1024
	case !ok0 && ok1 && ok2 && ok3 && ok4 && !ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 992/1024
	case ok0 && ok1 && ok2 && ok3 && ok4 && !ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w, w}
	// combination 993/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && !ok4 && ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w}
	// combination 994/1024
	case ok0 && !ok1 && !ok2 && !ok3 && !ok4 && ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 995/1024
	case !ok0 && ok1 && !ok2 && !ok3 && !ok4 && ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 996/1024
	case ok0 && ok1 && !ok2 && !ok3 && !ok4 && ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 997/1024
	case !ok0 && !ok1 && ok2 && !ok3 && !ok4 && ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 998/1024
	case ok0 && !ok1 && ok2 && !ok3 && !ok4 && ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 999/1024
	case !ok0 && ok1 && ok2 && !ok3 && !ok4 && ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 1000/1024
	case ok0 && ok1 && ok2 && !ok3 && !ok4 && ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 1001/1024
	case !ok0 && !ok1 && !ok2 && ok3 && !ok4 && ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 1002/1024
	case ok0 && !ok1 && !ok2 && ok3 && !ok4 && ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 1003/1024
	case !ok0 && ok1 && !ok2 && ok3 && !ok4 && ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 1004/1024
	case ok0 && ok1 && !ok2 && ok3 && !ok4 && ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 1005/1024
	case !ok0 && !ok1 && ok2 && ok3 && !ok4 && ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 1006/1024
	case ok0 && !ok1 && ok2 && ok3 && !ok4 && ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 1007/1024
	case !ok0 && ok1 && ok2 && ok3 && !ok4 && ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 1008/1024
	case ok0 && ok1 && ok2 && ok3 && !ok4 && ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w, w}
	// combination 1009/1024
	case !ok0 && !ok1 && !ok2 && !ok3 && ok4 && ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w}
	// combination 1010/1024
	case ok0 && !ok1 && !ok2 && !ok3 && ok4 && ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 1011/1024
	case !ok0 && ok1 && !ok2 && !ok3 && ok4 && ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 1012/1024
	case ok0 && ok1 && !ok2 && !ok3 && ok4 && ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 1013/1024
	case !ok0 && !ok1 && ok2 && !ok3 && ok4 && ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 1014/1024
	case ok0 && !ok1 && ok2 && !ok3 && ok4 && ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 1015/1024
	case !ok0 && ok1 && ok2 && !ok3 && ok4 && ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 1016/1024
	case ok0 && ok1 && ok2 && !ok3 && ok4 && ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w, w}
	// combination 1017/1024
	case !ok0 && !ok1 && !ok2 && ok3 && ok4 && ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w}
	// combination 1018/1024
	case ok0 && !ok1 && !ok2 && ok3 && ok4 && ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 1019/1024
	case !ok0 && ok1 && !ok2 && ok3 && ok4 && ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 1020/1024
	case ok0 && ok1 && !ok2 && ok3 && ok4 && ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w, w}
	// combination 1021/1024
	case !ok0 && !ok1 && ok2 && ok3 && ok4 && ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w}
	// combination 1022/1024
	case ok0 && !ok1 && ok2 && ok3 && ok4 && ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w, w}
	// combination 1023/1024
	case !ok0 && ok1 && ok2 && ok3 && ok4 && ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w, w}
	// combination 1024/1024
	case ok0 && ok1 && ok2 && ok3 && ok4 && ok5 && ok6 && ok7 && ok8 && ok9:
		return struct {
			ConnUnwrapper
			driver.Conn
			driver.ConnBeginTx
			driver.ConnPrepareContext
			driver.Execer
			driver.ExecerContext
			driver.NamedValueChecker
			driver.Pinger
			driver.Queryer
			driver.QueryerContext
			driver.SessionResetter
			driver.Validator
		}{w, w, w, w, w, w, w, w, w, w, w, w}
	}
	panic("unreachable")
}
