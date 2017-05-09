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
)

// Driver opens an connection to the underlying database.
type Driver interface {
	Open(ctx context.Context, appName string, accessModel AccessModel, optionString string, migrate bool) (Conn, error)
}

// The DriverFunc type is an adapter such that an ordinary function
// can be used as a Driver.
type DriverFunc func(ctx context.Context, appName string, accessModel AccessModel, optionString string, migrate bool) (Conn, error)

// Open returns a Conn by calling the DriverFunc itself.
func (f DriverFunc) Open(ctx context.Context, appName string, accessModel AccessModel, name string, migrate bool) (Conn, error) {
	return f(ctx, appName, accessModel, name, migrate)
}
