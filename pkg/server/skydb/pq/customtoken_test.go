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

package pq

import (
	"database/sql"
	"testing"
	"time"

	"github.com/skygeario/skygear-server/pkg/server/skydb"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCustomTokenConn(t *testing.T) {
	var c *conn

	Convey("Conn", t, func() {
		c = getTestConn(t)
		defer cleanupConn(t, c)

		now := time.Now()
		tokenInfo := skydb.CustomTokenInfo{
			UserID:      "userid",
			PrincipalID: "faseng",
			CreatedAt:   &now,
		}

		Convey("create custom token info", func() {
			err := c.CreateCustomTokenInfo(&tokenInfo)
			So(err, ShouldBeNil)

			var userID string
			err = c.QueryRowx("SELECT user_id FROM _sso_custom_token WHERE principal_id = 'faseng'").
				Scan(&userID)
			So(err, ShouldBeNil)
			So(userID, ShouldEqual, "userid")
		})

		Convey("return ErrUserDuplicated when create duplicated custom token info", func() {
			So(c.CreateCustomTokenInfo(&tokenInfo), ShouldBeNil)
			So(c.CreateCustomTokenInfo(&tokenInfo), ShouldEqual, skydb.ErrUserDuplicated)
		})

		Convey("get ouath info", func() {
			So(c.CreateCustomTokenInfo(&tokenInfo), ShouldBeNil)

			tokenInfo := skydb.CustomTokenInfo{}
			err := c.GetCustomTokenInfo("faseng", &tokenInfo)
			So(err, ShouldBeNil)

			So(tokenInfo.UserID, ShouldResemble, "userid")
		})

		Convey("delete custom token info", func() {
			So(c.CreateCustomTokenInfo(&tokenInfo), ShouldBeNil)
			So(c.DeleteCustomTokenInfo("faseng"), ShouldBeNil)

			err := c.QueryRowx("SELECT false FROM _sso_custom_token WHERE user_id = $1", "userid").Scan()
			So(err, ShouldEqual, sql.ErrNoRows)
		})

	})
}
