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

func TestAuthRecordKeys(t *testing.T) {
	Convey("EnsureUserAuthRecordKeys", t, func() {
		c := getTestConn(t)
		defer cleanupConn(t, c)

		checkContainConstraintWithName := func(do func(bool), actual []dbIndex, expected string) {
			for _, index := range actual {
				if index.name == expected {
					do(true)
					return
				}
			}

			do(false)
		}

		shouldContainConstraintWithName := func(actual []dbIndex, expected string) {
			checkContainConstraintWithName(func(found bool) {
				So(found, ShouldBeTrue)
			}, actual, expected)
		}

		shouldNotContainConstraintWithName := func(actual []dbIndex, expected string) {
			checkContainConstraintWithName(func(found bool) {
				So(found, ShouldBeFalse)
			}, actual, expected)
		}

		Convey("canMigrate is true", func() {
			Convey("no error for default user record", func() {
				c.authRecordKeys = [][]string{[]string{"username"}, []string{"email"}}
				err := c.EnsureAuthRecordKeysValid()
				So(err, ShouldBeNil)
			})

			Convey("no error for non existing column, no new column created", func() {
				c.authRecordKeys = [][]string{[]string{"iamyourfather"}}
				err := c.EnsureAuthRecordKeysValid()
				So(err, ShouldBeNil)

				var exists bool
				c.QueryRowx(`
				SELECT EXISTS (
					SELECT 1
					FROM information_schema.columns
					WHERE table_name = 'user' AND column_name = 'iamyourfather'
				)`).Scan(&exists)
				So(exists, ShouldBeFalse)
			})

			Convey("error for existing column with invalid type", func() {
				_, err := c.PublicDB().Extend("user", skydb.RecordSchema{
					"iamyourfather": skydb.FieldType{Type: skydb.TypeJSON},
				})
				So(err, ShouldBeNil)

				c.authRecordKeys = [][]string{[]string{"iamyourfather"}}
				err = c.EnsureAuthRecordKeysValid()
				So(err, ShouldNotBeNil)
			})

			Convey("create unique constraints for existing non unique fields", func() {
				_, err := c.PublicDB().Extend("user", skydb.RecordSchema{
					"iamyourfather": skydb.FieldType{Type: skydb.TypeString},
					"iamyourmother": skydb.FieldType{Type: skydb.TypeString},
				})
				So(err, ShouldBeNil)

				c.authRecordKeys = [][]string{[]string{"iamyourfather", "iamyourmother"}}
				err = c.EnsureAuthRecordKeysValid()
				So(err, ShouldBeNil)

				So(c.PublicDB().Save(&skydb.Record{
					ID:        skydb.NewRecordID("user", "johndoe"),
					OwnerID:   "johndoe",
					CreatedAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
					CreatorID: "johndoe",
					UpdatedAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
					UpdaterID: "johndoe",
					Data: map[string]interface{}{
						"iamyourfather": "father",
						"iamyourmother": "mother",
					},
				}), ShouldBeNil)

				So(c.PublicDB().Save(&skydb.Record{
					ID:        skydb.NewRecordID("user", "john.doe"),
					OwnerID:   "john.doe",
					CreatedAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
					CreatorID: "john.doe",
					UpdatedAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
					UpdaterID: "john.doe",
					Data: map[string]interface{}{
						"iamyourfather": "father",
						"iamyourmother": "Mother",
					},
				}), ShouldBeNil)

				So(c.PublicDB().Save(&skydb.Record{
					ID:        skydb.NewRecordID("user", "john.do.e"),
					OwnerID:   "john.do.e",
					CreatedAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
					CreatorID: "john.do.e",
					UpdatedAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
					UpdaterID: "john.do.e",
					Data: map[string]interface{}{
						"iamyourfather": "father",
						"iamyourmother": "Mother",
					},
				}), ShouldNotBeNil)
			})

			Convey("remove managed unique constraints for fields no longer auth record keys", func() {
				_, err := c.PublicDB().Extend("user", skydb.RecordSchema{
					"iamyourfather": skydb.FieldType{Type: skydb.TypeString},
					"iamyourmother": skydb.FieldType{Type: skydb.TypeString},
				})
				So(err, ShouldBeNil)

				db := c.PublicDB().(*database)
				c.authRecordKeys = [][]string{[]string{"iamyourfather"}, []string{"iamyourmother"}}
				err = c.EnsureAuthRecordKeysValid()
				So(err, ShouldBeNil)
				indexes, err := db.getIndexes("user")
				So(err, ShouldBeNil)
				shouldContainConstraintWithName(indexes, "auth_record_keys_user_iamyourfather_key")
				shouldContainConstraintWithName(indexes, "auth_record_keys_user_iamyourmother_key")

				c.authRecordKeys = [][]string{[]string{"iamyourfather"}}
				err = c.EnsureAuthRecordKeysValid()
				So(err, ShouldBeNil)
				indexes, err = db.getIndexes("user")
				So(err, ShouldBeNil)
				shouldContainConstraintWithName(indexes, "auth_record_keys_user_iamyourfather_key")
				shouldNotContainConstraintWithName(indexes, "auth_record_keys_user_iamyourmother_key")
			})
		})

		Convey("canMigrate is false", func() {
			c.canMigrate = false

			Convey("no error for default user record", func() {
				c.authRecordKeys = [][]string{[]string{"username"}, []string{"email"}}
				err := c.EnsureAuthRecordKeysValid()
				So(err, ShouldBeNil)
			})

			Convey("error for non existing column", func() {
				c.authRecordKeys = [][]string{[]string{"iamyourfather"}}
				err := c.EnsureAuthRecordKeysValid()
				So(err, ShouldNotBeNil)
			})

			Convey("error for existing column with invalid type", func() {
				c.canMigrate = true
				_, err := c.PublicDB().Extend("user", skydb.RecordSchema{
					"iamyourfather": skydb.FieldType{Type: skydb.TypeJSON},
				})
				So(err, ShouldBeNil)
				c.canMigrate = false

				c.authRecordKeys = [][]string{[]string{"iamyourfather"}}
				err = c.EnsureAuthRecordKeysValid()
				So(err, ShouldNotBeNil)
			})

			Convey("error for existing column without unique constraint", func() {
				c.canMigrate = true
				_, err := c.PublicDB().Extend("user", skydb.RecordSchema{
					"iamyourfather": skydb.FieldType{Type: skydb.TypeString},
				})
				So(err, ShouldBeNil)
				c.canMigrate = false

				c.authRecordKeys = [][]string{[]string{"iamyourfather"}}
				err = c.EnsureAuthRecordKeysValid()
				So(err, ShouldNotBeNil)
			})
		})
	})
}
