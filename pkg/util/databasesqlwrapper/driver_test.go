package databasesqlwrapper

import (
	driver "database/sql/driver"
	testing "testing"
)

func TestWrapDriver(t *testing.T) {
	{
		t.Log("combination 1/2: driver.Driver")
		wrapped := struct {
			driver.Driver
		}{}
		w := WrapDriver(wrapped, DriverInterceptor{})

		if _, ok := w.(driver.DriverContext); ok != false {
			t.Errorf("combination 1/2: unexpected interface driver.DriverContext")
		}

		if w, ok := w.(DriverUnwrapper); ok {
			if w.UnwrapDriver() != wrapped {
				t.Errorf("combination 1/2: UnwrapDriver() failed")
			}
		} else {
			t.Errorf("combination 1/2: DriverUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 2/2: driver.Driver driver.DriverContext")
		wrapped := struct {
			driver.Driver
			driver.DriverContext
		}{}
		w := WrapDriver(wrapped, DriverInterceptor{})

		if _, ok := w.(driver.DriverContext); ok != true {
			t.Errorf("combination 2/2: unexpected interface driver.DriverContext")
		}

		if w, ok := w.(DriverUnwrapper); ok {
			if w.UnwrapDriver() != wrapped {
				t.Errorf("combination 2/2: UnwrapDriver() failed")
			}
		} else {
			t.Errorf("combination 2/2: DriverUnwrapper interface not implemented")
		}
	}
}
