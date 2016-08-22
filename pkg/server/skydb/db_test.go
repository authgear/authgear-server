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
	"testing"
)

type fakeConn struct {
	Conn
	AppName      string
	AccessModel  AccessModel
	OptionString string
}

type fakeDriver struct {
	Driver
}

func (driver fakeDriver) Open(appName string, accessModel AccessModel, optionString string, migrate bool) (Conn, error) {
	return fakeConn{
		AppName:      appName,
		AccessModel:  accessModel,
		OptionString: optionString,
	}, nil
}

func TestOpen(t *testing.T) {
	defer unregisterAllDrivers()

	Register("fakeImpl", fakeDriver{})

	if driver, err := Open("fakeImpl", "com.example.app.test", "role", "fakeOption", true); err != nil {
		t.Fatalf("got err: %v, want a driver", err.Error())
	} else {
		if driver, ok := driver.(fakeConn); !ok {
			t.Fatalf("got driver = %v, want a driver of type fakeDriver", driver)
		} else {
			if driver.AppName != "com.example.app.test" {
				t.Fatalf("got driver.AppName = %v, want \"com.example.app.test\"", driver.AppName)
			}
			if driver.AccessModel != RoleBasedAccess {
				t.Fatalf("got driver.AccessModel = %v, want \"RoleBasedAccess\"", driver.AccessModel)
			}
			if driver.OptionString != "fakeOption" {
				t.Fatalf("got driver.OptionString = %v, want \"fakeOption\"", driver.OptionString)
			}
		}
	}
}
