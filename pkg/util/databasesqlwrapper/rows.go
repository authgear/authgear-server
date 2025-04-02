package databasesqlwrapper

import (
	driver "database/sql/driver"
	reflect "reflect"
)

type RowsUnwrapper interface {
	UnwrapRows() driver.Rows
}

func UnwrapRows(w driver.Rows) driver.Rows {
	if ww, ok := w.(RowsUnwrapper); ok {
		return UnwrapRows(ww.UnwrapRows())
	} else {
		return w
	}
}

type Rows_Close func() error
type Rows_Columns func() []string
type Rows_Next func([]driver.Value) error
type Rows_ColumnTypeDatabaseTypeName func(int) string
type Rows_ColumnTypeLength func(int) (int64, bool)
type Rows_ColumnTypeNullable func(int) (bool, bool)
type Rows_ColumnTypePrecisionScale func(int) (int64, int64, bool)
type Rows_ColumnTypeScanType func(int) reflect.Type
type Rows_HasNextResultSet func() bool
type Rows_NextResultSet func() error

type RowsInterceptor struct {
	Close                      func(Rows_Close) Rows_Close
	Columns                    func(Rows_Columns) Rows_Columns
	Next                       func(Rows_Next) Rows_Next
	ColumnTypeDatabaseTypeName func(Rows_ColumnTypeDatabaseTypeName) Rows_ColumnTypeDatabaseTypeName
	ColumnTypeLength           func(Rows_ColumnTypeLength) Rows_ColumnTypeLength
	ColumnTypeNullable         func(Rows_ColumnTypeNullable) Rows_ColumnTypeNullable
	ColumnTypePrecisionScale   func(Rows_ColumnTypePrecisionScale) Rows_ColumnTypePrecisionScale
	ColumnTypeScanType         func(Rows_ColumnTypeScanType) Rows_ColumnTypeScanType
	HasNextResultSet           func(Rows_HasNextResultSet) Rows_HasNextResultSet
	NextResultSet              func(Rows_NextResultSet) Rows_NextResultSet
}

type RowsWrapper struct {
	wrapped     driver.Rows
	interceptor RowsInterceptor
}

func (w *RowsWrapper) UnwrapRows() driver.Rows {
	return w.wrapped
}

func (w *RowsWrapper) Close() error {
	f := w.wrapped.(driver.Rows).Close
	if w.interceptor.Close != nil {
		f = w.interceptor.Close(f)
	}
	return f()
}

func (w *RowsWrapper) Columns() []string {
	f := w.wrapped.(driver.Rows).Columns
	if w.interceptor.Columns != nil {
		f = w.interceptor.Columns(f)
	}
	return f()
}

func (w *RowsWrapper) Next(a0 []driver.Value) error {
	f := w.wrapped.(driver.Rows).Next
	if w.interceptor.Next != nil {
		f = w.interceptor.Next(f)
	}
	return f(a0)
}

func (w *RowsWrapper) ColumnTypeDatabaseTypeName(a0 int) string {
	f := w.wrapped.(driver.RowsColumnTypeDatabaseTypeName).ColumnTypeDatabaseTypeName
	if w.interceptor.ColumnTypeDatabaseTypeName != nil {
		f = w.interceptor.ColumnTypeDatabaseTypeName(f)
	}
	return f(a0)
}

func (w *RowsWrapper) ColumnTypeLength(a0 int) (int64, bool) {
	f := w.wrapped.(driver.RowsColumnTypeLength).ColumnTypeLength
	if w.interceptor.ColumnTypeLength != nil {
		f = w.interceptor.ColumnTypeLength(f)
	}
	return f(a0)
}

func (w *RowsWrapper) ColumnTypeNullable(a0 int) (bool, bool) {
	f := w.wrapped.(driver.RowsColumnTypeNullable).ColumnTypeNullable
	if w.interceptor.ColumnTypeNullable != nil {
		f = w.interceptor.ColumnTypeNullable(f)
	}
	return f(a0)
}

func (w *RowsWrapper) ColumnTypePrecisionScale(a0 int) (int64, int64, bool) {
	f := w.wrapped.(driver.RowsColumnTypePrecisionScale).ColumnTypePrecisionScale
	if w.interceptor.ColumnTypePrecisionScale != nil {
		f = w.interceptor.ColumnTypePrecisionScale(f)
	}
	return f(a0)
}

func (w *RowsWrapper) ColumnTypeScanType(a0 int) reflect.Type {
	f := w.wrapped.(driver.RowsColumnTypeScanType).ColumnTypeScanType
	if w.interceptor.ColumnTypeScanType != nil {
		f = w.interceptor.ColumnTypeScanType(f)
	}
	return f(a0)
}

func (w *RowsWrapper) HasNextResultSet() bool {
	f := w.wrapped.(driver.RowsNextResultSet).HasNextResultSet
	if w.interceptor.HasNextResultSet != nil {
		f = w.interceptor.HasNextResultSet(f)
	}
	return f()
}

func (w *RowsWrapper) NextResultSet() error {
	f := w.wrapped.(driver.RowsNextResultSet).NextResultSet
	if w.interceptor.NextResultSet != nil {
		f = w.interceptor.NextResultSet(f)
	}
	return f()
}

func WrapRows(wrapped driver.Rows, interceptor RowsInterceptor) driver.Rows {
	w := &RowsWrapper{wrapped: wrapped, interceptor: interceptor}
	_, ok0 := wrapped.(driver.RowsColumnTypeDatabaseTypeName)
	_, ok1 := wrapped.(driver.RowsColumnTypeLength)
	_, ok2 := wrapped.(driver.RowsColumnTypeNullable)
	_, ok3 := wrapped.(driver.RowsColumnTypePrecisionScale)
	_, ok4 := wrapped.(driver.RowsColumnTypeScanType)
	_, ok5 := wrapped.(driver.RowsNextResultSet)
	switch {
	// combination 1/64
	case !ok0 && !ok1 && !ok2 && !ok3 && !ok4 && !ok5:
		return struct {
			RowsUnwrapper
			driver.Rows
		}{w, w}
	// combination 2/64
	case ok0 && !ok1 && !ok2 && !ok3 && !ok4 && !ok5:
		type combination_1_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_1_0
		}{w, w, w}
	// combination 3/64
	case !ok0 && ok1 && !ok2 && !ok3 && !ok4 && !ok5:
		type combination_2_1 interface{ ColumnTypeLength(int) (int64, bool) }
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_2_1
		}{w, w, w}
	// combination 4/64
	case ok0 && ok1 && !ok2 && !ok3 && !ok4 && !ok5:
		type combination_3_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_3_1 interface{ ColumnTypeLength(int) (int64, bool) }
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_3_0
			combination_3_1
		}{w, w, w, w}
	// combination 5/64
	case !ok0 && !ok1 && ok2 && !ok3 && !ok4 && !ok5:
		type combination_4_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_4_2
		}{w, w, w}
	// combination 6/64
	case ok0 && !ok1 && ok2 && !ok3 && !ok4 && !ok5:
		type combination_5_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_5_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_5_0
			combination_5_2
		}{w, w, w, w}
	// combination 7/64
	case !ok0 && ok1 && ok2 && !ok3 && !ok4 && !ok5:
		type combination_6_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_6_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_6_1
			combination_6_2
		}{w, w, w, w}
	// combination 8/64
	case ok0 && ok1 && ok2 && !ok3 && !ok4 && !ok5:
		type combination_7_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_7_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_7_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_7_0
			combination_7_1
			combination_7_2
		}{w, w, w, w, w}
	// combination 9/64
	case !ok0 && !ok1 && !ok2 && ok3 && !ok4 && !ok5:
		type combination_8_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_8_3
		}{w, w, w}
	// combination 10/64
	case ok0 && !ok1 && !ok2 && ok3 && !ok4 && !ok5:
		type combination_9_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_9_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_9_0
			combination_9_3
		}{w, w, w, w}
	// combination 11/64
	case !ok0 && ok1 && !ok2 && ok3 && !ok4 && !ok5:
		type combination_10_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_10_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_10_1
			combination_10_3
		}{w, w, w, w}
	// combination 12/64
	case ok0 && ok1 && !ok2 && ok3 && !ok4 && !ok5:
		type combination_11_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_11_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_11_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_11_0
			combination_11_1
			combination_11_3
		}{w, w, w, w, w}
	// combination 13/64
	case !ok0 && !ok1 && ok2 && ok3 && !ok4 && !ok5:
		type combination_12_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_12_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_12_2
			combination_12_3
		}{w, w, w, w}
	// combination 14/64
	case ok0 && !ok1 && ok2 && ok3 && !ok4 && !ok5:
		type combination_13_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_13_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_13_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_13_0
			combination_13_2
			combination_13_3
		}{w, w, w, w, w}
	// combination 15/64
	case !ok0 && ok1 && ok2 && ok3 && !ok4 && !ok5:
		type combination_14_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_14_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_14_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_14_1
			combination_14_2
			combination_14_3
		}{w, w, w, w, w}
	// combination 16/64
	case ok0 && ok1 && ok2 && ok3 && !ok4 && !ok5:
		type combination_15_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_15_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_15_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_15_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_15_0
			combination_15_1
			combination_15_2
			combination_15_3
		}{w, w, w, w, w, w}
	// combination 17/64
	case !ok0 && !ok1 && !ok2 && !ok3 && ok4 && !ok5:
		type combination_16_4 interface{ ColumnTypeScanType(int) reflect.Type }
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_16_4
		}{w, w, w}
	// combination 18/64
	case ok0 && !ok1 && !ok2 && !ok3 && ok4 && !ok5:
		type combination_17_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_17_4 interface{ ColumnTypeScanType(int) reflect.Type }
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_17_0
			combination_17_4
		}{w, w, w, w}
	// combination 19/64
	case !ok0 && ok1 && !ok2 && !ok3 && ok4 && !ok5:
		type combination_18_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_18_4 interface{ ColumnTypeScanType(int) reflect.Type }
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_18_1
			combination_18_4
		}{w, w, w, w}
	// combination 20/64
	case ok0 && ok1 && !ok2 && !ok3 && ok4 && !ok5:
		type combination_19_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_19_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_19_4 interface{ ColumnTypeScanType(int) reflect.Type }
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_19_0
			combination_19_1
			combination_19_4
		}{w, w, w, w, w}
	// combination 21/64
	case !ok0 && !ok1 && ok2 && !ok3 && ok4 && !ok5:
		type combination_20_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_20_4 interface{ ColumnTypeScanType(int) reflect.Type }
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_20_2
			combination_20_4
		}{w, w, w, w}
	// combination 22/64
	case ok0 && !ok1 && ok2 && !ok3 && ok4 && !ok5:
		type combination_21_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_21_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_21_4 interface{ ColumnTypeScanType(int) reflect.Type }
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_21_0
			combination_21_2
			combination_21_4
		}{w, w, w, w, w}
	// combination 23/64
	case !ok0 && ok1 && ok2 && !ok3 && ok4 && !ok5:
		type combination_22_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_22_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_22_4 interface{ ColumnTypeScanType(int) reflect.Type }
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_22_1
			combination_22_2
			combination_22_4
		}{w, w, w, w, w}
	// combination 24/64
	case ok0 && ok1 && ok2 && !ok3 && ok4 && !ok5:
		type combination_23_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_23_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_23_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_23_4 interface{ ColumnTypeScanType(int) reflect.Type }
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_23_0
			combination_23_1
			combination_23_2
			combination_23_4
		}{w, w, w, w, w, w}
	// combination 25/64
	case !ok0 && !ok1 && !ok2 && ok3 && ok4 && !ok5:
		type combination_24_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_24_4 interface{ ColumnTypeScanType(int) reflect.Type }
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_24_3
			combination_24_4
		}{w, w, w, w}
	// combination 26/64
	case ok0 && !ok1 && !ok2 && ok3 && ok4 && !ok5:
		type combination_25_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_25_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_25_4 interface{ ColumnTypeScanType(int) reflect.Type }
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_25_0
			combination_25_3
			combination_25_4
		}{w, w, w, w, w}
	// combination 27/64
	case !ok0 && ok1 && !ok2 && ok3 && ok4 && !ok5:
		type combination_26_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_26_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_26_4 interface{ ColumnTypeScanType(int) reflect.Type }
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_26_1
			combination_26_3
			combination_26_4
		}{w, w, w, w, w}
	// combination 28/64
	case ok0 && ok1 && !ok2 && ok3 && ok4 && !ok5:
		type combination_27_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_27_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_27_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_27_4 interface{ ColumnTypeScanType(int) reflect.Type }
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_27_0
			combination_27_1
			combination_27_3
			combination_27_4
		}{w, w, w, w, w, w}
	// combination 29/64
	case !ok0 && !ok1 && ok2 && ok3 && ok4 && !ok5:
		type combination_28_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_28_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_28_4 interface{ ColumnTypeScanType(int) reflect.Type }
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_28_2
			combination_28_3
			combination_28_4
		}{w, w, w, w, w}
	// combination 30/64
	case ok0 && !ok1 && ok2 && ok3 && ok4 && !ok5:
		type combination_29_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_29_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_29_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_29_4 interface{ ColumnTypeScanType(int) reflect.Type }
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_29_0
			combination_29_2
			combination_29_3
			combination_29_4
		}{w, w, w, w, w, w}
	// combination 31/64
	case !ok0 && ok1 && ok2 && ok3 && ok4 && !ok5:
		type combination_30_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_30_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_30_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_30_4 interface{ ColumnTypeScanType(int) reflect.Type }
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_30_1
			combination_30_2
			combination_30_3
			combination_30_4
		}{w, w, w, w, w, w}
	// combination 32/64
	case ok0 && ok1 && ok2 && ok3 && ok4 && !ok5:
		type combination_31_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_31_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_31_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_31_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_31_4 interface{ ColumnTypeScanType(int) reflect.Type }
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_31_0
			combination_31_1
			combination_31_2
			combination_31_3
			combination_31_4
		}{w, w, w, w, w, w, w}
	// combination 33/64
	case !ok0 && !ok1 && !ok2 && !ok3 && !ok4 && ok5:
		type combination_32_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_32_5
		}{w, w, w}
	// combination 34/64
	case ok0 && !ok1 && !ok2 && !ok3 && !ok4 && ok5:
		type combination_33_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_33_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_33_0
			combination_33_5
		}{w, w, w, w}
	// combination 35/64
	case !ok0 && ok1 && !ok2 && !ok3 && !ok4 && ok5:
		type combination_34_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_34_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_34_1
			combination_34_5
		}{w, w, w, w}
	// combination 36/64
	case ok0 && ok1 && !ok2 && !ok3 && !ok4 && ok5:
		type combination_35_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_35_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_35_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_35_0
			combination_35_1
			combination_35_5
		}{w, w, w, w, w}
	// combination 37/64
	case !ok0 && !ok1 && ok2 && !ok3 && !ok4 && ok5:
		type combination_36_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_36_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_36_2
			combination_36_5
		}{w, w, w, w}
	// combination 38/64
	case ok0 && !ok1 && ok2 && !ok3 && !ok4 && ok5:
		type combination_37_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_37_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_37_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_37_0
			combination_37_2
			combination_37_5
		}{w, w, w, w, w}
	// combination 39/64
	case !ok0 && ok1 && ok2 && !ok3 && !ok4 && ok5:
		type combination_38_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_38_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_38_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_38_1
			combination_38_2
			combination_38_5
		}{w, w, w, w, w}
	// combination 40/64
	case ok0 && ok1 && ok2 && !ok3 && !ok4 && ok5:
		type combination_39_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_39_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_39_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_39_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_39_0
			combination_39_1
			combination_39_2
			combination_39_5
		}{w, w, w, w, w, w}
	// combination 41/64
	case !ok0 && !ok1 && !ok2 && ok3 && !ok4 && ok5:
		type combination_40_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_40_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_40_3
			combination_40_5
		}{w, w, w, w}
	// combination 42/64
	case ok0 && !ok1 && !ok2 && ok3 && !ok4 && ok5:
		type combination_41_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_41_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_41_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_41_0
			combination_41_3
			combination_41_5
		}{w, w, w, w, w}
	// combination 43/64
	case !ok0 && ok1 && !ok2 && ok3 && !ok4 && ok5:
		type combination_42_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_42_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_42_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_42_1
			combination_42_3
			combination_42_5
		}{w, w, w, w, w}
	// combination 44/64
	case ok0 && ok1 && !ok2 && ok3 && !ok4 && ok5:
		type combination_43_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_43_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_43_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_43_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_43_0
			combination_43_1
			combination_43_3
			combination_43_5
		}{w, w, w, w, w, w}
	// combination 45/64
	case !ok0 && !ok1 && ok2 && ok3 && !ok4 && ok5:
		type combination_44_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_44_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_44_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_44_2
			combination_44_3
			combination_44_5
		}{w, w, w, w, w}
	// combination 46/64
	case ok0 && !ok1 && ok2 && ok3 && !ok4 && ok5:
		type combination_45_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_45_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_45_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_45_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_45_0
			combination_45_2
			combination_45_3
			combination_45_5
		}{w, w, w, w, w, w}
	// combination 47/64
	case !ok0 && ok1 && ok2 && ok3 && !ok4 && ok5:
		type combination_46_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_46_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_46_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_46_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_46_1
			combination_46_2
			combination_46_3
			combination_46_5
		}{w, w, w, w, w, w}
	// combination 48/64
	case ok0 && ok1 && ok2 && ok3 && !ok4 && ok5:
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
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_47_0
			combination_47_1
			combination_47_2
			combination_47_3
			combination_47_5
		}{w, w, w, w, w, w, w}
	// combination 49/64
	case !ok0 && !ok1 && !ok2 && !ok3 && ok4 && ok5:
		type combination_48_4 interface{ ColumnTypeScanType(int) reflect.Type }
		type combination_48_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_48_4
			combination_48_5
		}{w, w, w, w}
	// combination 50/64
	case ok0 && !ok1 && !ok2 && !ok3 && ok4 && ok5:
		type combination_49_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_49_4 interface{ ColumnTypeScanType(int) reflect.Type }
		type combination_49_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_49_0
			combination_49_4
			combination_49_5
		}{w, w, w, w, w}
	// combination 51/64
	case !ok0 && ok1 && !ok2 && !ok3 && ok4 && ok5:
		type combination_50_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_50_4 interface{ ColumnTypeScanType(int) reflect.Type }
		type combination_50_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_50_1
			combination_50_4
			combination_50_5
		}{w, w, w, w, w}
	// combination 52/64
	case ok0 && ok1 && !ok2 && !ok3 && ok4 && ok5:
		type combination_51_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_51_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_51_4 interface{ ColumnTypeScanType(int) reflect.Type }
		type combination_51_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_51_0
			combination_51_1
			combination_51_4
			combination_51_5
		}{w, w, w, w, w, w}
	// combination 53/64
	case !ok0 && !ok1 && ok2 && !ok3 && ok4 && ok5:
		type combination_52_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_52_4 interface{ ColumnTypeScanType(int) reflect.Type }
		type combination_52_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_52_2
			combination_52_4
			combination_52_5
		}{w, w, w, w, w}
	// combination 54/64
	case ok0 && !ok1 && ok2 && !ok3 && ok4 && ok5:
		type combination_53_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_53_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_53_4 interface{ ColumnTypeScanType(int) reflect.Type }
		type combination_53_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_53_0
			combination_53_2
			combination_53_4
			combination_53_5
		}{w, w, w, w, w, w}
	// combination 55/64
	case !ok0 && ok1 && ok2 && !ok3 && ok4 && ok5:
		type combination_54_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_54_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_54_4 interface{ ColumnTypeScanType(int) reflect.Type }
		type combination_54_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_54_1
			combination_54_2
			combination_54_4
			combination_54_5
		}{w, w, w, w, w, w}
	// combination 56/64
	case ok0 && ok1 && ok2 && !ok3 && ok4 && ok5:
		type combination_55_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_55_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_55_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_55_4 interface{ ColumnTypeScanType(int) reflect.Type }
		type combination_55_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_55_0
			combination_55_1
			combination_55_2
			combination_55_4
			combination_55_5
		}{w, w, w, w, w, w, w}
	// combination 57/64
	case !ok0 && !ok1 && !ok2 && ok3 && ok4 && ok5:
		type combination_56_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_56_4 interface{ ColumnTypeScanType(int) reflect.Type }
		type combination_56_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_56_3
			combination_56_4
			combination_56_5
		}{w, w, w, w, w}
	// combination 58/64
	case ok0 && !ok1 && !ok2 && ok3 && ok4 && ok5:
		type combination_57_0 interface{ ColumnTypeDatabaseTypeName(int) string }
		type combination_57_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_57_4 interface{ ColumnTypeScanType(int) reflect.Type }
		type combination_57_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_57_0
			combination_57_3
			combination_57_4
			combination_57_5
		}{w, w, w, w, w, w}
	// combination 59/64
	case !ok0 && ok1 && !ok2 && ok3 && ok4 && ok5:
		type combination_58_1 interface{ ColumnTypeLength(int) (int64, bool) }
		type combination_58_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_58_4 interface{ ColumnTypeScanType(int) reflect.Type }
		type combination_58_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_58_1
			combination_58_3
			combination_58_4
			combination_58_5
		}{w, w, w, w, w, w}
	// combination 60/64
	case ok0 && ok1 && !ok2 && ok3 && ok4 && ok5:
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
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_59_0
			combination_59_1
			combination_59_3
			combination_59_4
			combination_59_5
		}{w, w, w, w, w, w, w}
	// combination 61/64
	case !ok0 && !ok1 && ok2 && ok3 && ok4 && ok5:
		type combination_60_2 interface{ ColumnTypeNullable(int) (bool, bool) }
		type combination_60_3 interface {
			ColumnTypePrecisionScale(int) (int64, int64, bool)
		}
		type combination_60_4 interface{ ColumnTypeScanType(int) reflect.Type }
		type combination_60_5 interface {
			HasNextResultSet() bool
			NextResultSet() error
		}
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_60_2
			combination_60_3
			combination_60_4
			combination_60_5
		}{w, w, w, w, w, w}
	// combination 62/64
	case ok0 && !ok1 && ok2 && ok3 && ok4 && ok5:
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
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_61_0
			combination_61_2
			combination_61_3
			combination_61_4
			combination_61_5
		}{w, w, w, w, w, w, w}
	// combination 63/64
	case !ok0 && ok1 && ok2 && ok3 && ok4 && ok5:
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
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_62_1
			combination_62_2
			combination_62_3
			combination_62_4
			combination_62_5
		}{w, w, w, w, w, w, w}
	// combination 64/64
	case ok0 && ok1 && ok2 && ok3 && ok4 && ok5:
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
		return struct {
			RowsUnwrapper
			driver.Rows
			combination_63_0
			combination_63_1
			combination_63_2
			combination_63_3
			combination_63_4
			combination_63_5
		}{w, w, w, w, w, w, w, w}
	}
	panic("unreachable")
}
