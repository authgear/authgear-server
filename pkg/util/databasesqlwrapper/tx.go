package databasesqlwrapper

import (
	driver "database/sql/driver"
)

type TxUnwrapper interface {
	UnwrapTx() driver.Tx
}

func UnwrapTx(w driver.Tx) driver.Tx {
	if ww, ok := w.(TxUnwrapper); ok {
		return UnwrapTx(ww.UnwrapTx())
	} else {
		return w
	}
}

type Tx_Commit func() error
type Tx_Rollback func() error

type TxInterceptor struct {
	Commit   func(Tx_Commit) Tx_Commit
	Rollback func(Tx_Rollback) Tx_Rollback
}

type TxWrapper struct {
	wrapped     driver.Tx
	interceptor TxInterceptor
}

func (w *TxWrapper) UnwrapTx() driver.Tx {
	return w.wrapped
}

func (w *TxWrapper) Commit() error {
	f := w.wrapped.(driver.Tx).Commit
	if w.interceptor.Commit != nil {
		f = w.interceptor.Commit(f)
	}
	return f()
}

func (w *TxWrapper) Rollback() error {
	f := w.wrapped.(driver.Tx).Rollback
	if w.interceptor.Rollback != nil {
		f = w.interceptor.Rollback(f)
	}
	return f()
}

func WrapTx(wrapped driver.Tx, interceptor TxInterceptor) driver.Tx {
	w := &TxWrapper{wrapped: wrapped, interceptor: interceptor}
	return struct {
		TxUnwrapper
		driver.Tx
	}{w, w}
}
