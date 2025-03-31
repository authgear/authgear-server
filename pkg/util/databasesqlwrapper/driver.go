package databasesqlwrapper

import (
	driver "database/sql/driver"
)

type DriverUnwrapper interface {
	UnwrapDriver() driver.Driver
}

func UnwrapDriver(w driver.Driver) driver.Driver {
	if ww, ok := w.(DriverUnwrapper); ok {
		return UnwrapDriver(ww.UnwrapDriver())
	} else {
		return w
	}
}

type Driver_Open func(string) (driver.Conn, error)
type Driver_OpenConnector func(string) (driver.Connector, error)

type DriverInterceptor struct {
	Open          func(Driver_Open) Driver_Open
	OpenConnector func(Driver_OpenConnector) Driver_OpenConnector
}

type DriverWrapper struct {
	wrapped     driver.Driver
	interceptor DriverInterceptor
}

func (w *DriverWrapper) UnwrapDriver() driver.Driver {
	return w.wrapped
}

func (w *DriverWrapper) Open(a0 string) (driver.Conn, error) {
	f := w.wrapped.(driver.Driver).Open
	if w.interceptor.Open != nil {
		f = w.interceptor.Open(f)
	}
	return f(a0)
}

func (w *DriverWrapper) OpenConnector(a0 string) (driver.Connector, error) {
	f := w.wrapped.(driver.DriverContext).OpenConnector
	if w.interceptor.OpenConnector != nil {
		f = w.interceptor.OpenConnector(f)
	}
	return f(a0)
}

func WrapDriver(wrapped driver.Driver, interceptor DriverInterceptor) driver.Driver {
	w := &DriverWrapper{wrapped: wrapped, interceptor: interceptor}
	_, ok0 := wrapped.(driver.DriverContext)
	switch {
	// combination 1/2
	case !ok0:
		return struct {
			DriverUnwrapper
			driver.Driver
		}{w, w}
	// combination 2/2
	case ok0:
		return struct {
			DriverUnwrapper
			driver.Driver
			driver.DriverContext
		}{w, w, w}
	}
	panic("unreachable")
}
