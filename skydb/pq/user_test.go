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

	"github.com/skygeario/skygear-server/skydb"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/crypto/bcrypt"
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
			Roles:          []string{},
			Auth: skydb.AuthInfo{
				"com.example:johndoe": map[string]interface{}{
					"string": "string",
					"bool":   true,
					"number": float64(1),
				},
			},
		}

		Convey("default admin role", func() {
			var exists bool
			c.QueryRowx(`
				SELECT EXISTS (
					SELECT 1
					FROM app_com_oursky_skygear._role
					WHERE is_admin = TRUE
				)`).Scan(&exists)
			So(exists, ShouldBeTrue)
		})

		Convey("default admin user", func() {
			var username string
			var actualHashedPassword string

			c.QueryRowx(`
				SELECT u.username, u.password
				FROM app_com_oursky_skygear._user as u
					JOIN app_com_oursky_skygear._user_role as ur ON ur.user_id = u.id
					JOIN app_com_oursky_skygear._role as r ON ur.role_id = r.id
				WHERE r.is_admin = TRUE`,
			).Scan(&username, &actualHashedPassword)

			So(username, ShouldEqual, "admin")
			So(
				bcrypt.CompareHashAndPassword(
					[]byte(actualHashedPassword),
					[]byte("secret"),
				),
				ShouldBeNil,
			)
		})

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
			So(auth.AuthInfo, ShouldResemble, skydb.AuthInfo{
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

		Convey("get an existing User with case-preserved username and email", func() {
			userinfo := skydb.UserInfo{}
			userinfo.Username = "Capital.ONE"
			userinfo.Email = "capital.ONE@EXAMPLE.com"
			userinfo.ID = "userid-capital"
			So(c.CreateUser(&userinfo), ShouldBeNil)

			fetcheduserinfo := skydb.UserInfo{}
			err := c.GetUser("userid-capital", &fetcheduserinfo)
			So(err, ShouldBeNil)

			So(fetcheduserinfo.Username, ShouldEqual, "Capital.ONE")
			So(fetcheduserinfo.Email, ShouldEqual, "capital.ONE@EXAMPLE.com")
		})

		Convey("gets an existing User token valid since", func() {
			tokenValidSince := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
			userinfo.TokenValidSince = &tokenValidSince

			err := c.CreateUser(&userinfo)
			So(err, ShouldBeNil)

			fetcheduserinfo := skydb.UserInfo{}
			err = c.GetUser("userid", &fetcheduserinfo)
			So(err, ShouldBeNil)

			So(tokenValidSince.Equal(fetcheduserinfo.TokenValidSince.UTC()), ShouldBeTrue)
		})

		Convey("gets an existing User by username case insensitive", func() {
			err := c.CreateUser(&userinfo)
			So(err, ShouldBeNil)

			fetcheduserinfo := skydb.UserInfo{}
			err = c.GetUserByUsernameEmail("john.Doe", "", &fetcheduserinfo)
			So(err, ShouldBeNil)

			So(fetcheduserinfo, ShouldResemble, userinfo)
		})

		Convey("gets an existing User by email case insensitive", func() {
			err := c.CreateUser(&userinfo)
			So(err, ShouldBeNil)

			fetcheduserinfo := skydb.UserInfo{}
			err = c.GetUserByUsernameEmail("", "john.DOE@example.com", &fetcheduserinfo)
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

			email := ""
			err = c.QueryRowx("SELECT email FROM app_com_oursky_skygear._user WHERE id = 'userid'").
				Scan(&email)
			So(err, ShouldBeNil)
			So(email, ShouldEqual, "jane.doe@example.com")
		})

		Convey("query for empty", func() {
			userinfo.Email = ""
			err := c.CreateUser(&userinfo)
			So(err, ShouldBeNil)

			emails := []string{""}
			results, err := c.QueryUser(emails, []string{})
			So(err, ShouldBeNil)
			So(len(results), ShouldEqual, 0)
		})

		Convey("query for users with email", func() {
			err := c.CreateUser(&userinfo)
			So(err, ShouldBeNil)

			userinfo.Username = "jane.doe"
			userinfo.Email = "jane.doe@example.com"
			userinfo.ID = "userid2"
			So(c.CreateUser(&userinfo), ShouldBeNil)

			emails := []string{"john.doe@example.com", "jane.doe@example.com"}
			results, err := c.QueryUser(emails, []string{})
			So(err, ShouldBeNil)

			userids := []string{}
			for _, userinfo := range results {
				userids = append(userids, userinfo.ID)
			}
			So(userids, ShouldContain, "userid")
			So(userids, ShouldContain, "userid2")
		})

		Convey("query for users with email and username", func() {
			err := c.CreateUser(&userinfo)
			So(err, ShouldBeNil)

			userinfo.Username = "jane.doe"
			userinfo.Email = "jane.doe@example.com"
			userinfo.ID = "userid2"
			So(c.CreateUser(&userinfo), ShouldBeNil)

			emails := []string{"john.doe@example.com"}
			usernames := []string{"jane.doe"}
			results, err := c.QueryUser(emails, usernames)
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
			So(count, ShouldEqual, 3) // including default admin user

			err = c.DeleteUser("2")
			So(err, ShouldBeNil)

			c.QueryRowx("SELECT COUNT(*) FROM app_com_oursky_skygear._user").Scan(&count)
			So(count, ShouldEqual, 2) // including default admin user
		})
	})
}

func TestUserEagerLoadRole(t *testing.T) {
	var c *conn

	Convey("User eager load roles", t, func() {
		c = getTestConn(t)
		defer cleanupConn(t, c)

		userinfo := skydb.UserInfo{
			ID:             "userid",
			Username:       "john.doe",
			Email:          "john.doe@example.com",
			Roles:          []string{"user"},
			HashedPassword: []byte(""),
		}
		c.CreateUser(&userinfo)

		Convey("with GetUser", func() {
			fetchedUserinfo := skydb.UserInfo{}
			So(c.GetUser("userid", &fetchedUserinfo), ShouldBeNil)
			So(fetchedUserinfo, ShouldResemble, userinfo)
		})

		Convey("with UserQuery", func() {
			results, err := c.QueryUser([]string{
				"john.doe@example.com",
			}, []string{})
			So(err, ShouldBeNil)

			So(results[0], ShouldResemble, userinfo)
		})
	})
}
