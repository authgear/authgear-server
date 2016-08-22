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

package skyerr

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"fmt"
)

func TestNewError(t *testing.T) {
	Convey("An Error", t, func() {
		err := NewError(10000, "some message")

		Convey("returns code correctly", func() {
			So(err.Code(), ShouldEqual, 10000)
		})

		Convey("returns message correctly", func() {
			So(err.Message(), ShouldEqual, "some message")
		})

		Convey("Error()s in format {code}: {message}", func() {
			So(err.Error(), ShouldEqual, "UnexpectedError: some message")
		})

		Convey("has format {code}: {message} when being written", func() {
			So(fmt.Sprintf("%v", err), ShouldEqual, "UnexpectedError: some message")
		})
	})
}

func TestNewErrorf(t *testing.T) {
	Convey("NewErrorf", t, func() {
		err := NewErrorf(2, "obj1: %v, obj2: %v", "string", 0)

		Convey("creates err with correct code", func() {
			So(err.Code(), ShouldEqual, 2)
		})

		Convey("creates err with correct message", func() {
			So(err.Message(), ShouldEqual, "obj1: string, obj2: 0")
		})
	})
}
