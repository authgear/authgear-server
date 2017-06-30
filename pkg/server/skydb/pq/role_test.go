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
			authinfo := skydb.AuthInfo{
				ID: "userid",
			}
			err := c.CreateAuth(&authinfo)
			So(err, ShouldBeNil)
			authinfo.Roles = []string{
				"admin",
				"writer",
			}
			err = c.UpdateUserRoles(&authinfo)
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

			rows, err := c.Queryx("SELECT role_id FROM _auth_role WHERE auth_id = 'userid'")
			So(err, ShouldBeNil)
			roles := []string{}
			for rows.Next() {
				rows.Scan(&role)
				roles = append(roles, role)
			}
			So(roles, ShouldResemble, authinfo.Roles)
		})

		Convey("clear roles of a user keep the role definition", func() {
			authinfo := skydb.AuthInfo{
				ID: "userid",
				Roles: []string{
					"admin",
					"writer",
				},
			}
			err := c.CreateAuth(&authinfo)
			So(err, ShouldBeNil)
			authinfo.Roles = nil
			err = c.UpdateUserRoles(&authinfo)
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

			rows, err := c.Queryx("SELECT role_id FROM _auth_role WHERE auth_id = 'userid'")
			So(err, ShouldBeNil)
			So(rows.Next(), ShouldBeFalse)
		})
	})
}

func TestRoleAssignRevoke(t *testing.T) {
	var c *conn
	Convey("AssignRoles", t, func() {
		c = getTestConn(t)
		defer cleanupConn(t, c)

		Convey("assign roles to user without roles", func() {
			authinfo := skydb.AuthInfo{
				ID: "userid",
			}
			err := c.CreateAuth(&authinfo)
			roles := []string{
				"admin",
				"user",
			}
			err = c.AssignRoles([]string{
				"userid",
			}, roles)
			rows, err := c.Queryx("SELECT role_id FROM _auth_role WHERE auth_id = 'userid'")
			So(err, ShouldBeNil)
			result := []string{}
			var role string
			for rows.Next() {
				rows.Scan(&role)
				result = append(result, role)
			}
			So(result, ShouldResemble, roles)
		})

		Convey("assign roles to users with existing roles", func() {
			authinfo := skydb.AuthInfo{
				ID: "userid",
				Roles: []string{
					"admin",
				},
			}
			err := c.CreateAuth(&authinfo)
			So(err, ShouldBeNil)
			authinfo = skydb.AuthInfo{
				ID: "userid2",
				Roles: []string{
					"user",
				},
			}
			err = c.CreateAuth(&authinfo)
			So(err, ShouldBeNil)

			roles := []string{
				"admin",
				"user",
			}
			err = c.AssignRoles([]string{
				"userid",
				"userid2",
			}, roles)
			So(err, ShouldBeNil)
			rows, err := c.Queryx(
				"SELECT * FROM _auth_role WHERE auth_id IN ( 'userid', 'userid2' )")
			So(err, ShouldBeNil)
			count := 0
			for rows.Next() {
				count++
			}
			So(count, ShouldEqual, 4)
		})
	})

	Convey("RevokeRoles", t, func() {
		c = getTestConn(t)
		defer cleanupConn(t, c)
		Convey("revoke roles from users with a role", func() {
			authinfo := skydb.AuthInfo{
				ID: "userid",
				Roles: []string{
					"admin",
					"user",
				},
			}
			err := c.CreateAuth(&authinfo)
			authinfo = skydb.AuthInfo{
				ID: "userid2",
				Roles: []string{
					"user",
				},
			}

			roles := []string{
				"admin",
				"user",
			}
			err = c.RevokeRoles([]string{
				"userid",
				"userid2",
			}, roles)
			rows, err := c.Queryx(
				"SELECT role_id FROM _auth_role WHERE auth_id IN ( 'userid', 'userid2' )")
			So(err, ShouldBeNil)
			So(rows.Next(), ShouldBeFalse)
		})

		Convey("revoke roles from users without a role", func() {
			authinfo := skydb.AuthInfo{
				ID: "userid",
			}
			err := c.CreateAuth(&authinfo)
			authinfo = skydb.AuthInfo{
				ID: "userid2",
			}

			roles := []string{
				"admin",
				"user",
			}
			err = c.RevokeRoles([]string{
				"userid",
				"userid2",
			}, roles)
			rows, err := c.Queryx(
				"SELECT role_id FROM _auth_role WHERE auth_id IN ( 'userid', 'userid2' )")
			So(err, ShouldBeNil)
			So(rows.Next(), ShouldBeFalse)
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
			var role string
			roles := []string{}
			for rows.Next() {
				rows.Scan(&role)
				roles = append(roles, role)
			}
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
			var role string
			roles := []string{}
			for rows.Next() {
				rows.Scan(&role)
				roles = append(roles, role)
			}
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
