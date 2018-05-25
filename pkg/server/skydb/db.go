// Copyright 2015-present Oursky Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package skydb

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

var drivers = map[string]Driver{}

// Register makes an Skygear Server database driver available
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
	"role":     RoleBasedAccess,
	"relation": RelationBasedAccess,
}

// GetAccessModel convert the string config to internal const
func GetAccessModel(accessString string) AccessModel {
	var (
		model AccessModel
		ok    bool
	)
	if model, ok = accessModelMap[accessString]; !ok {
		logrus.Errorf("Received a not supported Access Contol option: %v", accessString)
	}
	return model
}

// DBConfig represents optional configuration.
// The zero value is sensible defaults.
type DBConfig struct {
	CanMigrate             bool
	PasswordHistoryEnabled bool
}

// DBOpener aliases the function for opening Conn
type DBOpener func(context.Context, string, string, string, string, DBConfig) (Conn, error)

// Open returns an implementation of Conn to use w.r.t implName.
//
// optionString is passed to the driver and is implementation specific.
// For example, in a SQL implementation it will be something
// like "sql://localhost/db0"
func Open(ctx context.Context, implName string, appName string, accessString string, optionString string, config DBConfig) (Conn, error) {
	accessModel := GetAccessModel(accessString)
	if driver, ok := drivers[implName]; ok {
		return driver.Open(ctx, appName, accessModel, optionString, config)
	}

	return nil, fmt.Errorf("Implementation not registered: %v", implName)
}
