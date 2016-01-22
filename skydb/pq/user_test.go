package pq

import (
	"database/sql"
	"testing"

	"github.com/oursky/skygear/skydb"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUserCRUD(t *testing.T) {
	var c *conn

	Convey("Conn", t, func() {
		c = getTestConn(t)
		defer cleanupConn(t, c)

		userinfo := skydb.UserInfo{
			ID:             "userid",
			Username:       "john.doe",
			Email:          "john.doe@example.com",
			HashedPassword: []byte("$2a$10$RbmNb3Rw.PONA2QTcpjBg.1E00zdSI6dWTUwZi.XC0wZm9OhOEvKO"),
			Auth: skydb.AuthInfo{
				"com.example:johndoe": map[string]interface{}{
					"string": "string",
					"bool":   true,
					"number": float64(1),
				},
			},
		}

		Convey("creates user", func() {
			err := c.CreateUser(&userinfo)
			So(err, ShouldBeNil)

			email := ""
			password := []byte{}
			auth := authInfoValue{}
			err = c.QueryRowx("SELECT email, password, auth FROM app_com_oursky_skygear._user WHERE id = 'userid'").
				Scan(&email, &password, &auth)
			So(err, ShouldBeNil)

			So(email, ShouldEqual, "john.doe@example.com")
			So(password, ShouldResemble, []byte("$2a$10$RbmNb3Rw.PONA2QTcpjBg.1E00zdSI6dWTUwZi.XC0wZm9OhOEvKO"))
			So(auth, ShouldResemble, authInfoValue{
				"com.example:johndoe": map[string]interface{}{
					"string": "string",
					"bool":   true,
					"number": float64(1),
				},
			})
		})

		Convey("returns ErrUserDuplicated when user to create already exists", func() {
			err := c.CreateUser(&userinfo)
			So(err, ShouldBeNil)

			err = c.CreateUser(&userinfo)
			So(err, ShouldEqual, skydb.ErrUserDuplicated)
		})

		Convey("returns ErrUserDuplicated when user with same username", func() {
			err := c.CreateUser(&userinfo)
			So(err, ShouldBeNil)

			err = c.CreateUser(&skydb.UserInfo{
				Username:       "john.doe",
				HashedPassword: []byte("$2a$10$RbmNb3Rw.PONA2QTcpjBg.1E00zdSI6dWTUwZi.XC0wZm9OhOEvKO"),
			})
			So(err, ShouldEqual, skydb.ErrUserDuplicated)
		})

		Convey("returns ErrUserDuplicated when user with same email", func() {
			err := c.CreateUser(&userinfo)
			So(err, ShouldBeNil)

			err = c.CreateUser(&skydb.UserInfo{
				Email:          "john.doe@example.com",
				HashedPassword: []byte("$2a$10$RbmNb3Rw.PONA2QTcpjBg.1E00zdSI6dWTUwZi.XC0wZm9OhOEvKO"),
			})
			So(err, ShouldEqual, skydb.ErrUserDuplicated)
		})

		Convey("gets an existing User", func() {
			err := c.CreateUser(&userinfo)
			So(err, ShouldBeNil)

			fetcheduserinfo := skydb.UserInfo{}
			err = c.GetUser("userid", &fetcheduserinfo)
			So(err, ShouldBeNil)

			So(fetcheduserinfo, ShouldResemble, userinfo)
		})

		Convey("gets an existing User by principal", func() {
			err := c.CreateUser(&userinfo)
			So(err, ShouldBeNil)

			fetcheduserinfo := skydb.UserInfo{}
			err = c.GetUserByPrincipalID("com.example:johndoe", &fetcheduserinfo)
			So(err, ShouldBeNil)

			So(fetcheduserinfo, ShouldResemble, userinfo)
		})

		Convey("returns ErrUserNotFound when the user does not exist", func() {
			err := c.GetUser("userid", (*skydb.UserInfo)(nil))
			So(err, ShouldEqual, skydb.ErrUserNotFound)
		})

		Convey("returns ErrUserNotFound when the user does not exist by principal", func() {
			err := c.GetUserByPrincipalID("com.example:janedoe", (*skydb.UserInfo)(nil))
			So(err, ShouldEqual, skydb.ErrUserNotFound)
		})

		Convey("updates a user", func() {
			err := c.CreateUser(&userinfo)
			So(err, ShouldBeNil)

			userinfo.Email = "jane.doe@example.com"

			err = c.UpdateUser(&userinfo)
			So(err, ShouldBeNil)

			updateduserinfo := userInfo{}
			err = c.Get(&updateduserinfo, "SELECT id, email, password, auth FROM app_com_oursky_skygear._user WHERE id = $1", "userid")
			So(err, ShouldBeNil)
			So(updateduserinfo, ShouldResemble, userInfo{
				ID:             "userid",
				Email:          "jane.doe@example.com",
				HashedPassword: []byte("$2a$10$RbmNb3Rw.PONA2QTcpjBg.1E00zdSI6dWTUwZi.XC0wZm9OhOEvKO"),
				Auth: authInfoValue{
					"com.example:johndoe": map[string]interface{}{
						"string": "string",
						"bool":   true,
						"number": float64(1),
					},
				},
			})
		})

		Convey("query for empty", func() {
			userinfo.Email = ""
			err := c.CreateUser(&userinfo)
			So(err, ShouldBeNil)

			emails := []string{""}
			results, err := c.QueryUser(emails)
			So(err, ShouldBeNil)
			So(len(results), ShouldEqual, 0)
		})

		Convey("query for users", func() {
			err := c.CreateUser(&userinfo)
			So(err, ShouldBeNil)

			userinfo.Username = "jane.doe"
			userinfo.Email = "jane.doe@example.com"
			userinfo.ID = "userid2"
			So(c.CreateUser(&userinfo), ShouldBeNil)

			emails := []string{"john.doe@example.com", "jane.doe@example.com"}
			results, err := c.QueryUser(emails)
			So(err, ShouldBeNil)

			userids := []string{}
			for _, userinfo := range results {
				userids = append(userids, userinfo.ID)
			}
			So(userids, ShouldContain, "userid")
			So(userids, ShouldContain, "userid2")
		})

		Convey("returns ErrUserNotFound when the user to update does not exist", func() {
			err := c.UpdateUser(&userinfo)
			So(err, ShouldEqual, skydb.ErrUserNotFound)
		})

		Convey("deletes an existing user", func() {
			err := c.CreateUser(&userinfo)
			So(err, ShouldBeNil)

			err = c.DeleteUser("userid")
			So(err, ShouldBeNil)

			placeholder := []byte{}
			err = c.QueryRowx("SELECT false FROM app_com_oursky_skygear._user WHERE id = $1", "userid").Scan(&placeholder)
			So(err, ShouldEqual, sql.ErrNoRows)
			So(placeholder, ShouldBeEmpty)
		})

		Convey("returns ErrUserNotFound when the user to delete does not exist", func() {
			err := c.DeleteUser("notexistid")
			So(err, ShouldEqual, skydb.ErrUserNotFound)
		})

		Convey("deletes only the desired user", func() {
			userinfo.ID = "1"
			userinfo.Username = "user1"
			userinfo.Email = "user1@skygear.com"
			err := c.CreateUser(&userinfo)
			So(err, ShouldBeNil)

			userinfo.ID = "2"
			userinfo.Username = "user2"
			userinfo.Email = "user2@skygear.com"
			err = c.CreateUser(&userinfo)
			So(err, ShouldBeNil)

			count := 0
			c.QueryRowx("SELECT COUNT(*) FROM app_com_oursky_skygear._user").Scan(&count)
			So(count, ShouldEqual, 2)

			err = c.DeleteUser("2")
			So(err, ShouldBeNil)

			c.QueryRowx("SELECT COUNT(*) FROM app_com_oursky_skygear._user").Scan(&count)
			So(count, ShouldEqual, 1)
		})
	})
}
