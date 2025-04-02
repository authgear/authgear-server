package databasesqlwrapper

import (
	context "context"
	driver "database/sql/driver"
)

type StmtUnwrapper interface {
	UnwrapStmt() driver.Stmt
}

func UnwrapStmt(w driver.Stmt) driver.Stmt {
	if ww, ok := w.(StmtUnwrapper); ok {
		return UnwrapStmt(ww.UnwrapStmt())
	} else {
		return w
	}
}

type Stmt_Close func() error
type Stmt_Exec func([]driver.Value) (driver.Result, error)
type Stmt_NumInput func() int
type Stmt_Query func([]driver.Value) (driver.Rows, error)
type Stmt_ColumnConverter func(int) driver.ValueConverter
type Stmt_CheckNamedValue func(*driver.NamedValue) error
type Stmt_ExecContext func(context.Context, []driver.NamedValue) (driver.Result, error)
type Stmt_QueryContext func(context.Context, []driver.NamedValue) (driver.Rows, error)

type StmtInterceptor struct {
	Close           func(Stmt_Close) Stmt_Close
	Exec            func(Stmt_Exec) Stmt_Exec
	NumInput        func(Stmt_NumInput) Stmt_NumInput
	Query           func(Stmt_Query) Stmt_Query
	ColumnConverter func(Stmt_ColumnConverter) Stmt_ColumnConverter
	CheckNamedValue func(Stmt_CheckNamedValue) Stmt_CheckNamedValue
	ExecContext     func(Stmt_ExecContext) Stmt_ExecContext
	QueryContext    func(Stmt_QueryContext) Stmt_QueryContext
}

type StmtWrapper struct {
	wrapped     driver.Stmt
	interceptor StmtInterceptor
}

func (w *StmtWrapper) UnwrapStmt() driver.Stmt {
	return w.wrapped
}

func (w *StmtWrapper) Close() error {
	f := w.wrapped.(driver.Stmt).Close
	if w.interceptor.Close != nil {
		f = w.interceptor.Close(f)
	}
	return f()
}

func (w *StmtWrapper) Exec(a0 []driver.Value) (driver.Result, error) {
	f := w.wrapped.(driver.Stmt).Exec
	if w.interceptor.Exec != nil {
		f = w.interceptor.Exec(f)
	}
	return f(a0)
}

func (w *StmtWrapper) NumInput() int {
	f := w.wrapped.(driver.Stmt).NumInput
	if w.interceptor.NumInput != nil {
		f = w.interceptor.NumInput(f)
	}
	return f()
}

func (w *StmtWrapper) Query(a0 []driver.Value) (driver.Rows, error) {
	f := w.wrapped.(driver.Stmt).Query
	if w.interceptor.Query != nil {
		f = w.interceptor.Query(f)
	}
	return f(a0)
}

func (w *StmtWrapper) ColumnConverter(a0 int) driver.ValueConverter {
	f := w.wrapped.(driver.ColumnConverter).ColumnConverter
	if w.interceptor.ColumnConverter != nil {
		f = w.interceptor.ColumnConverter(f)
	}
	return f(a0)
}

func (w *StmtWrapper) CheckNamedValue(a0 *driver.NamedValue) error {
	f := w.wrapped.(driver.NamedValueChecker).CheckNamedValue
	if w.interceptor.CheckNamedValue != nil {
		f = w.interceptor.CheckNamedValue(f)
	}
	return f(a0)
}

func (w *StmtWrapper) ExecContext(a0 context.Context, a1 []driver.NamedValue) (driver.Result, error) {
	f := w.wrapped.(driver.StmtExecContext).ExecContext
	if w.interceptor.ExecContext != nil {
		f = w.interceptor.ExecContext(f)
	}
	return f(a0, a1)
}

func (w *StmtWrapper) QueryContext(a0 context.Context, a1 []driver.NamedValue) (driver.Rows, error) {
	f := w.wrapped.(driver.StmtQueryContext).QueryContext
	if w.interceptor.QueryContext != nil {
		f = w.interceptor.QueryContext(f)
	}
	return f(a0, a1)
}

func WrapStmt(wrapped driver.Stmt, interceptor StmtInterceptor) driver.Stmt {
	w := &StmtWrapper{wrapped: wrapped, interceptor: interceptor}
	_, ok0 := wrapped.(driver.ColumnConverter)
	_, ok1 := wrapped.(driver.NamedValueChecker)
	_, ok2 := wrapped.(driver.StmtExecContext)
	_, ok3 := wrapped.(driver.StmtQueryContext)
	switch {
	// combination 1/16
	case !ok0 && !ok1 && !ok2 && !ok3:
		return struct {
			StmtUnwrapper
			driver.Stmt
		}{w, w}
	// combination 2/16
	case ok0 && !ok1 && !ok2 && !ok3:
		type combination_1_0 interface {
			ColumnConverter(int) driver.ValueConverter
		}
		return struct {
			StmtUnwrapper
			driver.Stmt
			combination_1_0
		}{w, w, w}
	// combination 3/16
	case !ok0 && ok1 && !ok2 && !ok3:
		return struct {
			StmtUnwrapper
			driver.Stmt
			driver.NamedValueChecker
		}{w, w, w}
	// combination 4/16
	case ok0 && ok1 && !ok2 && !ok3:
		type combination_3_0 interface {
			ColumnConverter(int) driver.ValueConverter
		}
		return struct {
			StmtUnwrapper
			driver.Stmt
			combination_3_0
			driver.NamedValueChecker
		}{w, w, w, w}
	// combination 5/16
	case !ok0 && !ok1 && ok2 && !ok3:
		return struct {
			StmtUnwrapper
			driver.Stmt
			driver.StmtExecContext
		}{w, w, w}
	// combination 6/16
	case ok0 && !ok1 && ok2 && !ok3:
		type combination_5_0 interface {
			ColumnConverter(int) driver.ValueConverter
		}
		return struct {
			StmtUnwrapper
			driver.Stmt
			combination_5_0
			driver.StmtExecContext
		}{w, w, w, w}
	// combination 7/16
	case !ok0 && ok1 && ok2 && !ok3:
		return struct {
			StmtUnwrapper
			driver.Stmt
			driver.NamedValueChecker
			driver.StmtExecContext
		}{w, w, w, w}
	// combination 8/16
	case ok0 && ok1 && ok2 && !ok3:
		type combination_7_0 interface {
			ColumnConverter(int) driver.ValueConverter
		}
		return struct {
			StmtUnwrapper
			driver.Stmt
			combination_7_0
			driver.NamedValueChecker
			driver.StmtExecContext
		}{w, w, w, w, w}
	// combination 9/16
	case !ok0 && !ok1 && !ok2 && ok3:
		return struct {
			StmtUnwrapper
			driver.Stmt
			driver.StmtQueryContext
		}{w, w, w}
	// combination 10/16
	case ok0 && !ok1 && !ok2 && ok3:
		type combination_9_0 interface {
			ColumnConverter(int) driver.ValueConverter
		}
		return struct {
			StmtUnwrapper
			driver.Stmt
			combination_9_0
			driver.StmtQueryContext
		}{w, w, w, w}
	// combination 11/16
	case !ok0 && ok1 && !ok2 && ok3:
		return struct {
			StmtUnwrapper
			driver.Stmt
			driver.NamedValueChecker
			driver.StmtQueryContext
		}{w, w, w, w}
	// combination 12/16
	case ok0 && ok1 && !ok2 && ok3:
		type combination_11_0 interface {
			ColumnConverter(int) driver.ValueConverter
		}
		return struct {
			StmtUnwrapper
			driver.Stmt
			combination_11_0
			driver.NamedValueChecker
			driver.StmtQueryContext
		}{w, w, w, w, w}
	// combination 13/16
	case !ok0 && !ok1 && ok2 && ok3:
		return struct {
			StmtUnwrapper
			driver.Stmt
			driver.StmtExecContext
			driver.StmtQueryContext
		}{w, w, w, w}
	// combination 14/16
	case ok0 && !ok1 && ok2 && ok3:
		type combination_13_0 interface {
			ColumnConverter(int) driver.ValueConverter
		}
		return struct {
			StmtUnwrapper
			driver.Stmt
			combination_13_0
			driver.StmtExecContext
			driver.StmtQueryContext
		}{w, w, w, w, w}
	// combination 15/16
	case !ok0 && ok1 && ok2 && ok3:
		return struct {
			StmtUnwrapper
			driver.Stmt
			driver.NamedValueChecker
			driver.StmtExecContext
			driver.StmtQueryContext
		}{w, w, w, w, w}
	// combination 16/16
	case ok0 && ok1 && ok2 && ok3:
		type combination_15_0 interface {
			ColumnConverter(int) driver.ValueConverter
		}
		return struct {
			StmtUnwrapper
			driver.Stmt
			combination_15_0
			driver.NamedValueChecker
			driver.StmtExecContext
			driver.StmtQueryContext
		}{w, w, w, w, w, w}
	}
	panic("unreachable")
}
