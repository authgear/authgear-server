package oddb

import (
	"fmt"
)

var drivers = map[string]Driver{}

// Register makes an Ourd database driver available
// with the given name.
//
// Register panics if it is called with a nil driver or
// the same driver name is being registered twice.
func Register(name string, driver Driver) {
	if driver == nil {
		panic("oddb: Register driver is nil")
	}
	if _, dup := drivers[name]; dup {
		panic("oddb: Register called twice for driver " + name)
	}
	drivers[name] = driver
}

// unregisterAllDrivers unregisters all previously registered drivers.
// Intended for testing.
func unregisterAllDrivers() {
	drivers = map[string]Driver{}
}

// Open returns an implementation of Conn to use w.r.t implName.
//
// optionString is passed to the driver and is implementation specific.
// For example, in a SQL implementation it will be something
// like "sql://localhost/db0"
func Open(implName string, appName string, optionString string) (Conn, error) {
	if driver, ok := drivers[implName]; ok {
		return driver.Open(appName, optionString)
	}
	return nil, fmt.Errorf("Implementation not registered: %v", implName)
}
