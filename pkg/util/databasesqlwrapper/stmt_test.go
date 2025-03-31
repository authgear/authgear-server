package databasesqlwrapper

import (
	driver "database/sql/driver"
	testing "testing"
)

func TestWrapStmt(t *testing.T) {
	{
		t.Log("combination 1/16: driver.Stmt")
		wrapped := struct {
			driver.Stmt
		}{}
		w := WrapStmt(wrapped, StmtInterceptor{})

		if _, ok := w.(driver.ColumnConverter); ok != false {
			t.Errorf("combination 1/16: unexpected interface driver.ColumnConverter")
		}
		if _, ok := w.(driver.NamedValueChecker); ok != false {
			t.Errorf("combination 1/16: unexpected interface driver.NamedValueChecker")
		}
		if _, ok := w.(driver.StmtExecContext); ok != false {
			t.Errorf("combination 1/16: unexpected interface driver.StmtExecContext")
		}
		if _, ok := w.(driver.StmtQueryContext); ok != false {
			t.Errorf("combination 1/16: unexpected interface driver.StmtQueryContext")
		}

		if w, ok := w.(StmtUnwrapper); ok {
			if w.UnwrapStmt() != wrapped {
				t.Errorf("combination 1/16: UnwrapStmt() failed")
			}
		} else {
			t.Errorf("combination 1/16: StmtUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 2/16: driver.Stmt driver.ColumnConverter")
		type combination_1_0 interface {
			ColumnConverter(int) driver.ValueConverter
		}
		wrapped := struct {
			driver.Stmt
			combination_1_0
		}{}
		w := WrapStmt(wrapped, StmtInterceptor{})

		if _, ok := w.(driver.ColumnConverter); ok != true {
			t.Errorf("combination 2/16: unexpected interface driver.ColumnConverter")
		}
		if _, ok := w.(driver.NamedValueChecker); ok != false {
			t.Errorf("combination 2/16: unexpected interface driver.NamedValueChecker")
		}
		if _, ok := w.(driver.StmtExecContext); ok != false {
			t.Errorf("combination 2/16: unexpected interface driver.StmtExecContext")
		}
		if _, ok := w.(driver.StmtQueryContext); ok != false {
			t.Errorf("combination 2/16: unexpected interface driver.StmtQueryContext")
		}

		if w, ok := w.(StmtUnwrapper); ok {
			if w.UnwrapStmt() != wrapped {
				t.Errorf("combination 2/16: UnwrapStmt() failed")
			}
		} else {
			t.Errorf("combination 2/16: StmtUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 3/16: driver.Stmt driver.NamedValueChecker")
		wrapped := struct {
			driver.Stmt
			driver.NamedValueChecker
		}{}
		w := WrapStmt(wrapped, StmtInterceptor{})

		if _, ok := w.(driver.ColumnConverter); ok != false {
			t.Errorf("combination 3/16: unexpected interface driver.ColumnConverter")
		}
		if _, ok := w.(driver.NamedValueChecker); ok != true {
			t.Errorf("combination 3/16: unexpected interface driver.NamedValueChecker")
		}
		if _, ok := w.(driver.StmtExecContext); ok != false {
			t.Errorf("combination 3/16: unexpected interface driver.StmtExecContext")
		}
		if _, ok := w.(driver.StmtQueryContext); ok != false {
			t.Errorf("combination 3/16: unexpected interface driver.StmtQueryContext")
		}

		if w, ok := w.(StmtUnwrapper); ok {
			if w.UnwrapStmt() != wrapped {
				t.Errorf("combination 3/16: UnwrapStmt() failed")
			}
		} else {
			t.Errorf("combination 3/16: StmtUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 4/16: driver.Stmt driver.ColumnConverter driver.NamedValueChecker")
		type combination_3_0 interface {
			ColumnConverter(int) driver.ValueConverter
		}
		wrapped := struct {
			driver.Stmt
			combination_3_0
			driver.NamedValueChecker
		}{}
		w := WrapStmt(wrapped, StmtInterceptor{})

		if _, ok := w.(driver.ColumnConverter); ok != true {
			t.Errorf("combination 4/16: unexpected interface driver.ColumnConverter")
		}
		if _, ok := w.(driver.NamedValueChecker); ok != true {
			t.Errorf("combination 4/16: unexpected interface driver.NamedValueChecker")
		}
		if _, ok := w.(driver.StmtExecContext); ok != false {
			t.Errorf("combination 4/16: unexpected interface driver.StmtExecContext")
		}
		if _, ok := w.(driver.StmtQueryContext); ok != false {
			t.Errorf("combination 4/16: unexpected interface driver.StmtQueryContext")
		}

		if w, ok := w.(StmtUnwrapper); ok {
			if w.UnwrapStmt() != wrapped {
				t.Errorf("combination 4/16: UnwrapStmt() failed")
			}
		} else {
			t.Errorf("combination 4/16: StmtUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 5/16: driver.Stmt driver.StmtExecContext")
		wrapped := struct {
			driver.Stmt
			driver.StmtExecContext
		}{}
		w := WrapStmt(wrapped, StmtInterceptor{})

		if _, ok := w.(driver.ColumnConverter); ok != false {
			t.Errorf("combination 5/16: unexpected interface driver.ColumnConverter")
		}
		if _, ok := w.(driver.NamedValueChecker); ok != false {
			t.Errorf("combination 5/16: unexpected interface driver.NamedValueChecker")
		}
		if _, ok := w.(driver.StmtExecContext); ok != true {
			t.Errorf("combination 5/16: unexpected interface driver.StmtExecContext")
		}
		if _, ok := w.(driver.StmtQueryContext); ok != false {
			t.Errorf("combination 5/16: unexpected interface driver.StmtQueryContext")
		}

		if w, ok := w.(StmtUnwrapper); ok {
			if w.UnwrapStmt() != wrapped {
				t.Errorf("combination 5/16: UnwrapStmt() failed")
			}
		} else {
			t.Errorf("combination 5/16: StmtUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 6/16: driver.Stmt driver.ColumnConverter driver.StmtExecContext")
		type combination_5_0 interface {
			ColumnConverter(int) driver.ValueConverter
		}
		wrapped := struct {
			driver.Stmt
			combination_5_0
			driver.StmtExecContext
		}{}
		w := WrapStmt(wrapped, StmtInterceptor{})

		if _, ok := w.(driver.ColumnConverter); ok != true {
			t.Errorf("combination 6/16: unexpected interface driver.ColumnConverter")
		}
		if _, ok := w.(driver.NamedValueChecker); ok != false {
			t.Errorf("combination 6/16: unexpected interface driver.NamedValueChecker")
		}
		if _, ok := w.(driver.StmtExecContext); ok != true {
			t.Errorf("combination 6/16: unexpected interface driver.StmtExecContext")
		}
		if _, ok := w.(driver.StmtQueryContext); ok != false {
			t.Errorf("combination 6/16: unexpected interface driver.StmtQueryContext")
		}

		if w, ok := w.(StmtUnwrapper); ok {
			if w.UnwrapStmt() != wrapped {
				t.Errorf("combination 6/16: UnwrapStmt() failed")
			}
		} else {
			t.Errorf("combination 6/16: StmtUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 7/16: driver.Stmt driver.NamedValueChecker driver.StmtExecContext")
		wrapped := struct {
			driver.Stmt
			driver.NamedValueChecker
			driver.StmtExecContext
		}{}
		w := WrapStmt(wrapped, StmtInterceptor{})

		if _, ok := w.(driver.ColumnConverter); ok != false {
			t.Errorf("combination 7/16: unexpected interface driver.ColumnConverter")
		}
		if _, ok := w.(driver.NamedValueChecker); ok != true {
			t.Errorf("combination 7/16: unexpected interface driver.NamedValueChecker")
		}
		if _, ok := w.(driver.StmtExecContext); ok != true {
			t.Errorf("combination 7/16: unexpected interface driver.StmtExecContext")
		}
		if _, ok := w.(driver.StmtQueryContext); ok != false {
			t.Errorf("combination 7/16: unexpected interface driver.StmtQueryContext")
		}

		if w, ok := w.(StmtUnwrapper); ok {
			if w.UnwrapStmt() != wrapped {
				t.Errorf("combination 7/16: UnwrapStmt() failed")
			}
		} else {
			t.Errorf("combination 7/16: StmtUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 8/16: driver.Stmt driver.ColumnConverter driver.NamedValueChecker driver.StmtExecContext")
		type combination_7_0 interface {
			ColumnConverter(int) driver.ValueConverter
		}
		wrapped := struct {
			driver.Stmt
			combination_7_0
			driver.NamedValueChecker
			driver.StmtExecContext
		}{}
		w := WrapStmt(wrapped, StmtInterceptor{})

		if _, ok := w.(driver.ColumnConverter); ok != true {
			t.Errorf("combination 8/16: unexpected interface driver.ColumnConverter")
		}
		if _, ok := w.(driver.NamedValueChecker); ok != true {
			t.Errorf("combination 8/16: unexpected interface driver.NamedValueChecker")
		}
		if _, ok := w.(driver.StmtExecContext); ok != true {
			t.Errorf("combination 8/16: unexpected interface driver.StmtExecContext")
		}
		if _, ok := w.(driver.StmtQueryContext); ok != false {
			t.Errorf("combination 8/16: unexpected interface driver.StmtQueryContext")
		}

		if w, ok := w.(StmtUnwrapper); ok {
			if w.UnwrapStmt() != wrapped {
				t.Errorf("combination 8/16: UnwrapStmt() failed")
			}
		} else {
			t.Errorf("combination 8/16: StmtUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 9/16: driver.Stmt driver.StmtQueryContext")
		wrapped := struct {
			driver.Stmt
			driver.StmtQueryContext
		}{}
		w := WrapStmt(wrapped, StmtInterceptor{})

		if _, ok := w.(driver.ColumnConverter); ok != false {
			t.Errorf("combination 9/16: unexpected interface driver.ColumnConverter")
		}
		if _, ok := w.(driver.NamedValueChecker); ok != false {
			t.Errorf("combination 9/16: unexpected interface driver.NamedValueChecker")
		}
		if _, ok := w.(driver.StmtExecContext); ok != false {
			t.Errorf("combination 9/16: unexpected interface driver.StmtExecContext")
		}
		if _, ok := w.(driver.StmtQueryContext); ok != true {
			t.Errorf("combination 9/16: unexpected interface driver.StmtQueryContext")
		}

		if w, ok := w.(StmtUnwrapper); ok {
			if w.UnwrapStmt() != wrapped {
				t.Errorf("combination 9/16: UnwrapStmt() failed")
			}
		} else {
			t.Errorf("combination 9/16: StmtUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 10/16: driver.Stmt driver.ColumnConverter driver.StmtQueryContext")
		type combination_9_0 interface {
			ColumnConverter(int) driver.ValueConverter
		}
		wrapped := struct {
			driver.Stmt
			combination_9_0
			driver.StmtQueryContext
		}{}
		w := WrapStmt(wrapped, StmtInterceptor{})

		if _, ok := w.(driver.ColumnConverter); ok != true {
			t.Errorf("combination 10/16: unexpected interface driver.ColumnConverter")
		}
		if _, ok := w.(driver.NamedValueChecker); ok != false {
			t.Errorf("combination 10/16: unexpected interface driver.NamedValueChecker")
		}
		if _, ok := w.(driver.StmtExecContext); ok != false {
			t.Errorf("combination 10/16: unexpected interface driver.StmtExecContext")
		}
		if _, ok := w.(driver.StmtQueryContext); ok != true {
			t.Errorf("combination 10/16: unexpected interface driver.StmtQueryContext")
		}

		if w, ok := w.(StmtUnwrapper); ok {
			if w.UnwrapStmt() != wrapped {
				t.Errorf("combination 10/16: UnwrapStmt() failed")
			}
		} else {
			t.Errorf("combination 10/16: StmtUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 11/16: driver.Stmt driver.NamedValueChecker driver.StmtQueryContext")
		wrapped := struct {
			driver.Stmt
			driver.NamedValueChecker
			driver.StmtQueryContext
		}{}
		w := WrapStmt(wrapped, StmtInterceptor{})

		if _, ok := w.(driver.ColumnConverter); ok != false {
			t.Errorf("combination 11/16: unexpected interface driver.ColumnConverter")
		}
		if _, ok := w.(driver.NamedValueChecker); ok != true {
			t.Errorf("combination 11/16: unexpected interface driver.NamedValueChecker")
		}
		if _, ok := w.(driver.StmtExecContext); ok != false {
			t.Errorf("combination 11/16: unexpected interface driver.StmtExecContext")
		}
		if _, ok := w.(driver.StmtQueryContext); ok != true {
			t.Errorf("combination 11/16: unexpected interface driver.StmtQueryContext")
		}

		if w, ok := w.(StmtUnwrapper); ok {
			if w.UnwrapStmt() != wrapped {
				t.Errorf("combination 11/16: UnwrapStmt() failed")
			}
		} else {
			t.Errorf("combination 11/16: StmtUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 12/16: driver.Stmt driver.ColumnConverter driver.NamedValueChecker driver.StmtQueryContext")
		type combination_11_0 interface {
			ColumnConverter(int) driver.ValueConverter
		}
		wrapped := struct {
			driver.Stmt
			combination_11_0
			driver.NamedValueChecker
			driver.StmtQueryContext
		}{}
		w := WrapStmt(wrapped, StmtInterceptor{})

		if _, ok := w.(driver.ColumnConverter); ok != true {
			t.Errorf("combination 12/16: unexpected interface driver.ColumnConverter")
		}
		if _, ok := w.(driver.NamedValueChecker); ok != true {
			t.Errorf("combination 12/16: unexpected interface driver.NamedValueChecker")
		}
		if _, ok := w.(driver.StmtExecContext); ok != false {
			t.Errorf("combination 12/16: unexpected interface driver.StmtExecContext")
		}
		if _, ok := w.(driver.StmtQueryContext); ok != true {
			t.Errorf("combination 12/16: unexpected interface driver.StmtQueryContext")
		}

		if w, ok := w.(StmtUnwrapper); ok {
			if w.UnwrapStmt() != wrapped {
				t.Errorf("combination 12/16: UnwrapStmt() failed")
			}
		} else {
			t.Errorf("combination 12/16: StmtUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 13/16: driver.Stmt driver.StmtExecContext driver.StmtQueryContext")
		wrapped := struct {
			driver.Stmt
			driver.StmtExecContext
			driver.StmtQueryContext
		}{}
		w := WrapStmt(wrapped, StmtInterceptor{})

		if _, ok := w.(driver.ColumnConverter); ok != false {
			t.Errorf("combination 13/16: unexpected interface driver.ColumnConverter")
		}
		if _, ok := w.(driver.NamedValueChecker); ok != false {
			t.Errorf("combination 13/16: unexpected interface driver.NamedValueChecker")
		}
		if _, ok := w.(driver.StmtExecContext); ok != true {
			t.Errorf("combination 13/16: unexpected interface driver.StmtExecContext")
		}
		if _, ok := w.(driver.StmtQueryContext); ok != true {
			t.Errorf("combination 13/16: unexpected interface driver.StmtQueryContext")
		}

		if w, ok := w.(StmtUnwrapper); ok {
			if w.UnwrapStmt() != wrapped {
				t.Errorf("combination 13/16: UnwrapStmt() failed")
			}
		} else {
			t.Errorf("combination 13/16: StmtUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 14/16: driver.Stmt driver.ColumnConverter driver.StmtExecContext driver.StmtQueryContext")
		type combination_13_0 interface {
			ColumnConverter(int) driver.ValueConverter
		}
		wrapped := struct {
			driver.Stmt
			combination_13_0
			driver.StmtExecContext
			driver.StmtQueryContext
		}{}
		w := WrapStmt(wrapped, StmtInterceptor{})

		if _, ok := w.(driver.ColumnConverter); ok != true {
			t.Errorf("combination 14/16: unexpected interface driver.ColumnConverter")
		}
		if _, ok := w.(driver.NamedValueChecker); ok != false {
			t.Errorf("combination 14/16: unexpected interface driver.NamedValueChecker")
		}
		if _, ok := w.(driver.StmtExecContext); ok != true {
			t.Errorf("combination 14/16: unexpected interface driver.StmtExecContext")
		}
		if _, ok := w.(driver.StmtQueryContext); ok != true {
			t.Errorf("combination 14/16: unexpected interface driver.StmtQueryContext")
		}

		if w, ok := w.(StmtUnwrapper); ok {
			if w.UnwrapStmt() != wrapped {
				t.Errorf("combination 14/16: UnwrapStmt() failed")
			}
		} else {
			t.Errorf("combination 14/16: StmtUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 15/16: driver.Stmt driver.NamedValueChecker driver.StmtExecContext driver.StmtQueryContext")
		wrapped := struct {
			driver.Stmt
			driver.NamedValueChecker
			driver.StmtExecContext
			driver.StmtQueryContext
		}{}
		w := WrapStmt(wrapped, StmtInterceptor{})

		if _, ok := w.(driver.ColumnConverter); ok != false {
			t.Errorf("combination 15/16: unexpected interface driver.ColumnConverter")
		}
		if _, ok := w.(driver.NamedValueChecker); ok != true {
			t.Errorf("combination 15/16: unexpected interface driver.NamedValueChecker")
		}
		if _, ok := w.(driver.StmtExecContext); ok != true {
			t.Errorf("combination 15/16: unexpected interface driver.StmtExecContext")
		}
		if _, ok := w.(driver.StmtQueryContext); ok != true {
			t.Errorf("combination 15/16: unexpected interface driver.StmtQueryContext")
		}

		if w, ok := w.(StmtUnwrapper); ok {
			if w.UnwrapStmt() != wrapped {
				t.Errorf("combination 15/16: UnwrapStmt() failed")
			}
		} else {
			t.Errorf("combination 15/16: StmtUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 16/16: driver.Stmt driver.ColumnConverter driver.NamedValueChecker driver.StmtExecContext driver.StmtQueryContext")
		type combination_15_0 interface {
			ColumnConverter(int) driver.ValueConverter
		}
		wrapped := struct {
			driver.Stmt
			combination_15_0
			driver.NamedValueChecker
			driver.StmtExecContext
			driver.StmtQueryContext
		}{}
		w := WrapStmt(wrapped, StmtInterceptor{})

		if _, ok := w.(driver.ColumnConverter); ok != true {
			t.Errorf("combination 16/16: unexpected interface driver.ColumnConverter")
		}
		if _, ok := w.(driver.NamedValueChecker); ok != true {
			t.Errorf("combination 16/16: unexpected interface driver.NamedValueChecker")
		}
		if _, ok := w.(driver.StmtExecContext); ok != true {
			t.Errorf("combination 16/16: unexpected interface driver.StmtExecContext")
		}
		if _, ok := w.(driver.StmtQueryContext); ok != true {
			t.Errorf("combination 16/16: unexpected interface driver.StmtQueryContext")
		}

		if w, ok := w.(StmtUnwrapper); ok {
			if w.UnwrapStmt() != wrapped {
				t.Errorf("combination 16/16: UnwrapStmt() failed")
			}
		} else {
			t.Errorf("combination 16/16: StmtUnwrapper interface not implemented")
		}
	}
}
