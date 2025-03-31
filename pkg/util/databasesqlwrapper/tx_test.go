package databasesqlwrapper

import (
	driver "database/sql/driver"
	testing "testing"
)

func TestWrapTx(t *testing.T) {
	{
		t.Log("combination 1/1: driver.Tx")
		wrapped := struct {
			driver.Tx
		}{}
		w := WrapTx(wrapped, TxInterceptor{})

		if w, ok := w.(TxUnwrapper); ok {
			if w.UnwrapTx() != wrapped {
				t.Errorf("combination 1/1: UnwrapTx() failed")
			}
		} else {
			t.Errorf("combination 1/1: TxUnwrapper interface not implemented")
		}
	}
}
