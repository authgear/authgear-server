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
	"golang.org/x/crypto/bcrypt"
)

func TestAuthCRUD(t *testing.T) {
	var c *conn

	Convey("Conn", t, func() {
		c = getTestConn(t)
		defer cleanupConn(t, c)

		authinfo := skydb.AuthInfo{
			ID:             "userid",
			HashedPassword: []byte("$2a$10$RbmNb3Rw.PONA2QTcpjBg.1E00zdSI6dWTUwZi.XC0wZm9OhOEvKO"),
			Roles:          []string{},
			ProviderInfo: skydb.ProviderInfo{
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
					FROM _role
					WHERE is_admin = TRUE
				)`).Scan(&exists)
			So(exists, ShouldBeTrue)
		})

		Convey("default admin user", func() {
			var actualHashedPassword string

			c.QueryRowx(`
				SELECT u.password
				FROM _auth as u
					JOIN _auth_role as ur ON ur.auth_id = u.id
					JOIN _role as r ON ur.role_id = r.id
				WHERE r.is_admin = TRUE`,
			).Scan(&actualHashedPassword)

			So(
				bcrypt.CompareHashAndPassword(
					[]byte(actualHashedPassword),
					[]byte("secret"),
				),
				ShouldBeNil,
			)
		})

		Convey("creates user", func() {
			err := c.CreateAuth(&authinfo)
			So(err, ShouldBeNil)

			password := []byte{}
			providerInfo := providerInfoValue{}
			err = c.QueryRowx("SELECT password, provider_info FROM _auth WHERE id = 'userid'").
				Scan(&password, &providerInfo)
			So(err, ShouldBeNil)
			So(password, ShouldResemble, []byte("$2a$10$RbmNb3Rw.PONA2QTcpjBg.1E00zdSI6dWTUwZi.XC0wZm9OhOEvKO"))
			So(providerInfo.ProviderInfo, ShouldResemble, skydb.ProviderInfo{
				"com.example:johndoe": map[string]interface{}{
					"string": "string",
					"bool":   true,
					"number": float64(1),
				},
			})
		})

		Convey("returns ErrUserDuplicated when user to create already exists", func() {
			So(c.CreateAuth(&authinfo), ShouldBeNil)
			So(c.CreateAuth(&authinfo), ShouldEqual, skydb.ErrUserDuplicated)
		})

		Convey("gets an existing User", func() {
			err := c.CreateAuth(&authinfo)
			So(err, ShouldBeNil)

			fetchedauthinfo := skydb.AuthInfo{}
			err = c.GetAuth("userid", &fetchedauthinfo)
			So(err, ShouldBeNil)

			So(fetchedauthinfo, ShouldResemble, authinfo)
		})

		Convey("gets an existing User token valid since", func() {
			tokenValidSince := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
			authinfo.TokenValidSince = &tokenValidSince

			err := c.CreateAuth(&authinfo)
			So(err, ShouldBeNil)

			fetchedauthinfo := skydb.AuthInfo{}
			err = c.GetAuth("userid", &fetchedauthinfo)
			So(err, ShouldBeNil)

			So(tokenValidSince.Equal(fetchedauthinfo.TokenValidSince.UTC()), ShouldBeTrue)
		})

		Convey("gets an existing User last login at", func() {
			lastLoginAt := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
			authinfo.LastLoginAt = &lastLoginAt

			err := c.CreateAuth(&authinfo)
			So(err, ShouldBeNil)

			fetchedauthinfo := skydb.AuthInfo{}
			err = c.GetAuth("userid", &fetchedauthinfo)
			So(err, ShouldBeNil)

			So(lastLoginAt.Equal(fetchedauthinfo.LastLoginAt.UTC()), ShouldBeTrue)
		})

		Convey("gets an existing User last seen at", func() {
			lastSeenAt := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
			authinfo.LastSeenAt = &lastSeenAt

			err := c.CreateAuth(&authinfo)
			So(err, ShouldBeNil)

			fetchedauthinfo := skydb.AuthInfo{}
			err = c.GetAuth("userid", &fetchedauthinfo)
			So(err, ShouldBeNil)

			So(
				lastSeenAt.Equal(fetchedauthinfo.LastSeenAt.UTC()),
				ShouldBeTrue,
			)
		})

		Convey("gets an existing User by principal", func() {
			err := c.CreateAuth(&authinfo)
			So(err, ShouldBeNil)

			fetchedauthinfo := skydb.AuthInfo{}
			err = c.GetAuthByPrincipalID("com.example:johndoe", &fetchedauthinfo)
			So(err, ShouldBeNil)

			So(fetchedauthinfo, ShouldResemble, authinfo)
		})

		Convey("returns ErrUserNotFound when the user does not exist", func() {
			err := c.GetAuth("userid", (*skydb.AuthInfo)(nil))
			So(err, ShouldEqual, skydb.ErrUserNotFound)
		})

		Convey("returns ErrUserNotFound when the user does not exist by principal", func() {
			err := c.GetAuthByPrincipalID("com.example:janedoe", (*skydb.AuthInfo)(nil))
			So(err, ShouldEqual, skydb.ErrUserNotFound)
		})

		Convey("updates a user", func() {
			err := c.CreateAuth(&authinfo)
			So(err, ShouldBeNil)

			authinfo.HashedPassword = []byte("newsecret")

			err = c.UpdateAuth(&authinfo)
			So(err, ShouldBeNil)

			hashedPassword := []byte("")
			err = c.QueryRowx("SELECT password FROM _auth WHERE id = 'userid'").
				Scan(&hashedPassword)
			So(err, ShouldBeNil)
			So(hashedPassword, ShouldResemble, []byte("newsecret"))
		})

		Convey("returns ErrUserNotFound when the user to update does not exist", func() {
			err := c.UpdateAuth(&authinfo)
			So(err, ShouldEqual, skydb.ErrUserNotFound)
		})

		Convey("deletes an existing user", func() {
			So(c.CreateAuth(&authinfo), ShouldBeNil)
			So(c.DeleteAuth("userid"), ShouldBeNil)

			placeholder := []byte{}
			err := c.QueryRowx("SELECT false FROM _auth WHERE id = $1", "userid").Scan(&placeholder)
			So(err, ShouldEqual, sql.ErrNoRows)
			So(placeholder, ShouldBeEmpty)

			err = c.QueryRowx(`SELECT false FROM "user" WHERE _id = $1`, "userid").Scan(&placeholder)
			So(err, ShouldEqual, sql.ErrNoRows)
			So(placeholder, ShouldBeEmpty)
		})

		Convey("returns ErrUserNotFound when the user to delete does not exist", func() {
			So(c.DeleteAuth("notexistid"), ShouldEqual, skydb.ErrUserNotFound)
		})

		Convey("deletes only the desired user", func() {
			authinfo.ID = "1"
			So(c.CreateAuth(&authinfo), ShouldBeNil)

			authinfo.ID = "2"
			So(c.CreateAuth(&authinfo), ShouldBeNil)

			count := 0
			c.QueryRowx("SELECT COUNT(*) FROM _auth").Scan(&count)
			So(count, ShouldEqual, 3) // including default admin user

			err := c.DeleteAuth("2")
			So(err, ShouldBeNil)

			c.QueryRowx("SELECT COUNT(*) FROM _auth").Scan(&count)
			So(count, ShouldEqual, 2) // including default admin user
		})
	})
}

func TestAuthEagerLoadRole(t *testing.T) {
	var c *conn

	Convey("User eager load roles", t, func() {
		c = getTestConn(t)
		defer cleanupConn(t, c)

		authinfo := skydb.AuthInfo{
			ID:             "userid",
			Roles:          []string{"user"},
			HashedPassword: []byte(""),
		}
		c.CreateAuth(&authinfo)

		Convey("with GetUser", func() {
			fetchedAuthinfo := skydb.AuthInfo{}
			So(c.GetAuth("userid", &fetchedAuthinfo), ShouldBeNil)
			So(fetchedAuthinfo, ShouldResemble, authinfo)
		})
	})
}
