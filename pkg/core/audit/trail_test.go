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

package audit

import (
	"github.com/evalphobia/logrus_fluent"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestCreateHook(t *testing.T) {
	Convey("createHook with unsupported scheme", t, func() {
		hook, err := createHook("http://is-not-supported")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "unknown handler: http, http://is-not-supported")
		So(hook, ShouldBeNil)
	})

	Convey("createHook with malformed file path", t, func() {
		hook, err := createHook("file://malformed-path")
		So(err, ShouldNotBeNil)
		So(hook, ShouldBeNil)
	})

	Convey("createHook with malformed host", t, func() {
		hook, err := createHook("fluentd:/malformed-host")
		So(err, ShouldNotBeNil)
		So(hook, ShouldBeNil)
	})

	Convey("createHook with fluentd", t, func() {
		hook, err := createHook("fluentd://my-fluentd:12345")
		So(err, ShouldBeNil)
		So(hook, ShouldNotBeNil)
		fluentd := hook.(*logrus_fluent.FluentHook).Fluent
		So(fluentd.FluentHost, ShouldEqual, "my-fluentd")
		So(fluentd.FluentPort, ShouldEqual, 12345)
	})

	Convey("createHook with file url", t, func() {
		hook, err := createHook("file:///var/log/skygear/audit.log")
		So(err, ShouldBeNil)
		So(hook, ShouldNotBeNil)
	})
}
