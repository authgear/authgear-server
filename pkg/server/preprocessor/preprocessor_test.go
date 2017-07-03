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

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skydb/skydbtest"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
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
		pp := InjectAuthIfPresent{}
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
