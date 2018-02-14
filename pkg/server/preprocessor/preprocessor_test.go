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

package preprocessor

import (
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/server/audit"
	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skydb/mock_skydb"
	"github.com/skygeario/skygear-server/pkg/server/skydb/skydbtest"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
	. "github.com/skygeario/skygear-server/pkg/server/skytest"
)

type injectDatabasePreprocessorConn struct {
	skydb.Conn
}

func (conn *injectDatabasePreprocessorConn) PublicDB() skydb.Database {
	return &injectDatabasePreprocessorDB{
		databaseType: skydb.PublicDatabase,
		userID:       "",
	}
}

func (conn *injectDatabasePreprocessorConn) PrivateDB(userID string) skydb.Database {
	return &injectDatabasePreprocessorDB{
		databaseType: skydb.PrivateDatabase,
		userID:       userID,
	}
}

func (conn *injectDatabasePreprocessorConn) UnionDB() skydb.Database {
	return &injectDatabasePreprocessorDB{
		databaseType: skydb.UnionDatabase,
		userID:       "",
	}
}

type injectDatabasePreprocessorDB struct {
	databaseType skydb.DatabaseType
	userID       string
	skydb.Database
}

func (db *injectDatabasePreprocessorDB) DatabaseType() skydb.DatabaseType {
	return db.databaseType
}

func (db *injectDatabasePreprocessorDB) ID() string {
	return db.userID
}

func TestInjectDatabaseProcessor(t *testing.T) {
	Convey("InjectDatabase", t, func() {
		pp := InjectDatabase{}
		conn := injectDatabasePreprocessorConn{}

		Convey("should inject public DB by default", func() {
			payload := router.Payload{
				Data:   map[string]interface{}{},
				Meta:   map[string]interface{}{},
				DBConn: &conn,
			}
			resp := router.Response{}

			So(pp.Preprocess(&payload, &resp), ShouldEqual, http.StatusOK)
			So(resp.Err, ShouldBeNil)
			So(payload.Database.DatabaseType(), ShouldEqual, skydb.PublicDatabase)
		})

		Convey("should inject public DB", func() {
			payload := router.Payload{
				Data: map[string]interface{}{
					"database_id": "_public",
				},
				Meta:   map[string]interface{}{},
				DBConn: &conn,
			}
			resp := router.Response{}

			So(pp.Preprocess(&payload, &resp), ShouldEqual, http.StatusOK)
			So(resp.Err, ShouldBeNil)
			So(payload.Database.DatabaseType(), ShouldEqual, skydb.PublicDatabase)
		})

		Convey("should inject private DB", func() {
			payload := router.Payload{
				Data: map[string]interface{}{
					"database_id": "_private",
				},
				Meta: map[string]interface{}{},
				AuthInfo: &skydb.AuthInfo{
					ID: "alice",
				},
				DBConn: &conn,
			}
			resp := router.Response{}

			So(pp.Preprocess(&payload, &resp), ShouldEqual, http.StatusOK)
			So(resp.Err, ShouldBeNil)
			So(payload.Database.DatabaseType(), ShouldEqual, skydb.PrivateDatabase)
			So(payload.Database.ID(), ShouldEqual, "alice")
		})

		Convey("should not inject private DB if not logged in", func() {
			payload := router.Payload{
				Data: map[string]interface{}{
					"database_id": "_private",
				},
				Meta:   map[string]interface{}{},
				DBConn: &conn,
			}
			resp := router.Response{}

			So(pp.Preprocess(&payload, &resp), ShouldEqual, http.StatusUnauthorized)
			So(resp.Err.Code(), ShouldEqual, skyerr.NotAuthenticated)
		})

		Convey("should not inject union DB", func() {
			payload := router.Payload{
				Data: map[string]interface{}{
					"database_id": "_union",
				},
				Meta:      map[string]interface{}{},
				AccessKey: router.MasterAccessKey,
				DBConn:    &conn,
			}
			resp := router.Response{}

			So(pp.Preprocess(&payload, &resp), ShouldEqual, http.StatusOK)
			So(resp.Err, ShouldBeNil)
			So(payload.Database.DatabaseType(), ShouldEqual, skydb.UnionDatabase)
		})

		Convey("should not inject union DB if no master key", func() {
			payload := router.Payload{
				Data: map[string]interface{}{
					"database_id": "_union",
				},
				Meta:      map[string]interface{}{},
				AccessKey: router.ClientAccessKey,
				DBConn:    &conn,
			}
			resp := router.Response{}

			So(pp.Preprocess(&payload, &resp), ShouldEqual, http.StatusUnauthorized)
			So(resp.Err.Code(), ShouldEqual, skyerr.NotAuthenticated)
		})

		Convey("should inject explicit private DB", func() {
			payload := router.Payload{
				Data: map[string]interface{}{
					"database_id": "alice",
				},
				Meta: map[string]interface{}{},
				AuthInfo: &skydb.AuthInfo{
					ID: "alice",
				},
				DBConn: &conn,
			}
			resp := router.Response{}

			So(pp.Preprocess(&payload, &resp), ShouldEqual, http.StatusOK)
			So(resp.Err, ShouldBeNil)
			So(payload.Database.DatabaseType(), ShouldEqual, skydb.PrivateDatabase)
			So(payload.Database.ID(), ShouldEqual, "alice")
		})

		Convey("should inject explicit private DB if master key", func() {
			payload := router.Payload{
				Data: map[string]interface{}{
					"database_id": "alice",
				},
				Meta:      map[string]interface{}{},
				AccessKey: router.MasterAccessKey,
				DBConn:    &conn,
			}
			resp := router.Response{}

			So(pp.Preprocess(&payload, &resp), ShouldEqual, http.StatusOK)
			So(resp.Err, ShouldBeNil)
			So(payload.Database.DatabaseType(), ShouldEqual, skydb.PrivateDatabase)
			So(payload.Database.ID(), ShouldEqual, "alice")
		})
	})
}

type injectUserPreprocessorAccessToken struct {
	issuedAt time.Time
}

func (t injectUserPreprocessorAccessToken) IssuedAt() time.Time {
	return t.issuedAt
}

func TestInjectAuthProcessor(t *testing.T) {
	Convey("InjectAuth", t, func() {
		pp := InjectAuth{
			Required: true,
		}
		conn := skydbtest.NewMapConn()

		withoutTokenValidSince := skydb.AuthInfo{
			ID: "userid1",
		}
		So(conn.CreateAuth(&withoutTokenValidSince), ShouldBeNil)

		pastTime := time.Now().Add(-1 * time.Hour)
		withPastTokenValidSince := skydb.AuthInfo{
			ID:              "userid2",
			TokenValidSince: &pastTime,
		}
		So(conn.CreateAuth(&withPastTokenValidSince), ShouldBeNil)

		futureTime := time.Now().Add(1 * time.Hour)
		withFutureTokenValidSince := skydb.AuthInfo{
			ID:              "userid3",
			TokenValidSince: &futureTime,
		}
		So(conn.CreateAuth(&withFutureTokenValidSince), ShouldBeNil)

		validPasswordSince := time.Date(2017, 12, 1, 0, 0, 0, 0, time.UTC)
		validPassword := skydb.AuthInfo{
			ID:              "valid_password",
			HashedPassword:  []byte("unimportant"),
			TokenValidSince: &validPasswordSince,
		}
		So(conn.CreateAuth(&validPassword), ShouldBeNil)

		invalidPasswordSince := time.Date(2017, 10, 1, 0, 0, 0, 0, time.UTC)
		invalidPassword := skydb.AuthInfo{
			ID:              "invalid_password",
			HashedPassword:  []byte("unimportant"),
			TokenValidSince: &invalidPasswordSince,
		}
		So(conn.CreateAuth(&invalidPassword), ShouldBeNil)

		Convey("should inject user with access token", func() {
			payload := router.Payload{
				Data:        map[string]interface{}{},
				Meta:        map[string]interface{}{},
				DBConn:      conn,
				AuthInfoID:  "userid1",
				AccessToken: injectUserPreprocessorAccessToken{},
			}
			resp := router.Response{}

			So(pp.Preprocess(&payload, &resp), ShouldEqual, http.StatusOK)
			So(resp.Err, ShouldBeNil)
			So(payload.AuthInfo, ShouldResemble, &withoutTokenValidSince)
		})

		Convey("should inject user without access token", func() {
			// Note: AuthInfoID can be set by master key, hence without
			// access token.
			payload := router.Payload{
				Data:        map[string]interface{}{},
				Meta:        map[string]interface{}{},
				DBConn:      conn,
				AuthInfoID:  "userid1",
				AccessToken: nil,
			}
			resp := router.Response{}

			So(pp.Preprocess(&payload, &resp), ShouldEqual, http.StatusOK)
			So(resp.Err, ShouldBeNil)
			So(payload.AuthInfo, ShouldResemble, &withoutTokenValidSince)
		})

		Convey("should inject user with valid issued time", func() {
			payload := router.Payload{
				Data:       map[string]interface{}{},
				Meta:       map[string]interface{}{},
				DBConn:     conn,
				AuthInfoID: "userid2",
				AccessToken: injectUserPreprocessorAccessToken{
					issuedAt: time.Now(),
				},
			}
			resp := router.Response{}

			So(pp.Preprocess(&payload, &resp), ShouldEqual, http.StatusOK)
			So(resp.Err, ShouldBeNil)
			So(payload.AuthInfo, ShouldResemble, &withPastTokenValidSince)
		})

		Convey("should not inject user with invalid issued time", func() {
			payload := router.Payload{
				Data:       map[string]interface{}{},
				Meta:       map[string]interface{}{},
				DBConn:     conn,
				AuthInfoID: "userid3",
				AccessToken: injectUserPreprocessorAccessToken{
					issuedAt: time.Now(),
				},
			}
			resp := router.Response{}

			So(pp.Preprocess(&payload, &resp), ShouldEqual, http.StatusUnauthorized)
			So(resp.Err.Code(), ShouldEqual, skyerr.AccessTokenNotAccepted)
		})

		Convey("should check password expiry", func() {
			timeNow = func() time.Time { return time.Date(2017, 12, 2, 0, 0, 0, 0, time.UTC) }
			restore := skydb.MockTimeNowForTestingOnly(timeNow)
			defer restore()
			ppWithPasswordExpiryDays := InjectAuth{
				PwExpiryDays: 30,
				Required:     true,
			}

			payload1 := router.Payload{
				Data:       map[string]interface{}{},
				Meta:       map[string]interface{}{},
				DBConn:     conn,
				AuthInfoID: "valid_password",
				AccessToken: injectUserPreprocessorAccessToken{
					issuedAt: time.Now(),
				},
			}
			resp1 := router.Response{}

			So(ppWithPasswordExpiryDays.Preprocess(&payload1, &resp1), ShouldEqual, http.StatusOK)
			So(resp1.Err, ShouldBeNil)
			So(payload1.AuthInfo.ID, ShouldEqual, "valid_password")

			payload2 := router.Payload{
				Data:       map[string]interface{}{},
				Meta:       map[string]interface{}{},
				DBConn:     conn,
				AuthInfoID: "invalid_password",
				AccessToken: injectUserPreprocessorAccessToken{
					issuedAt: time.Now(),
				},
			}
			resp2 := router.Response{}

			So(ppWithPasswordExpiryDays.Preprocess(&payload2, &resp2), ShouldEqual, http.StatusUnauthorized)
			So(resp2.Err, ShouldNotBeNil)
			So(
				resp2.Err,
				ShouldEqualSkyError,
				skyerr.PasswordPolicyViolated,
				"password expired",
				map[string]interface{}{
					"reason": audit.PasswordExpired.String(),
				},
			)
		})

		Convey("should deny disabled user", func() {
			disabledUser := skydb.AuthInfo{
				ID:             "some-uuid",
				HashedPassword: []byte("unimportant"),
				Disabled:       true,
			}
			So(conn.CreateAuth(&disabledUser), ShouldBeNil)

			payload := router.Payload{
				Data:       map[string]interface{}{},
				Meta:       map[string]interface{}{},
				DBConn:     conn,
				AuthInfoID: "some-uuid",
			}
			resp := router.Response{}
			So(pp.Preprocess(&payload, &resp), ShouldEqual, http.StatusForbidden)

			So(resp.Err, ShouldNotBeNil)
			So(resp.Err.Code(), ShouldEqual, skyerr.UserDisabled)
		})

		Convey("should create and inject user when master key is used", func() {
			// Note: AuthInfoID can be set by master key, hence without
			// access token.
			payload := router.Payload{
				Data:        map[string]interface{}{},
				Meta:        map[string]interface{}{},
				DBConn:      conn,
				AuthInfoID:  "newuser",
				AccessToken: nil,
				AccessKey:   router.MasterAccessKey,
			}
			resp := router.Response{}

			So(pp.Preprocess(&payload, &resp), ShouldEqual, http.StatusOK)
			So(resp.Err, ShouldBeNil)
			So(payload.AuthInfo.ID, ShouldResemble, "newuser")

			_, ok := conn.UserMap["newuser"]
			So(ok, ShouldBeTrue)
		})

		Convey("should create _god user when master key is used and no user", func() {
			// Note: AuthInfoID can be set by master key, hence without
			// access token.
			payload := router.Payload{
				Data:        map[string]interface{}{},
				Meta:        map[string]interface{}{},
				DBConn:      conn,
				AuthInfoID:  "",
				AccessToken: nil,
				AccessKey:   router.MasterAccessKey,
			}
			resp := router.Response{}

			So(pp.Preprocess(&payload, &resp), ShouldEqual, http.StatusOK)
			So(resp.Err, ShouldBeNil)
			So(payload.AuthInfo.ID, ShouldResemble, "_god")

			_, ok := conn.UserMap["_god"]
			So(ok, ShouldBeTrue)
		})
	})
}

func TestRequireAdminOrMasterKey(t *testing.T) {
	Convey("RequireAdminOrMasterKey", t, func() {
		pp := RequireAdminOrMasterKey{}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		conn := mock_skydb.NewMockConn(ctrl)
		conn.EXPECT().GetAdminRoles().Return([]string{"admin"}, nil).AnyTimes()

		Convey("should ok with master key", func() {
			payload := router.Payload{
				DBConn:    conn,
				AccessKey: router.MasterAccessKey,
			}
			resp := router.Response{}

			So(pp.Preprocess(&payload, &resp), ShouldEqual, http.StatusOK)
			So(resp.Err, ShouldBeNil)
		})

		Convey("should ok with admin user", func() {
			payload := router.Payload{
				DBConn: conn,
				AuthInfo: &skydb.AuthInfo{
					Roles: []string{"admin"},
				},
			}
			resp := router.Response{}

			So(pp.Preprocess(&payload, &resp), ShouldEqual, http.StatusOK)
			So(resp.Err, ShouldBeNil)
		})

		Convey("should fail without user", func() {
			payload := router.Payload{
				DBConn:   conn,
				AuthInfo: &skydb.AuthInfo{},
			}
			resp := router.Response{}

			So(pp.Preprocess(&payload, &resp), ShouldEqual, http.StatusUnauthorized)
			So(resp.Err, ShouldNotBeNil)
		})

		Convey("should fail with not having admin role", func() {
			payload := router.Payload{
				DBConn: conn,
				AuthInfo: &skydb.AuthInfo{
					Roles: []string{"guest"},
				},
			}
			resp := router.Response{}

			So(pp.Preprocess(&payload, &resp), ShouldEqual, http.StatusUnauthorized)
			So(resp.Err, ShouldNotBeNil)
		})
	})
}

func TestInjectUserProcessor(t *testing.T) {
	Convey("InjectUser", t, func() {
		realTimeNow := timeNow
		timeNow = func() time.Time { return time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC) }
		defer func() {
			timeNow = realTimeNow
		}()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		db := mock_skydb.NewMockTxDatabase(ctrl)

		pp := InjectUser{}
		conn := skydbtest.NewMapConn()
		conn.InternalPublicDB = db

		authInfo := skydb.AuthInfo{
			ID: "userid1",
		}
		user := skydb.Record{
			ID: skydb.NewRecordID("user", "userid1"),
			Data: map[string]interface{}{
				"username": "john.doe",
				"email":    "john.doe@example.com",
			},
		}

		Convey("should inject user with authinfo id", func() {
			payload := router.Payload{
				Data:        map[string]interface{}{},
				Meta:        map[string]interface{}{},
				DBConn:      conn,
				Database:    db,
				AuthInfoID:  "userid1",
				AuthInfo:    &authInfo,
				AccessToken: injectUserPreprocessorAccessToken{},
				User:        nil,
			}
			resp := router.Response{}

			db.EXPECT().
				Get(skydb.NewRecordID("user", "userid1"), gomock.Any()).
				SetArg(1, user).
				Return(nil).
				AnyTimes()

			So(pp.Preprocess(&payload, &resp), ShouldEqual, http.StatusOK)
			So(resp.Err, ShouldBeNil)
			So(*payload.User, ShouldResemble, user)
		})

		Convey("should skip inject user without authinfo", func() {
			payload := router.Payload{
				Data:        map[string]interface{}{},
				Meta:        map[string]interface{}{},
				DBConn:      conn,
				Database:    db,
				AuthInfoID:  "",
				AuthInfo:    nil,
				AccessToken: injectUserPreprocessorAccessToken{},
				User:        nil,
			}
			resp := router.Response{}

			So(pp.Preprocess(&payload, &resp), ShouldEqual, http.StatusOK)
			So(resp.Err, ShouldBeNil)
			So(payload.User, ShouldEqual, nil)
		})

		Convey("should inject user with authinfo id, but no user record", func() {
			authInfo := skydb.AuthInfo{
				ID: "userid2",
			}
			payload := router.Payload{
				Data:        map[string]interface{}{},
				Meta:        map[string]interface{}{},
				DBConn:      conn,
				Database:    db,
				AuthInfoID:  "userid2",
				AuthInfo:    &authInfo,
				AccessToken: injectUserPreprocessorAccessToken{},
				User:        nil,
			}
			resp := router.Response{}

			db.EXPECT().
				Get(skydb.NewRecordID("user", "userid2"), gomock.Any()).
				Return(skydb.ErrRecordNotFound).
				AnyTimes()
			txBegin := db.EXPECT().Begin().AnyTimes()
			db.EXPECT().Commit().After(txBegin)

			db.EXPECT().UserRecordType().Return("user").AnyTimes()
			db.EXPECT().GetSchema("user").Return(skydb.RecordSchema{}, nil).AnyTimes()

			skydbtest.ExpectDBSaveUser(db, nil, func(record *skydb.Record) {
				So(record.ID.Type, ShouldEqual, "user")
				So(record.ID, ShouldResemble, skydb.NewRecordID("user", "userid2"))
			}, nil)

			So(pp.Preprocess(&payload, &resp), ShouldEqual, http.StatusOK)
			So(resp.Err, ShouldBeNil)

			user := skydb.Record{
				ID:         skydb.NewRecordID("user", "userid2"),
				DatabaseID: "_public",
				OwnerID:    "userid2",
				CreatedAt:  time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
				CreatorID:  "userid2",
				UpdatedAt:  time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
				UpdaterID:  "userid2",
			}
			So(*payload.User, ShouldResemble, user)
		})
	})
}
