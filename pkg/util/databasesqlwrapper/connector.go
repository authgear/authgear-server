package databasesqlwrapper

import (
	context "context"
	driver "database/sql/driver"
	io "io"
)

type ConnectorUnwrapper interface {
	UnwrapConnector() driver.Connector
}

func UnwrapConnector(w driver.Connector) driver.Connector {
	if ww, ok := w.(ConnectorUnwrapper); ok {
		return UnwrapConnector(ww.UnwrapConnector())
	} else {
		return w
	}
}

type Connector_Connect func(context.Context) (driver.Conn, error)
type Connector_Driver func() driver.Driver
type Connector_Close func() error

type ConnectorInterceptor struct {
	Connect func(Connector_Connect) Connector_Connect
	Driver  func(Connector_Driver) Connector_Driver
	Close   func(Connector_Close) Connector_Close
}

type ConnectorWrapper struct {
	wrapped     driver.Connector
	interceptor ConnectorInterceptor
}

func (w *ConnectorWrapper) UnwrapConnector() driver.Connector {
	return w.wrapped
}

func (w *ConnectorWrapper) Connect(a0 context.Context) (driver.Conn, error) {
	f := w.wrapped.(driver.Connector).Connect
	if w.interceptor.Connect != nil {
		f = w.interceptor.Connect(f)
	}
	return f(a0)
}

func (w *ConnectorWrapper) Driver() driver.Driver {
	f := w.wrapped.(driver.Connector).Driver
	if w.interceptor.Driver != nil {
		f = w.interceptor.Driver(f)
	}
	return f()
}

func (w *ConnectorWrapper) Close() error {
	f := w.wrapped.(io.Closer).Close
	if w.interceptor.Close != nil {
		f = w.interceptor.Close(f)
	}
	return f()
}

func WrapConnector(wrapped driver.Connector, interceptor ConnectorInterceptor) driver.Connector {
	w := &ConnectorWrapper{wrapped: wrapped, interceptor: interceptor}
	_, ok0 := wrapped.(io.Closer)
	switch {
	// combination 1/2
	case !ok0:
		return struct {
			ConnectorUnwrapper
			driver.Connector
		}{w, w}
	// combination 2/2
	case ok0:
		return struct {
			ConnectorUnwrapper
			driver.Connector
			io.Closer
		}{w, w, w}
	}
	panic("unreachable")
}
