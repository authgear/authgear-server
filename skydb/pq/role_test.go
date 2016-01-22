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
