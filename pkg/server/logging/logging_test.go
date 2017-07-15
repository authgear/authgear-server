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

package logging

import (
	"testing"

	"github.com/sirupsen/logrus"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLogger(t *testing.T) {
	Convey("loggers", t, func() {
		defer func() {
			loggers = map[string]*logrus.Logger{
				"": logrus.StandardLogger(),
			}
		}()

		Convey("get new logger", func() {
			logger := Logger("hello")
			So(logger, ShouldNotBeNil)
			So(logger, ShouldResemble, loggers["hello"])
		})

		Convey("same logger for same name", func() {
			So(Logger("hello"), ShouldPointTo, Logger("hello"))
		})

		Convey("get root logger", func() {
			So(Logger(""), ShouldPointTo, logrus.StandardLogger())
		})

		Convey("get loggers", func() {
			Logger("hello")
			Logger("world")

			returnedLoggers := Loggers()
			So(returnedLoggers, ShouldContainKey, "hello")
			So(returnedLoggers, ShouldContainKey, "world")
		})
	})
}
