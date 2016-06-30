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

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/router"
	"github.com/skygeario/skygear-server/skydb"
	"github.com/skygeario/skygear-server/skyerr"
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
				UserInfo: &skydb.UserInfo{
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
				UserInfo: &skydb.UserInfo{
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
