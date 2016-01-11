package skydb

import (
	"fmt"
)

var drivers = map[string]Driver{}

// Register makes an Skygear database driver available
// with the given name.
//
// Register panics if it is called with a nil driver or
// the same driver name is being registered twice.
func Register(name string, driver Driver) {
	if driver == nil {
		panic("skydb: Register driver is nil")
	}
	if _, dup := drivers[name]; dup {
		panic("skydb: Register called twice for driver " + name)
	}
	drivers[name] = driver
}

// unregisterAllDrivers unregisters all previously registered drivers.
// Intended for testing.
func unregisterAllDrivers() {
	drivers = map[string]Driver{}
}

var accessModelMap = map[string]AccessModel{
	"role":     RoleBaseAC,
	"relation": RelationBaseAC,
}

// Open returns an implementation of Conn to use w.r.t implName.
//
// optionString is passed to the driver and is implementation specific.
// For example, in a SQL implementation it will be something
// like "sql://localhost/db0"
func Open(implName string, appName string, accessString string, optionString string) (Conn, error) {
	var (
		accessModel AccessModel
		ok          bool
	)
	if accessModel, ok = accessModelMap[accessString]; !ok {
		fmt.Errorf("Received a not supported Access Contol option: %v", accessString)
	}
	if driver, ok := drivers[implName]; ok {
		return driver.Open(appName, accessModel, optionString)
	}

	return nil, fmt.Errorf("Implementation not registered: %v", implName)
}
