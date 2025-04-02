package databasesqlwrapper

import (
	driver "database/sql/driver"
	reflect "reflect"
	testing "testing"
)

func TestWrapRows(t *testing.T) {
	{
		t.Log("combination 1/64: driver.Rows")
		wrapped := struct {
			driver.Rows
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != false {
			t.Errorf("combination 1/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != false {
			t.Errorf("combination 1/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != false {
			t.Errorf("combination 1/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != false {
			t.Errorf("combination 1/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != false {
			t.Errorf("combination 1/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != false {
			t.Errorf("combination 1/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 1/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 1/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 2/64: driver.Rows driver.RowsColumnTypeDatabaseTypeName")
		type combination_1_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		wrapped := struct {
			driver.Rows
			combination_1_0
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != true {
			t.Errorf("combination 2/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != false {
			t.Errorf("combination 2/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != false {
			t.Errorf("combination 2/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != false {
			t.Errorf("combination 2/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != false {
			t.Errorf("combination 2/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != false {
			t.Errorf("combination 2/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 2/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 2/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 3/64: driver.Rows driver.RowsColumnTypeLength")
		type combination_2_1 interface{ ColumnTypeLength(int) (int64, bool) }
		wrapped := struct {
			driver.Rows
			combination_2_1
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != false {
			t.Errorf("combination 3/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != true {
			t.Errorf("combination 3/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != false {
			t.Errorf("combination 3/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != false {
			t.Errorf("combination 3/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != false {
			t.Errorf("combination 3/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != false {
			t.Errorf("combination 3/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 3/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 3/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 4/64: driver.Rows driver.RowsColumnTypeDatabaseTypeName driver.RowsColumnTypeLength")
		type combination_3_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_3_1 interface{ ColumnTypeLength(int) (int64, bool) }
		wrapped := struct {
			driver.Rows
			combination_3_0
			combination_3_1
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != true {
			t.Errorf("combination 4/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != true {
			t.Errorf("combination 4/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != false {
			t.Errorf("combination 4/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != false {
			t.Errorf("combination 4/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != false {
			t.Errorf("combination 4/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != false {
			t.Errorf("combination 4/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 4/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 4/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 5/64: driver.Rows driver.RowsColumnTypeNullable")
		type combination_4_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		wrapped := struct {
			driver.Rows
			combination_4_2
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != false {
			t.Errorf("combination 5/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != false {
			t.Errorf("combination 5/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != true {
			t.Errorf("combination 5/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != false {
			t.Errorf("combination 5/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != false {
			t.Errorf("combination 5/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != false {
			t.Errorf("combination 5/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 5/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 5/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 6/64: driver.Rows driver.RowsColumnTypeDatabaseTypeName driver.RowsColumnTypeNullable")
		type combination_5_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_5_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		wrapped := struct {
			driver.Rows
			combination_5_0
			combination_5_2
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != true {
			t.Errorf("combination 6/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != false {
			t.Errorf("combination 6/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != true {
			t.Errorf("combination 6/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != false {
			t.Errorf("combination 6/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != false {
			t.Errorf("combination 6/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != false {
			t.Errorf("combination 6/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 6/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 6/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 7/64: driver.Rows driver.RowsColumnTypeLength driver.RowsColumnTypeNullable")
		type combination_6_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_6_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		wrapped := struct {
			driver.Rows
			combination_6_1
			combination_6_2
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != false {
			t.Errorf("combination 7/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != true {
			t.Errorf("combination 7/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != true {
			t.Errorf("combination 7/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != false {
			t.Errorf("combination 7/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != false {
			t.Errorf("combination 7/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != false {
			t.Errorf("combination 7/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 7/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 7/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 8/64: driver.Rows driver.RowsColumnTypeDatabaseTypeName driver.RowsColumnTypeLength driver.RowsColumnTypeNullable")
		type combination_7_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_7_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_7_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		wrapped := struct {
			driver.Rows
			combination_7_0
			combination_7_1
			combination_7_2
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != true {
			t.Errorf("combination 8/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != true {
			t.Errorf("combination 8/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != true {
			t.Errorf("combination 8/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != false {
			t.Errorf("combination 8/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != false {
			t.Errorf("combination 8/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != false {
			t.Errorf("combination 8/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 8/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 8/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 9/64: driver.Rows driver.RowsColumnTypePrecisionScale")
		type combination_8_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		wrapped := struct {
			driver.Rows
			combination_8_3
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != false {
			t.Errorf("combination 9/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != false {
			t.Errorf("combination 9/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != false {
			t.Errorf("combination 9/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != true {
			t.Errorf("combination 9/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != false {
			t.Errorf("combination 9/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != false {
			t.Errorf("combination 9/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 9/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 9/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 10/64: driver.Rows driver.RowsColumnTypeDatabaseTypeName driver.RowsColumnTypePrecisionScale")
		type combination_9_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_9_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		wrapped := struct {
			driver.Rows
			combination_9_0
			combination_9_3
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != true {
			t.Errorf("combination 10/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != false {
			t.Errorf("combination 10/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != false {
			t.Errorf("combination 10/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != true {
			t.Errorf("combination 10/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != false {
			t.Errorf("combination 10/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != false {
			t.Errorf("combination 10/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 10/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 10/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 11/64: driver.Rows driver.RowsColumnTypeLength driver.RowsColumnTypePrecisionScale")
		type combination_10_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_10_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		wrapped := struct {
			driver.Rows
			combination_10_1
			combination_10_3
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != false {
			t.Errorf("combination 11/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != true {
			t.Errorf("combination 11/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != false {
			t.Errorf("combination 11/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != true {
			t.Errorf("combination 11/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != false {
			t.Errorf("combination 11/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != false {
			t.Errorf("combination 11/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 11/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 11/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 12/64: driver.Rows driver.RowsColumnTypeDatabaseTypeName driver.RowsColumnTypeLength driver.RowsColumnTypePrecisionScale")
		type combination_11_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_11_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_11_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		wrapped := struct {
			driver.Rows
			combination_11_0
			combination_11_1
			combination_11_3
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != true {
			t.Errorf("combination 12/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != true {
			t.Errorf("combination 12/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != false {
			t.Errorf("combination 12/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != true {
			t.Errorf("combination 12/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != false {
			t.Errorf("combination 12/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != false {
			t.Errorf("combination 12/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 12/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 12/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 13/64: driver.Rows driver.RowsColumnTypeNullable driver.RowsColumnTypePrecisionScale")
		type combination_12_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_12_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		wrapped := struct {
			driver.Rows
			combination_12_2
			combination_12_3
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != false {
			t.Errorf("combination 13/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != false {
			t.Errorf("combination 13/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != true {
			t.Errorf("combination 13/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != true {
			t.Errorf("combination 13/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != false {
			t.Errorf("combination 13/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != false {
			t.Errorf("combination 13/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 13/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 13/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 14/64: driver.Rows driver.RowsColumnTypeDatabaseTypeName driver.RowsColumnTypeNullable driver.RowsColumnTypePrecisionScale")
		type combination_13_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_13_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_13_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		wrapped := struct {
			driver.Rows
			combination_13_0
			combination_13_2
			combination_13_3
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != true {
			t.Errorf("combination 14/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != false {
			t.Errorf("combination 14/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != true {
			t.Errorf("combination 14/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != true {
			t.Errorf("combination 14/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != false {
			t.Errorf("combination 14/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != false {
			t.Errorf("combination 14/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 14/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 14/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 15/64: driver.Rows driver.RowsColumnTypeLength driver.RowsColumnTypeNullable driver.RowsColumnTypePrecisionScale")
		type combination_14_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_14_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_14_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		wrapped := struct {
			driver.Rows
			combination_14_1
			combination_14_2
			combination_14_3
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != false {
			t.Errorf("combination 15/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != true {
			t.Errorf("combination 15/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != true {
			t.Errorf("combination 15/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != true {
			t.Errorf("combination 15/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != false {
			t.Errorf("combination 15/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != false {
			t.Errorf("combination 15/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 15/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 15/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 16/64: driver.Rows driver.RowsColumnTypeDatabaseTypeName driver.RowsColumnTypeLength driver.RowsColumnTypeNullable driver.RowsColumnTypePrecisionScale")
		type combination_15_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_15_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_15_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_15_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		wrapped := struct {
			driver.Rows
			combination_15_0
			combination_15_1
			combination_15_2
			combination_15_3
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != true {
			t.Errorf("combination 16/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != true {
			t.Errorf("combination 16/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != true {
			t.Errorf("combination 16/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != true {
			t.Errorf("combination 16/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != false {
			t.Errorf("combination 16/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != false {
			t.Errorf("combination 16/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 16/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 16/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 17/64: driver.Rows driver.RowsColumnTypeScanType")
		type combination_16_4 interface{ ColumnTypeScanType(int) reflect.Type }
		wrapped := struct {
			driver.Rows
			combination_16_4
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != false {
			t.Errorf("combination 17/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != false {
			t.Errorf("combination 17/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != false {
			t.Errorf("combination 17/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != false {
			t.Errorf("combination 17/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != true {
			t.Errorf("combination 17/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != false {
			t.Errorf("combination 17/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 17/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 17/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 18/64: driver.Rows driver.RowsColumnTypeDatabaseTypeName driver.RowsColumnTypeScanType")
		type combination_17_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_17_4 interface{ ColumnTypeScanType(int) reflect.Type }
		wrapped := struct {
			driver.Rows
			combination_17_0
			combination_17_4
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != true {
			t.Errorf("combination 18/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != false {
			t.Errorf("combination 18/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != false {
			t.Errorf("combination 18/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != false {
			t.Errorf("combination 18/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != true {
			t.Errorf("combination 18/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != false {
			t.Errorf("combination 18/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 18/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 18/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 19/64: driver.Rows driver.RowsColumnTypeLength driver.RowsColumnTypeScanType")
		type combination_18_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_18_4 interface{ ColumnTypeScanType(int) reflect.Type }
		wrapped := struct {
			driver.Rows
			combination_18_1
			combination_18_4
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != false {
			t.Errorf("combination 19/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != true {
			t.Errorf("combination 19/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != false {
			t.Errorf("combination 19/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != false {
			t.Errorf("combination 19/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != true {
			t.Errorf("combination 19/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != false {
			t.Errorf("combination 19/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 19/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 19/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 20/64: driver.Rows driver.RowsColumnTypeDatabaseTypeName driver.RowsColumnTypeLength driver.RowsColumnTypeScanType")
		type combination_19_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_19_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_19_4 interface{ ColumnTypeScanType(int) reflect.Type }
		wrapped := struct {
			driver.Rows
			combination_19_0
			combination_19_1
			combination_19_4
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != true {
			t.Errorf("combination 20/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != true {
			t.Errorf("combination 20/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != false {
			t.Errorf("combination 20/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != false {
			t.Errorf("combination 20/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != true {
			t.Errorf("combination 20/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != false {
			t.Errorf("combination 20/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 20/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 20/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 21/64: driver.Rows driver.RowsColumnTypeNullable driver.RowsColumnTypeScanType")
		type combination_20_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_20_4 interface{ ColumnTypeScanType(int) reflect.Type }
		wrapped := struct {
			driver.Rows
			combination_20_2
			combination_20_4
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != false {
			t.Errorf("combination 21/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != false {
			t.Errorf("combination 21/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != true {
			t.Errorf("combination 21/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != false {
			t.Errorf("combination 21/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != true {
			t.Errorf("combination 21/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != false {
			t.Errorf("combination 21/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 21/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 21/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 22/64: driver.Rows driver.RowsColumnTypeDatabaseTypeName driver.RowsColumnTypeNullable driver.RowsColumnTypeScanType")
		type combination_21_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_21_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_21_4 interface{ ColumnTypeScanType(int) reflect.Type }
		wrapped := struct {
			driver.Rows
			combination_21_0
			combination_21_2
			combination_21_4
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != true {
			t.Errorf("combination 22/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != false {
			t.Errorf("combination 22/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != true {
			t.Errorf("combination 22/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != false {
			t.Errorf("combination 22/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != true {
			t.Errorf("combination 22/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != false {
			t.Errorf("combination 22/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 22/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 22/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 23/64: driver.Rows driver.RowsColumnTypeLength driver.RowsColumnTypeNullable driver.RowsColumnTypeScanType")
		type combination_22_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_22_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_22_4 interface{ ColumnTypeScanType(int) reflect.Type }
		wrapped := struct {
			driver.Rows
			combination_22_1
			combination_22_2
			combination_22_4
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != false {
			t.Errorf("combination 23/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != true {
			t.Errorf("combination 23/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != true {
			t.Errorf("combination 23/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != false {
			t.Errorf("combination 23/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != true {
			t.Errorf("combination 23/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != false {
			t.Errorf("combination 23/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 23/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 23/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 24/64: driver.Rows driver.RowsColumnTypeDatabaseTypeName driver.RowsColumnTypeLength driver.RowsColumnTypeNullable driver.RowsColumnTypeScanType")
		type combination_23_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_23_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_23_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_23_4 interface{ ColumnTypeScanType(int) reflect.Type }
		wrapped := struct {
			driver.Rows
			combination_23_0
			combination_23_1
			combination_23_2
			combination_23_4
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != true {
			t.Errorf("combination 24/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != true {
			t.Errorf("combination 24/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != true {
			t.Errorf("combination 24/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != false {
			t.Errorf("combination 24/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != true {
			t.Errorf("combination 24/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != false {
			t.Errorf("combination 24/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 24/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 24/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 25/64: driver.Rows driver.RowsColumnTypePrecisionScale driver.RowsColumnTypeScanType")
		type combination_24_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_24_4 interface{ ColumnTypeScanType(int) reflect.Type }
		wrapped := struct {
			driver.Rows
			combination_24_3
			combination_24_4
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != false {
			t.Errorf("combination 25/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != false {
			t.Errorf("combination 25/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != false {
			t.Errorf("combination 25/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != true {
			t.Errorf("combination 25/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != true {
			t.Errorf("combination 25/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != false {
			t.Errorf("combination 25/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 25/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 25/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 26/64: driver.Rows driver.RowsColumnTypeDatabaseTypeName driver.RowsColumnTypePrecisionScale driver.RowsColumnTypeScanType")
		type combination_25_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_25_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_25_4 interface{ ColumnTypeScanType(int) reflect.Type }
		wrapped := struct {
			driver.Rows
			combination_25_0
			combination_25_3
			combination_25_4
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != true {
			t.Errorf("combination 26/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != false {
			t.Errorf("combination 26/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != false {
			t.Errorf("combination 26/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != true {
			t.Errorf("combination 26/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != true {
			t.Errorf("combination 26/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != false {
			t.Errorf("combination 26/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 26/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 26/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 27/64: driver.Rows driver.RowsColumnTypeLength driver.RowsColumnTypePrecisionScale driver.RowsColumnTypeScanType")
		type combination_26_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_26_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_26_4 interface{ ColumnTypeScanType(int) reflect.Type }
		wrapped := struct {
			driver.Rows
			combination_26_1
			combination_26_3
			combination_26_4
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != false {
			t.Errorf("combination 27/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != true {
			t.Errorf("combination 27/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != false {
			t.Errorf("combination 27/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != true {
			t.Errorf("combination 27/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != true {
			t.Errorf("combination 27/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != false {
			t.Errorf("combination 27/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 27/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 27/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 28/64: driver.Rows driver.RowsColumnTypeDatabaseTypeName driver.RowsColumnTypeLength driver.RowsColumnTypePrecisionScale driver.RowsColumnTypeScanType")
		type combination_27_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_27_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_27_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_27_4 interface{ ColumnTypeScanType(int) reflect.Type }
		wrapped := struct {
			driver.Rows
			combination_27_0
			combination_27_1
			combination_27_3
			combination_27_4
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != true {
			t.Errorf("combination 28/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != true {
			t.Errorf("combination 28/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != false {
			t.Errorf("combination 28/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != true {
			t.Errorf("combination 28/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != true {
			t.Errorf("combination 28/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != false {
			t.Errorf("combination 28/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 28/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 28/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 29/64: driver.Rows driver.RowsColumnTypeNullable driver.RowsColumnTypePrecisionScale driver.RowsColumnTypeScanType")
		type combination_28_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_28_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_28_4 interface{ ColumnTypeScanType(int) reflect.Type }
		wrapped := struct {
			driver.Rows
			combination_28_2
			combination_28_3
			combination_28_4
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != false {
			t.Errorf("combination 29/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != false {
			t.Errorf("combination 29/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != true {
			t.Errorf("combination 29/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != true {
			t.Errorf("combination 29/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != true {
			t.Errorf("combination 29/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != false {
			t.Errorf("combination 29/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 29/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 29/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 30/64: driver.Rows driver.RowsColumnTypeDatabaseTypeName driver.RowsColumnTypeNullable driver.RowsColumnTypePrecisionScale driver.RowsColumnTypeScanType")
		type combination_29_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_29_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_29_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_29_4 interface{ ColumnTypeScanType(int) reflect.Type }
		wrapped := struct {
			driver.Rows
			combination_29_0
			combination_29_2
			combination_29_3
			combination_29_4
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != true {
			t.Errorf("combination 30/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != false {
			t.Errorf("combination 30/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != true {
			t.Errorf("combination 30/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != true {
			t.Errorf("combination 30/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != true {
			t.Errorf("combination 30/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != false {
			t.Errorf("combination 30/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 30/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 30/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 31/64: driver.Rows driver.RowsColumnTypeLength driver.RowsColumnTypeNullable driver.RowsColumnTypePrecisionScale driver.RowsColumnTypeScanType")
		type combination_30_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_30_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_30_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_30_4 interface{ ColumnTypeScanType(int) reflect.Type }
		wrapped := struct {
			driver.Rows
			combination_30_1
			combination_30_2
			combination_30_3
			combination_30_4
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != false {
			t.Errorf("combination 31/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != true {
			t.Errorf("combination 31/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != true {
			t.Errorf("combination 31/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != true {
			t.Errorf("combination 31/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != true {
			t.Errorf("combination 31/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != false {
			t.Errorf("combination 31/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 31/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 31/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 32/64: driver.Rows driver.RowsColumnTypeDatabaseTypeName driver.RowsColumnTypeLength driver.RowsColumnTypeNullable driver.RowsColumnTypePrecisionScale driver.RowsColumnTypeScanType")
		type combination_31_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_31_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_31_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_31_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_31_4 interface{ ColumnTypeScanType(int) reflect.Type }
		wrapped := struct {
			driver.Rows
			combination_31_0
			combination_31_1
			combination_31_2
			combination_31_3
			combination_31_4
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != true {
			t.Errorf("combination 32/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != true {
			t.Errorf("combination 32/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != true {
			t.Errorf("combination 32/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != true {
			t.Errorf("combination 32/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != true {
			t.Errorf("combination 32/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != false {
			t.Errorf("combination 32/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 32/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 32/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 33/64: driver.Rows driver.RowsNextResultSet")
		type combination_32_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		wrapped := struct {
			driver.Rows
			combination_32_5
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != false {
			t.Errorf("combination 33/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != false {
			t.Errorf("combination 33/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != false {
			t.Errorf("combination 33/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != false {
			t.Errorf("combination 33/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != false {
			t.Errorf("combination 33/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != true {
			t.Errorf("combination 33/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 33/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 33/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 34/64: driver.Rows driver.RowsColumnTypeDatabaseTypeName driver.RowsNextResultSet")
		type combination_33_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_33_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		wrapped := struct {
			driver.Rows
			combination_33_0
			combination_33_5
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != true {
			t.Errorf("combination 34/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != false {
			t.Errorf("combination 34/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != false {
			t.Errorf("combination 34/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != false {
			t.Errorf("combination 34/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != false {
			t.Errorf("combination 34/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != true {
			t.Errorf("combination 34/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 34/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 34/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 35/64: driver.Rows driver.RowsColumnTypeLength driver.RowsNextResultSet")
		type combination_34_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_34_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		wrapped := struct {
			driver.Rows
			combination_34_1
			combination_34_5
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != false {
			t.Errorf("combination 35/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != true {
			t.Errorf("combination 35/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != false {
			t.Errorf("combination 35/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != false {
			t.Errorf("combination 35/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != false {
			t.Errorf("combination 35/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != true {
			t.Errorf("combination 35/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 35/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 35/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 36/64: driver.Rows driver.RowsColumnTypeDatabaseTypeName driver.RowsColumnTypeLength driver.RowsNextResultSet")
		type combination_35_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_35_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_35_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		wrapped := struct {
			driver.Rows
			combination_35_0
			combination_35_1
			combination_35_5
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != true {
			t.Errorf("combination 36/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != true {
			t.Errorf("combination 36/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != false {
			t.Errorf("combination 36/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != false {
			t.Errorf("combination 36/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != false {
			t.Errorf("combination 36/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != true {
			t.Errorf("combination 36/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 36/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 36/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 37/64: driver.Rows driver.RowsColumnTypeNullable driver.RowsNextResultSet")
		type combination_36_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_36_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		wrapped := struct {
			driver.Rows
			combination_36_2
			combination_36_5
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != false {
			t.Errorf("combination 37/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != false {
			t.Errorf("combination 37/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != true {
			t.Errorf("combination 37/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != false {
			t.Errorf("combination 37/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != false {
			t.Errorf("combination 37/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != true {
			t.Errorf("combination 37/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 37/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 37/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 38/64: driver.Rows driver.RowsColumnTypeDatabaseTypeName driver.RowsColumnTypeNullable driver.RowsNextResultSet")
		type combination_37_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_37_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_37_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		wrapped := struct {
			driver.Rows
			combination_37_0
			combination_37_2
			combination_37_5
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != true {
			t.Errorf("combination 38/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != false {
			t.Errorf("combination 38/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != true {
			t.Errorf("combination 38/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != false {
			t.Errorf("combination 38/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != false {
			t.Errorf("combination 38/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != true {
			t.Errorf("combination 38/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 38/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 38/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 39/64: driver.Rows driver.RowsColumnTypeLength driver.RowsColumnTypeNullable driver.RowsNextResultSet")
		type combination_38_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_38_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_38_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		wrapped := struct {
			driver.Rows
			combination_38_1
			combination_38_2
			combination_38_5
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != false {
			t.Errorf("combination 39/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != true {
			t.Errorf("combination 39/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != true {
			t.Errorf("combination 39/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != false {
			t.Errorf("combination 39/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != false {
			t.Errorf("combination 39/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != true {
			t.Errorf("combination 39/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 39/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 39/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 40/64: driver.Rows driver.RowsColumnTypeDatabaseTypeName driver.RowsColumnTypeLength driver.RowsColumnTypeNullable driver.RowsNextResultSet")
		type combination_39_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_39_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_39_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_39_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		wrapped := struct {
			driver.Rows
			combination_39_0
			combination_39_1
			combination_39_2
			combination_39_5
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != true {
			t.Errorf("combination 40/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != true {
			t.Errorf("combination 40/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != true {
			t.Errorf("combination 40/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != false {
			t.Errorf("combination 40/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != false {
			t.Errorf("combination 40/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != true {
			t.Errorf("combination 40/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 40/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 40/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 41/64: driver.Rows driver.RowsColumnTypePrecisionScale driver.RowsNextResultSet")
		type combination_40_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_40_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		wrapped := struct {
			driver.Rows
			combination_40_3
			combination_40_5
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != false {
			t.Errorf("combination 41/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != false {
			t.Errorf("combination 41/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != false {
			t.Errorf("combination 41/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != true {
			t.Errorf("combination 41/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != false {
			t.Errorf("combination 41/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != true {
			t.Errorf("combination 41/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 41/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 41/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 42/64: driver.Rows driver.RowsColumnTypeDatabaseTypeName driver.RowsColumnTypePrecisionScale driver.RowsNextResultSet")
		type combination_41_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_41_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_41_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		wrapped := struct {
			driver.Rows
			combination_41_0
			combination_41_3
			combination_41_5
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != true {
			t.Errorf("combination 42/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != false {
			t.Errorf("combination 42/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != false {
			t.Errorf("combination 42/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != true {
			t.Errorf("combination 42/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != false {
			t.Errorf("combination 42/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != true {
			t.Errorf("combination 42/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 42/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 42/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 43/64: driver.Rows driver.RowsColumnTypeLength driver.RowsColumnTypePrecisionScale driver.RowsNextResultSet")
		type combination_42_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_42_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_42_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		wrapped := struct {
			driver.Rows
			combination_42_1
			combination_42_3
			combination_42_5
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != false {
			t.Errorf("combination 43/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != true {
			t.Errorf("combination 43/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != false {
			t.Errorf("combination 43/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != true {
			t.Errorf("combination 43/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != false {
			t.Errorf("combination 43/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != true {
			t.Errorf("combination 43/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 43/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 43/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 44/64: driver.Rows driver.RowsColumnTypeDatabaseTypeName driver.RowsColumnTypeLength driver.RowsColumnTypePrecisionScale driver.RowsNextResultSet")
		type combination_43_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_43_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_43_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_43_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		wrapped := struct {
			driver.Rows
			combination_43_0
			combination_43_1
			combination_43_3
			combination_43_5
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != true {
			t.Errorf("combination 44/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != true {
			t.Errorf("combination 44/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != false {
			t.Errorf("combination 44/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != true {
			t.Errorf("combination 44/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != false {
			t.Errorf("combination 44/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != true {
			t.Errorf("combination 44/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 44/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 44/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 45/64: driver.Rows driver.RowsColumnTypeNullable driver.RowsColumnTypePrecisionScale driver.RowsNextResultSet")
		type combination_44_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_44_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_44_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		wrapped := struct {
			driver.Rows
			combination_44_2
			combination_44_3
			combination_44_5
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != false {
			t.Errorf("combination 45/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != false {
			t.Errorf("combination 45/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != true {
			t.Errorf("combination 45/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != true {
			t.Errorf("combination 45/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != false {
			t.Errorf("combination 45/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != true {
			t.Errorf("combination 45/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 45/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 45/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 46/64: driver.Rows driver.RowsColumnTypeDatabaseTypeName driver.RowsColumnTypeNullable driver.RowsColumnTypePrecisionScale driver.RowsNextResultSet")
		type combination_45_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_45_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_45_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_45_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		wrapped := struct {
			driver.Rows
			combination_45_0
			combination_45_2
			combination_45_3
			combination_45_5
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != true {
			t.Errorf("combination 46/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != false {
			t.Errorf("combination 46/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != true {
			t.Errorf("combination 46/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != true {
			t.Errorf("combination 46/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != false {
			t.Errorf("combination 46/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != true {
			t.Errorf("combination 46/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 46/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 46/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 47/64: driver.Rows driver.RowsColumnTypeLength driver.RowsColumnTypeNullable driver.RowsColumnTypePrecisionScale driver.RowsNextResultSet")
		type combination_46_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_46_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_46_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_46_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		wrapped := struct {
			driver.Rows
			combination_46_1
			combination_46_2
			combination_46_3
			combination_46_5
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != false {
			t.Errorf("combination 47/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != true {
			t.Errorf("combination 47/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != true {
			t.Errorf("combination 47/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != true {
			t.Errorf("combination 47/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != false {
			t.Errorf("combination 47/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != true {
			t.Errorf("combination 47/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 47/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 47/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 48/64: driver.Rows driver.RowsColumnTypeDatabaseTypeName driver.RowsColumnTypeLength driver.RowsColumnTypeNullable driver.RowsColumnTypePrecisionScale driver.RowsNextResultSet")
		type combination_47_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_47_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_47_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_47_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_47_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		wrapped := struct {
			driver.Rows
			combination_47_0
			combination_47_1
			combination_47_2
			combination_47_3
			combination_47_5
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != true {
			t.Errorf("combination 48/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != true {
			t.Errorf("combination 48/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != true {
			t.Errorf("combination 48/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != true {
			t.Errorf("combination 48/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != false {
			t.Errorf("combination 48/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != true {
			t.Errorf("combination 48/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 48/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 48/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 49/64: driver.Rows driver.RowsColumnTypeScanType driver.RowsNextResultSet")
		type combination_48_4 interface{ ColumnTypeScanType(int) reflect.Type }
		type combination_48_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		wrapped := struct {
			driver.Rows
			combination_48_4
			combination_48_5
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != false {
			t.Errorf("combination 49/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != false {
			t.Errorf("combination 49/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != false {
			t.Errorf("combination 49/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != false {
			t.Errorf("combination 49/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != true {
			t.Errorf("combination 49/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != true {
			t.Errorf("combination 49/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 49/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 49/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 50/64: driver.Rows driver.RowsColumnTypeDatabaseTypeName driver.RowsColumnTypeScanType driver.RowsNextResultSet")
		type combination_49_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_49_4 interface{ ColumnTypeScanType(int) reflect.Type }
		type combination_49_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		wrapped := struct {
			driver.Rows
			combination_49_0
			combination_49_4
			combination_49_5
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != true {
			t.Errorf("combination 50/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != false {
			t.Errorf("combination 50/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != false {
			t.Errorf("combination 50/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != false {
			t.Errorf("combination 50/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != true {
			t.Errorf("combination 50/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != true {
			t.Errorf("combination 50/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 50/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 50/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 51/64: driver.Rows driver.RowsColumnTypeLength driver.RowsColumnTypeScanType driver.RowsNextResultSet")
		type combination_50_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_50_4 interface{ ColumnTypeScanType(int) reflect.Type }
		type combination_50_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		wrapped := struct {
			driver.Rows
			combination_50_1
			combination_50_4
			combination_50_5
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != false {
			t.Errorf("combination 51/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != true {
			t.Errorf("combination 51/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != false {
			t.Errorf("combination 51/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != false {
			t.Errorf("combination 51/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != true {
			t.Errorf("combination 51/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != true {
			t.Errorf("combination 51/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 51/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 51/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 52/64: driver.Rows driver.RowsColumnTypeDatabaseTypeName driver.RowsColumnTypeLength driver.RowsColumnTypeScanType driver.RowsNextResultSet")
		type combination_51_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_51_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_51_4 interface{ ColumnTypeScanType(int) reflect.Type }
		type combination_51_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		wrapped := struct {
			driver.Rows
			combination_51_0
			combination_51_1
			combination_51_4
			combination_51_5
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != true {
			t.Errorf("combination 52/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != true {
			t.Errorf("combination 52/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != false {
			t.Errorf("combination 52/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != false {
			t.Errorf("combination 52/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != true {
			t.Errorf("combination 52/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != true {
			t.Errorf("combination 52/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 52/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 52/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 53/64: driver.Rows driver.RowsColumnTypeNullable driver.RowsColumnTypeScanType driver.RowsNextResultSet")
		type combination_52_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_52_4 interface{ ColumnTypeScanType(int) reflect.Type }
		type combination_52_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		wrapped := struct {
			driver.Rows
			combination_52_2
			combination_52_4
			combination_52_5
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != false {
			t.Errorf("combination 53/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != false {
			t.Errorf("combination 53/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != true {
			t.Errorf("combination 53/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != false {
			t.Errorf("combination 53/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != true {
			t.Errorf("combination 53/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != true {
			t.Errorf("combination 53/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 53/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 53/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 54/64: driver.Rows driver.RowsColumnTypeDatabaseTypeName driver.RowsColumnTypeNullable driver.RowsColumnTypeScanType driver.RowsNextResultSet")
		type combination_53_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_53_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_53_4 interface{ ColumnTypeScanType(int) reflect.Type }
		type combination_53_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		wrapped := struct {
			driver.Rows
			combination_53_0
			combination_53_2
			combination_53_4
			combination_53_5
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != true {
			t.Errorf("combination 54/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != false {
			t.Errorf("combination 54/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != true {
			t.Errorf("combination 54/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != false {
			t.Errorf("combination 54/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != true {
			t.Errorf("combination 54/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != true {
			t.Errorf("combination 54/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 54/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 54/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 55/64: driver.Rows driver.RowsColumnTypeLength driver.RowsColumnTypeNullable driver.RowsColumnTypeScanType driver.RowsNextResultSet")
		type combination_54_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_54_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_54_4 interface{ ColumnTypeScanType(int) reflect.Type }
		type combination_54_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		wrapped := struct {
			driver.Rows
			combination_54_1
			combination_54_2
			combination_54_4
			combination_54_5
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != false {
			t.Errorf("combination 55/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != true {
			t.Errorf("combination 55/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != true {
			t.Errorf("combination 55/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != false {
			t.Errorf("combination 55/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != true {
			t.Errorf("combination 55/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != true {
			t.Errorf("combination 55/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 55/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 55/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 56/64: driver.Rows driver.RowsColumnTypeDatabaseTypeName driver.RowsColumnTypeLength driver.RowsColumnTypeNullable driver.RowsColumnTypeScanType driver.RowsNextResultSet")
		type combination_55_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_55_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_55_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_55_4 interface{ ColumnTypeScanType(int) reflect.Type }
		type combination_55_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		wrapped := struct {
			driver.Rows
			combination_55_0
			combination_55_1
			combination_55_2
			combination_55_4
			combination_55_5
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != true {
			t.Errorf("combination 56/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != true {
			t.Errorf("combination 56/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != true {
			t.Errorf("combination 56/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != false {
			t.Errorf("combination 56/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != true {
			t.Errorf("combination 56/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != true {
			t.Errorf("combination 56/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 56/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 56/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 57/64: driver.Rows driver.RowsColumnTypePrecisionScale driver.RowsColumnTypeScanType driver.RowsNextResultSet")
		type combination_56_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_56_4 interface{ ColumnTypeScanType(int) reflect.Type }
		type combination_56_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		wrapped := struct {
			driver.Rows
			combination_56_3
			combination_56_4
			combination_56_5
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != false {
			t.Errorf("combination 57/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != false {
			t.Errorf("combination 57/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != false {
			t.Errorf("combination 57/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != true {
			t.Errorf("combination 57/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != true {
			t.Errorf("combination 57/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != true {
			t.Errorf("combination 57/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 57/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 57/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 58/64: driver.Rows driver.RowsColumnTypeDatabaseTypeName driver.RowsColumnTypePrecisionScale driver.RowsColumnTypeScanType driver.RowsNextResultSet")
		type combination_57_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_57_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_57_4 interface{ ColumnTypeScanType(int) reflect.Type }
		type combination_57_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		wrapped := struct {
			driver.Rows
			combination_57_0
			combination_57_3
			combination_57_4
			combination_57_5
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != true {
			t.Errorf("combination 58/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != false {
			t.Errorf("combination 58/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != false {
			t.Errorf("combination 58/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != true {
			t.Errorf("combination 58/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != true {
			t.Errorf("combination 58/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != true {
			t.Errorf("combination 58/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 58/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 58/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 59/64: driver.Rows driver.RowsColumnTypeLength driver.RowsColumnTypePrecisionScale driver.RowsColumnTypeScanType driver.RowsNextResultSet")
		type combination_58_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_58_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_58_4 interface{ ColumnTypeScanType(int) reflect.Type }
		type combination_58_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		wrapped := struct {
			driver.Rows
			combination_58_1
			combination_58_3
			combination_58_4
			combination_58_5
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != false {
			t.Errorf("combination 59/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != true {
			t.Errorf("combination 59/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != false {
			t.Errorf("combination 59/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != true {
			t.Errorf("combination 59/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != true {
			t.Errorf("combination 59/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != true {
			t.Errorf("combination 59/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 59/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 59/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 60/64: driver.Rows driver.RowsColumnTypeDatabaseTypeName driver.RowsColumnTypeLength driver.RowsColumnTypePrecisionScale driver.RowsColumnTypeScanType driver.RowsNextResultSet")
		type combination_59_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_59_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_59_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_59_4 interface{ ColumnTypeScanType(int) reflect.Type }
		type combination_59_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		wrapped := struct {
			driver.Rows
			combination_59_0
			combination_59_1
			combination_59_3
			combination_59_4
			combination_59_5
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != true {
			t.Errorf("combination 60/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != true {
			t.Errorf("combination 60/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != false {
			t.Errorf("combination 60/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != true {
			t.Errorf("combination 60/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != true {
			t.Errorf("combination 60/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != true {
			t.Errorf("combination 60/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 60/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 60/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 61/64: driver.Rows driver.RowsColumnTypeNullable driver.RowsColumnTypePrecisionScale driver.RowsColumnTypeScanType driver.RowsNextResultSet")
		type combination_60_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_60_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_60_4 interface{ ColumnTypeScanType(int) reflect.Type }
		type combination_60_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		wrapped := struct {
			driver.Rows
			combination_60_2
			combination_60_3
			combination_60_4
			combination_60_5
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != false {
			t.Errorf("combination 61/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != false {
			t.Errorf("combination 61/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != true {
			t.Errorf("combination 61/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != true {
			t.Errorf("combination 61/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != true {
			t.Errorf("combination 61/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != true {
			t.Errorf("combination 61/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 61/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 61/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 62/64: driver.Rows driver.RowsColumnTypeDatabaseTypeName driver.RowsColumnTypeNullable driver.RowsColumnTypePrecisionScale driver.RowsColumnTypeScanType driver.RowsNextResultSet")
		type combination_61_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_61_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_61_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_61_4 interface{ ColumnTypeScanType(int) reflect.Type }
		type combination_61_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		wrapped := struct {
			driver.Rows
			combination_61_0
			combination_61_2
			combination_61_3
			combination_61_4
			combination_61_5
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != true {
			t.Errorf("combination 62/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != false {
			t.Errorf("combination 62/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != true {
			t.Errorf("combination 62/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != true {
			t.Errorf("combination 62/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != true {
			t.Errorf("combination 62/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != true {
			t.Errorf("combination 62/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 62/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 62/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 63/64: driver.Rows driver.RowsColumnTypeLength driver.RowsColumnTypeNullable driver.RowsColumnTypePrecisionScale driver.RowsColumnTypeScanType driver.RowsNextResultSet")
		type combination_62_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_62_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_62_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_62_4 interface{ ColumnTypeScanType(int) reflect.Type }
		type combination_62_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		wrapped := struct {
			driver.Rows
			combination_62_1
			combination_62_2
			combination_62_3
			combination_62_4
			combination_62_5
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != false {
			t.Errorf("combination 63/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != true {
			t.Errorf("combination 63/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != true {
			t.Errorf("combination 63/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != true {
			t.Errorf("combination 63/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != true {
			t.Errorf("combination 63/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != true {
			t.Errorf("combination 63/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 63/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 63/64: RowsUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 64/64: driver.Rows driver.RowsColumnTypeDatabaseTypeName driver.RowsColumnTypeLength driver.RowsColumnTypeNullable driver.RowsColumnTypePrecisionScale driver.RowsColumnTypeScanType driver.RowsNextResultSet")
		type combination_63_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_63_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_63_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_63_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_63_4 interface{ ColumnTypeScanType(int) reflect.Type }
		type combination_63_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		wrapped := struct {
			driver.Rows
			combination_63_0
			combination_63_1
			combination_63_2
			combination_63_3
			combination_63_4
			combination_63_5
		}{}
		w := WrapRows(wrapped, RowsInterceptor{})

		if _, ok := w.(driver.RowsColumnTypeDatabaseTypeName); ok != true {
			t.Errorf("combination 64/64: unexpected interface driver.RowsColumnTypeDatabaseTypeName")
		}
		if _, ok := w.(driver.RowsColumnTypeLength); ok != true {
			t.Errorf("combination 64/64: unexpected interface driver.RowsColumnTypeLength")
		}
		if _, ok := w.(driver.RowsColumnTypeNullable); ok != true {
			t.Errorf("combination 64/64: unexpected interface driver.RowsColumnTypeNullable")
		}
		if _, ok := w.(driver.RowsColumnTypePrecisionScale); ok != true {
			t.Errorf("combination 64/64: unexpected interface driver.RowsColumnTypePrecisionScale")
		}
		if _, ok := w.(driver.RowsColumnTypeScanType); ok != true {
			t.Errorf("combination 64/64: unexpected interface driver.RowsColumnTypeScanType")
		}
		if _, ok := w.(driver.RowsNextResultSet); ok != true {
			t.Errorf("combination 64/64: unexpected interface driver.RowsNextResultSet")
		}

		if w, ok := w.(RowsUnwrapper); ok {
			if w.UnwrapRows() != wrapped {
				t.Errorf("combination 64/64: UnwrapRows() failed")
			}
		} else {
			t.Errorf("combination 64/64: RowsUnwrapper interface not implemented")
		}
	}
}
