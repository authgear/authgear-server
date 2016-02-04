package pq

import (
	"testing"

	"github.com/oursky/skygear/skydb"
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
			err = c.QueryRowx("SELECT id FROM app_com_oursky_skygear._role WHERE id = 'admin'").
				Scan(&role)
			So(err, ShouldBeNil)
			So(role, ShouldEqual, "admin")

			err = c.QueryRowx("SELECT id FROM app_com_oursky_skygear._role WHERE id = 'writer'").
				Scan(&role)
			So(err, ShouldBeNil)
			So(role, ShouldEqual, "writer")

			rows, err := c.Queryx("SELECT role_id FROM app_com_oursky_skygear._user_role WHERE user_id = 'userid'")
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
			err = c.QueryRowx("SELECT id FROM app_com_oursky_skygear._role WHERE id = 'admin'").
				Scan(&role)
			So(err, ShouldBeNil)
			So(role, ShouldEqual, "admin")

			err = c.QueryRowx("SELECT id FROM app_com_oursky_skygear._role WHERE id = 'writer'").
				Scan(&role)
			So(err, ShouldBeNil)
			So(role, ShouldEqual, "writer")

			rows, err := c.Queryx("SELECT role_id FROM app_com_oursky_skygear._user_role WHERE user_id = 'userid'")
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
			rows, err := c.Queryx("SELECT id FROM app_com_oursky_skygear._role WHERE is_admin = TRUE")
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
			err = c.QueryRowx("SELECT id FROM app_com_oursky_skygear._role WHERE is_admin = TRUE").
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
			rows, err := c.Queryx("SELECT id FROM app_com_oursky_skygear._role WHERE by_default = TRUE")
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
			err = c.QueryRowx("SELECT id FROM app_com_oursky_skygear._role WHERE by_default = TRUE").
				Scan(&role)
			So(err, ShouldBeNil)
			So(role, ShouldEqual, "free man")
		})
	})
}
