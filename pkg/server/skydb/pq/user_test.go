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
	"sort"
	"testing"
	"time"

	"github.com/skygeario/skygear-server/pkg/server/skydb"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAuthCRUD(t *testing.T) {
	var c *conn

	Convey("Conn", t, func() {
		c = getTestConn(t)
		defer cleanupConn(t, c)

		mockedTime := time.Date(2017, 12, 4, 1, 2, 3, 0, time.UTC)
		originalTimeNow := timeNow
		defer func() {
			timeNow = originalTimeNow
		}()
		timeNow = func() time.Time {
			return mockedTime
		}

		expiry := mockedTime.Add(time.Hour)
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
			Disabled:        true,
			DisabledMessage: "some reason",
			DisabledExpiry:  &expiry,
			Verified:        true,
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

		Convey("creates user with password history", func() {
			tokenValidSince := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
			authInfoWithPassword := skydb.NewAuthInfo("secret")
			authInfoWithPassword.ID = "userid"
			authInfoWithPassword.TokenValidSince = &tokenValidSince
			originalEnabled := c.passwordHistoryEnabled
			defer func() {
				c.passwordHistoryEnabled = originalEnabled
			}()
			c.passwordHistoryEnabled = true
			err := c.CreateAuth(&authInfoWithPassword)
			So(err, ShouldBeNil)

			var count int
			c.QueryRowx(`
				SELECT COUNT(1)
				FROM _password_history
				WHERE auth_id = 'userid'
			`).Scan(&count)
			So(count, ShouldEqual, 1)
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
			So(count, ShouldEqual, 2) // including default admin user

			err := c.DeleteAuth("2")
			So(err, ShouldBeNil)

			c.QueryRowx("SELECT COUNT(*) FROM _auth").Scan(&count)
			So(count, ShouldEqual, 1) // including default admin user
		})
	})
}

type passwordHistoryByLoggedAt []skydb.PasswordHistory

func (a passwordHistoryByLoggedAt) Len() int      { return len(a) }
func (a passwordHistoryByLoggedAt) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a passwordHistoryByLoggedAt) Less(i, j int) bool {
	return !a[i].LoggedAt.Before(a[j].LoggedAt)
}

func TestPasswordHistoryCRUD(t *testing.T) {
	var c *conn

	Convey("Conn", t, func() {
		c = getTestConn(t)
		defer cleanupConn(t, c)

		authID := "user1"
		hashedPassword := []byte("$2a$10$RbmNb3Rw.PONA2QTcpjBg.1E00zdSI6dWTUwZi.XC0wZm9OhOEvKO")
		fixtures := []skydb.PasswordHistory{
			skydb.PasswordHistory{
				ID:             "1",
				LoggedAt:       time.Date(2017, 12, 1, 0, 0, 0, 0, time.UTC),
				AuthID:         authID,
				HashedPassword: hashedPassword,
			},
			skydb.PasswordHistory{
				ID:             "2",
				LoggedAt:       time.Date(2017, 12, 2, 0, 0, 0, 0, time.UTC),
				AuthID:         authID,
				HashedPassword: hashedPassword,
			},
			skydb.PasswordHistory{
				ID:             "3",
				LoggedAt:       time.Date(2017, 12, 3, 0, 0, 0, 0, time.UTC),
				AuthID:         authID,
				HashedPassword: hashedPassword,
			},
		}
		for _, f := range fixtures {
			_, err := c.db.NamedExec(`
				INSERT INTO _password_history
				(id, auth_id, password, logged_at) VALUES
				(:id, :auth_id, :password, :logged_at)
			`, map[string]interface{}{
				"id":        f.ID,
				"auth_id":   f.AuthID,
				"password":  f.HashedPassword,
				"logged_at": f.LoggedAt,
			})
			if err != nil {
				panic(err)
			}
		}

		queryIDs := func() []string {
			rows, err := c.db.NamedQuery(`
				SELECT id FROM _password_history
				WHERE auth_id = :auth_id
				ORDER BY logged_at DESC
			`, map[string]interface{}{
				"auth_id": authID,
			})
			if err != nil {
				panic(err)
			}
			defer rows.Close()

			ids := []string{}
			for rows.Next() {
				var id string
				if err := rows.Scan(&id); err != nil {
					panic(err)
				}
				ids = append(ids, id)
			}

			return ids
		}

		Convey("Query password history by size", func() {
			h1, err := c.GetPasswordHistory(authID, 4, 0)
			So(err, ShouldBeNil)
			So(len(h1), ShouldEqual, 3)
			So(sort.IsSorted(passwordHistoryByLoggedAt(h1)), ShouldBeTrue)

			h2, err := c.GetPasswordHistory(authID, 2, 0)
			So(err, ShouldBeNil)
			So(len(h2), ShouldEqual, 2)
			So(sort.IsSorted(passwordHistoryByLoggedAt(h2)), ShouldBeTrue)
		})

		Convey("Query password history by days", func() {
			mockedTime := time.Date(2017, 12, 4, 1, 2, 3, 0, time.UTC)
			originalTimeNow := timeNow
			defer func() {
				timeNow = originalTimeNow
			}()
			timeNow = func() time.Time {
				return mockedTime
			}

			h1, err := c.GetPasswordHistory(authID, 0, 1)
			So(err, ShouldBeNil)
			So(len(h1), ShouldEqual, 1)
			So(sort.IsSorted(passwordHistoryByLoggedAt(h1)), ShouldBeTrue)

			h2, err := c.GetPasswordHistory(authID, 0, 2)
			So(err, ShouldBeNil)
			So(len(h2), ShouldEqual, 2)
			So(sort.IsSorted(passwordHistoryByLoggedAt(h2)), ShouldBeTrue)

			h3, err := c.GetPasswordHistory(authID, 0, 10)
			So(err, ShouldBeNil)
			So(len(h3), ShouldEqual, 3)
			So(sort.IsSorted(passwordHistoryByLoggedAt(h3)), ShouldBeTrue)
		})

		Convey("Query password history by size and days", func() {
			mockedTime := time.Date(2017, 12, 4, 1, 2, 3, 0, time.UTC)
			originalTimeNow := timeNow
			defer func() {
				timeNow = originalTimeNow
			}()
			timeNow = func() time.Time {
				return mockedTime
			}

			h1, err := c.GetPasswordHistory(authID, 1, 2)
			So(err, ShouldBeNil)
			So(len(h1), ShouldEqual, 2)
			So(sort.IsSorted(passwordHistoryByLoggedAt(h1)), ShouldBeTrue)

			h2, err := c.GetPasswordHistory(authID, 2, 1)
			So(err, ShouldBeNil)
			So(len(h2), ShouldEqual, 2)
			So(sort.IsSorted(passwordHistoryByLoggedAt(h2)), ShouldBeTrue)
		})

		Convey("Remove password history", func() {
			mockedTime := time.Date(2017, 12, 4, 1, 2, 3, 0, time.UTC)
			originalTimeNow := timeNow
			defer func() {
				timeNow = originalTimeNow
			}()
			timeNow = func() time.Time {
				return mockedTime
			}

			err := c.RemovePasswordHistory(authID, 1, 0)
			So(err, ShouldBeNil)

			ids := queryIDs()
			So(len(ids), ShouldEqual, 1)
			So(ids, ShouldResemble, []string{"3"})
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

		Convey("canMigrate is true", func() {
			c.canMigrate = true

			Convey("no error for default user record", func() {
				err := c.EnsureAuthRecordKeysExist([][]string{[]string{"username"}, []string{"email"}})
				So(err, ShouldBeNil)
			})

			Convey("no error for non existing column, new column created", func() {
				err := c.EnsureAuthRecordKeysExist([][]string{[]string{"iamyourfather"}})
				So(err, ShouldBeNil)

				var exists bool
				c.QueryRowx(`
				SELECT EXISTS (
					SELECT 1
					FROM information_schema.columns
					WHERE table_name = 'user' AND column_name = 'iamyourfather'
				)`).Scan(&exists)
				So(exists, ShouldBeTrue)
			})

			Convey("no error for existing column with non string type", func() {
				_, err := c.PublicDB().Extend("user", skydb.RecordSchema{
					"iamyourfather": skydb.FieldType{Type: skydb.TypeJSON},
				})
				So(err, ShouldBeNil)

				err = c.EnsureAuthRecordKeysExist([][]string{[]string{"iamyourfather"}})
				So(err, ShouldBeNil)
			})

			Convey("create unique constraints for existing non unique fields", func() {
				_, err := c.PublicDB().Extend("user", skydb.RecordSchema{
					"iamyourfather": skydb.FieldType{Type: skydb.TypeString},
					"iamyourmother": skydb.FieldType{Type: skydb.TypeString},
				})
				So(err, ShouldBeNil)

				err = c.EnsureAuthRecordKeysIndexesMatch([][]string{[]string{"iamyourfather", "iamyourmother"}})
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
				err = c.EnsureAuthRecordKeysIndexesMatch([][]string{[]string{"iamyourfather"}, []string{"iamyourmother"}})
				So(err, ShouldBeNil)
				indexes, err := db.GetIndexesByRecordType("user")
				So(err, ShouldBeNil)
				So(indexes, ShouldContainKey, "auth_record_keys_user_iamyourfather_key")
				So(indexes, ShouldContainKey, "auth_record_keys_user_iamyourmother_key")

				err = c.EnsureAuthRecordKeysIndexesMatch([][]string{[]string{"iamyourfather"}})
				So(err, ShouldBeNil)
				indexes, err = db.GetIndexesByRecordType("user")
				So(err, ShouldBeNil)
				So(indexes, ShouldContainKey, "auth_record_keys_user_iamyourfather_key")
				So(indexes, ShouldNotContainKey, "auth_record_keys_user_iamyourmother_key")
			})
		})

		Convey("canMigrate is false", func() {
			c.canMigrate = false

			Convey("no error for default user record", func() {
				err := c.EnsureAuthRecordKeysExist([][]string{[]string{"username"}, []string{"email"}})
				So(err, ShouldBeNil)
			})

			Convey("error for non existing column", func() {
				err := c.EnsureAuthRecordKeysExist([][]string{[]string{"iamyourfather"}})
				So(err, ShouldNotBeNil)
			})

			Convey("no error for existing column with non string type", func() {
				c.canMigrate = true
				_, err := c.PublicDB().Extend("user", skydb.RecordSchema{
					"iamyourfather": skydb.FieldType{Type: skydb.TypeJSON},
				})
				So(err, ShouldBeNil)
				c.canMigrate = false

				err = c.EnsureAuthRecordKeysExist([][]string{[]string{"iamyourfather"}})
				So(err, ShouldBeNil)
			})

			Convey("error for existing column without unique constraint", func() {
				c.canMigrate = true
				_, err := c.PublicDB().Extend("user", skydb.RecordSchema{
					"iamyourfather": skydb.FieldType{Type: skydb.TypeString},
				})
				So(err, ShouldBeNil)
				c.canMigrate = false

				err = c.EnsureAuthRecordKeysIndexesMatch([][]string{[]string{"iamyourfather"}})
				So(err, ShouldNotBeNil)
			})
		})
	})
}
