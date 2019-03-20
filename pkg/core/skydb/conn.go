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
	"errors"
)

// ErrUserDuplicated is returned by Conn.CreateAuth when
// the AuthInfo to be created has the same ID/username in the current container
var ErrUserDuplicated = errors.New("skydb: duplicated UserID or Username")

// ErrUserNotFound is returned by Conn.GetAuth, Conn.UpdateAuth and
// Conn.DeleteAuth when the AuthInfo's ID is not found
// in the current container
var ErrUserNotFound = errors.New("skydb: AuthInfo ID not found")

// ErrDeviceNotFound is returned by Conn.GetDevice, Conn.DeleteDevice,
// Conn.DeleteDevicesByToken and Conn.DeleteEmptyDevicesByTime, if the desired Device
// cannot be found in the current container
var ErrDeviceNotFound = errors.New("skydb: Specific device not found")

// ErrDatabaseIsReadOnly is returned by skydb.Database if the requested
// operation modifies the database and the database is readonly.
var ErrDatabaseIsReadOnly = errors.New("skydb: database is read only")
