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

package handler

import (
	"testing"
	"time"

	"github.com/skygeario/skygear-server/pkg/server/handler/handlertest"
	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	. "github.com/skygeario/skygear-server/pkg/server/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSetDisableUserHandler(t *testing.T) {
	Convey("SetDisableUserHandler", t, func() {
		conn := singleUserConn{}
		enabledAuthInfo := skydb.NewAuthInfo("chima")
		enabledAuthInfo.ID = "chima"
		conn.CreateAuth(&enabledAuthInfo)

		disabledAuthInfo := skydb.NewAuthInfo("faseng")
		disabledAuthInfo.ID = "faseng"
		disabledAuthInfo.Disabled = true
		disabledAuthInfo.DisabledMessage = "some reason"
		expiry := time.Date(2017, 7, 23, 19, 30, 24, 0, time.UTC)
		disabledAuthInfo.DisabledExpiry = &expiry
		conn.CreateAuth(&disabledAuthInfo)

		r := handlertest.NewSingleRouteRouter(&SetDisableUserHandler{}, func(p *router.Payload) {
			p.DBConn = &conn
		})

		Convey("should disable a user", func() {
			resp := r.POST(`
				{
					"auth_id": "chima",
					"disabled": true
				}
			`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `
				{
					"result": {"status": "OK"}
				}
			`)
			So(resp.Code, ShouldEqual, 200)

			var fetchedAuth skydb.AuthInfo
			conn.GetAuth("chima", &fetchedAuth)
			So(fetchedAuth.Disabled, ShouldBeTrue)
			So(fetchedAuth.DisabledMessage, ShouldBeEmpty)
			So(fetchedAuth.DisabledExpiry, ShouldBeNil)
		})

		Convey("should disable a user with message and expiry", func() {
			resp := r.POST(`
				{
					"auth_id": "chima",
					"disabled": true,
					"message": "some reason",
					"expiry": "2017-07-23T19:30:24Z"
				}
			`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `
				{
					"result": {"status": "OK"}
				}
			`)
			So(resp.Code, ShouldEqual, 200)

			var fetchedAuth skydb.AuthInfo
			conn.GetAuth("chima", &fetchedAuth)
			So(fetchedAuth.Disabled, ShouldBeTrue)
			So(fetchedAuth.DisabledMessage, ShouldEqual, "some reason")
			So(fetchedAuth.DisabledExpiry, ShouldNotBeNil)
			So(*fetchedAuth.DisabledExpiry, ShouldResemble, time.Date(2017, 7, 23, 19, 30, 24, 0, time.UTC))
		})

		Convey("should enable a user", func() {
			resp := r.POST(`
				{
					"auth_id": "faseng",
					"disabled": false
				}
			`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `
				{
					"result": {"status": "OK"}
				}
			`)
			So(resp.Code, ShouldEqual, 200)

			var fetchedAuth skydb.AuthInfo
			conn.GetAuth("faseng", &fetchedAuth)
			So(fetchedAuth.Disabled, ShouldBeFalse)
			So(fetchedAuth.DisabledMessage, ShouldBeEmpty)
			So(fetchedAuth.DisabledExpiry, ShouldBeNil)
		})
	})
}
