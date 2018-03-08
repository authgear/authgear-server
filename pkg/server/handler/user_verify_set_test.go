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

	"github.com/skygeario/skygear-server/pkg/server/handler/handlertest"
	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	. "github.com/skygeario/skygear-server/pkg/server/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSetVerifyUserHandler(t *testing.T) {
	Convey("SetVerifyUserHandler", t, func() {
		conn := singleUserConn{}
		authinfo1 := skydb.NewAuthInfo("chima")
		authinfo1.ID = "chima"
		authinfo2 := skydb.NewAuthInfo("faseng")
		authinfo2.ID = "faseng"
		authinfo2.Verified = true
		conn.CreateAuth(&authinfo1)
		conn.CreateAuth(&authinfo2)

		r := handlertest.NewSingleRouteRouter(&SetVerifyUserHandler{}, func(p *router.Payload) {
			p.DBConn = &conn
		})

		Convey("should verify a user", func() {
			resp := r.POST(`
				{
					"auth_id": "chima",
					"verified": true
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
			So(fetchedAuth.Verified, ShouldBeTrue)
		})

		Convey("should disable a user with message and expiry", func() {
			resp := r.POST(`
				{
					"auth_id": "faseng",
					"verified": false
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
			So(fetchedAuth.Verified, ShouldBeFalse)
		})
	})
}
