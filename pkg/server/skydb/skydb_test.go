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
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMockTest(t *testing.T) {
	Convey("MockTest", t, func() {
		Convey("Mock OK", func() {
			var mockT = func() time.Time { return time.Date(2017, 12, 2, 0, 0, 0, 0, time.UTC) }
			restore := MockTimeNowForTestingOnly(mockT)
			defer restore()
			So(timeNow(), ShouldEqual, mockT())
		})

		Convey("Mock restore OK", func() {
			var mockT = func() time.Time { return time.Date(2017, 12, 2, 0, 0, 0, 0, time.UTC) }
			restore := MockTimeNowForTestingOnly(mockT)
			defer func() {
				So(timeNow(), ShouldEqual, mockT())
				restore()
				So(timeNow(), ShouldNotEqual, mockT())
			}()
		})
	})
}
