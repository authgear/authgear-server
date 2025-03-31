package databasesqlwrapper

import (
	driver "database/sql/driver"
	io "io"
	testing "testing"
)

func TestWrapConnector(t *testing.T) {
	{
		t.Log("combination 1/2: driver.Connector")
		wrapped := struct {
			driver.Connector
		}{}
		w := WrapConnector(wrapped, ConnectorInterceptor{})

		if _, ok := w.(io.Closer); ok != false {
			t.Errorf("combination 1/2: unexpected interface io.Closer")
		}

		if w, ok := w.(ConnectorUnwrapper); ok {
			if w.UnwrapConnector() != wrapped {
				t.Errorf("combination 1/2: UnwrapConnector() failed")
			}
		} else {
			t.Errorf("combination 1/2: ConnectorUnwrapper interface not implemented")
		}
	}
	{
		t.Log("combination 2/2: driver.Connector io.Closer")
		wrapped := struct {
			driver.Connector
			io.Closer
		}{}
		w := WrapConnector(wrapped, ConnectorInterceptor{})

		if _, ok := w.(io.Closer); ok != true {
			t.Errorf("combination 2/2: unexpected interface io.Closer")
		}

		if w, ok := w.(ConnectorUnwrapper); ok {
			if w.UnwrapConnector() != wrapped {
				t.Errorf("combination 2/2: UnwrapConnector() failed")
			}
		} else {
			t.Errorf("combination 2/2: ConnectorUnwrapper interface not implemented")
		}
	}
}
