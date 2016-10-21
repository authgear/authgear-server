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
	"testing"

	"github.com/skygeario/skygear-server/pkg/server/skydb"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRoleCRUD(t *testing.T) {
	var c *conn

	Convey("Conn", t, func() {
		c = getTestConn(t)
		defer cleanupConn(t, c)

		Convey("add roles to a user", func() {
			userinfo := skydb.UserInfo{
				ID:       "userid",
				Username: "john.doe",
				Email:    "john.doe@example.com",
			}
			err := c.CreateUser(&userinfo)
			So(err, ShouldBeNil)
			userinfo.Roles = []string{
				"admin",
				"writer",
			}
			err = c.UpdateUserRoles(&userinfo)
			So(err, ShouldBeNil)

			var role string
			err = c.QueryRowx("SELECT id FROM _role WHERE id = 'admin'").
				Scan(&role)
			So(err, ShouldBeNil)
			So(role, ShouldEqual, "admin")

			err = c.QueryRowx("SELECT id FROM _role WHERE id = 'writer'").
				Scan(&role)
			So(err, ShouldBeNil)
			So(role, ShouldEqual, "writer")

			rows, err := c.Queryx("SELECT role_id FROM _user_role WHERE user_id = 'userid'")
			So(err, ShouldBeNil)
			c := 0
			roles := []string{}
			for rows.Next() {
				c++
				rows.Scan(&role)
				roles = append(roles, role)
			}
			So(c, ShouldEqual, 2)
			So(roles, ShouldResemble, userinfo.Roles)
		})

		Convey("clear roles of a user keep the role definition", func() {
			userinfo := skydb.UserInfo{
				ID: "userid",
				Roles: []string{
					"admin",
					"writer",
				},
			}
			err := c.CreateUser(&userinfo)
			So(err, ShouldBeNil)
			userinfo.Roles = nil
			err = c.UpdateUserRoles(&userinfo)
			So(err, ShouldBeNil)

			var role string
			err = c.QueryRowx("SELECT id FROM _role WHERE id = 'admin'").
				Scan(&role)
			So(err, ShouldBeNil)
			So(role, ShouldEqual, "admin")

			err = c.QueryRowx("SELECT id FROM _role WHERE id = 'writer'").
				Scan(&role)
			So(err, ShouldBeNil)
			So(role, ShouldEqual, "writer")

			rows, err := c.Queryx("SELECT role_id FROM _user_role WHERE user_id = 'userid'")
			So(err, ShouldBeNil)
			c := 0
			for rows.Next() {
				c++
			}
			So(c, ShouldEqual, 0)
		})

	})
}

func TestRoleType(t *testing.T) {
	var c *conn

	Convey("SetAdminRoles", t, func() {
		c = getTestConn(t)
		defer cleanupConn(t, c)

		Convey("set role is_admin to true", func() {
			err := c.SetAdminRoles([]string{
				"god",
				"buddha",
			})
			So(err, ShouldBeNil)
			rows, err := c.Queryx("SELECT id FROM _role WHERE is_admin = TRUE")
			So(err, ShouldBeNil)
			c := 0
			var role string
			roles := []string{}
			for rows.Next() {
				c++
				rows.Scan(&role)
				roles = append(roles, role)
			}
			So(c, ShouldEqual, 2)
			So(roles, ShouldResemble, []string{
				"god",
				"buddha",
			})
		})

		Convey("get all admin roles", func() {
			So(c.SetAdminRoles([]string{"god", "buddha"}), ShouldBeNil)
			adminRoles, err := c.GetAdminRoles()
			So(err, ShouldBeNil)
			So(adminRoles, ShouldResemble, []string{
				"god",
				"buddha",
			})
		})

		Convey("reset role is_admin to false on new admin role set", func() {
			err := c.SetAdminRoles([]string{
				"god",
				"buddha",
			})
			So(err, ShouldBeNil)
			err = c.SetAdminRoles([]string{
				"man",
			})
			So(err, ShouldBeNil)

			var role string
			err = c.QueryRowx("SELECT id FROM _role WHERE is_admin = TRUE").
				Scan(&role)
			So(err, ShouldBeNil)
			So(role, ShouldEqual, "man")
		})
	})

	Convey("SetDefaultRoles", t, func() {
		c = getTestConn(t)
		defer cleanupConn(t, c)

		Convey("set role by_default to true", func() {
			err := c.SetDefaultRoles([]string{
				"human",
				"chinese",
			})
			So(err, ShouldBeNil)
			rows, err := c.Queryx("SELECT id FROM _role WHERE by_default = TRUE")
			So(err, ShouldBeNil)
			c := 0
			var role string
			roles := []string{}
			for rows.Next() {
				c++
				rows.Scan(&role)
				roles = append(roles, role)
			}
			So(c, ShouldEqual, 2)
			So(roles, ShouldResemble, []string{
				"human",
				"chinese",
			})
		})

		Convey("get all admin roles", func() {
			So(c.SetDefaultRoles([]string{"human", "chinese"}), ShouldBeNil)
			defaultRoles, err := c.GetDefaultRoles()
			So(err, ShouldBeNil)
			So(defaultRoles, ShouldResemble, []string{
				"human",
				"chinese",
			})
		})

		Convey("reset role by_default to false on new default role set", func() {
			err := c.SetDefaultRoles([]string{
				"human",
				"chinese",
			})
			So(err, ShouldBeNil)
			err = c.SetDefaultRoles([]string{
				"free man",
			})
			So(err, ShouldBeNil)

			var role string
			err = c.QueryRowx("SELECT id FROM _role WHERE by_default = TRUE").
				Scan(&role)
			So(err, ShouldBeNil)
			So(role, ShouldEqual, "free man")
		})
	})
}
