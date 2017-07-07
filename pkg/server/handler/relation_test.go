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

package handler

import (
	"sort"

	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/skygeario/skygear-server/pkg/server/skytest"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/server/handler/handlertest"
	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skydb/mock_skydb"
)

type testRelationConn struct {
	AuthInfo     []skydb.AuthInfo
	RelationName string
	addedID      string
	removeID     string
	addErr       error
	removeErr    error
	skydb.Conn
}

func (conn *testRelationConn) GetAuth(id string, authinfo *skydb.AuthInfo) error {
	authinfo.ID = id
	return nil
}

type sortableAuthInfo []skydb.AuthInfo

func (a sortableAuthInfo) Len() int           { return len(a) }
func (a sortableAuthInfo) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a sortableAuthInfo) Less(i, j int) bool { return a[i].ID < a[j].ID }

func (conn *testRelationConn) QueryRelation(user string, name string, direction string, config skydb.QueryConfig) []skydb.AuthInfo {
	conn.RelationName = name
	if conn.AuthInfo == nil {
		return []skydb.AuthInfo{}
	}

	sort.Sort(sortableAuthInfo(conn.AuthInfo))
	if config.Limit == 0 {
		return conn.AuthInfo[config.Offset:]
	}

	if config.Offset+config.Limit-1 > uint64(len(conn.AuthInfo)) {
		return conn.AuthInfo[config.Offset:]
	}
	return conn.AuthInfo[config.Offset : config.Offset+config.Limit]
}

func (conn *testRelationConn) AddRelation(user string, name string, targetUser string) error {
	conn.RelationName = name
	conn.addedID = targetUser
	return nil
}

func (conn *testRelationConn) RemoveRelation(user string, name string, targetUser string) error {
	conn.RelationName = name
	conn.removeID = targetUser
	return nil
}

func (conn *testRelationConn) QueryRelationCount(user string, name string, direction string) (uint64, error) {
	conn.RelationName = name
	if conn.AuthInfo == nil {
		return 0, nil
	}
	count := uint64(len(conn.AuthInfo))
	return count, nil
}

func TestRelationHandler(t *testing.T) {
	Convey("RelationAddHandler", t, func() {
		conn := testRelationConn{}

		Convey("add new relation", func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			db := mock_skydb.NewMockTxDatabase(ctrl)
			db.EXPECT().
				Get(gomock.Any(), gomock.Any()).
				Do(func(recordID skydb.RecordID, record *skydb.Record) {
					if recordID.Type != "user" || recordID.Key != "some-friend" {
						panic("Wrong RecordID")
					}
				}).
				SetArg(1, skydb.Record{
					ID: skydb.NewRecordID("user", "some-friend"),
					Data: map[string]interface{}{
						"username": "testRelationConn",
					},
				}).
				Return(nil).
				AnyTimes()

			r := handlertest.NewSingleRouteRouter(&RelationAddHandler{}, func(p *router.Payload) {
				p.DBConn = &conn
				p.Database = db
			})

			resp := r.POST(`{
    "name": "friend",
    "targets": [
        "some-friend"
    ]
}`)

			So(conn.addedID, ShouldEqual, "some-friend")
			So(conn.RelationName, ShouldEqual, "_friend")
			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
    "result": [{
        "id": "some-friend",
        "type": "user",
        "data": {
            "_id": "some-friend",
            "username": "testRelationConn"
        }
    }]
}`)
		})
	})

	Convey("RelationQueryHandler", t, func() {
		conn := testRelationConn{}

		Convey("query outward relation", func() {
			r := handlertest.NewSingleRouteRouter(&RelationQueryHandler{}, func(p *router.Payload) {
				p.DBConn = &conn
			})

			resp := r.POST(`{
    "name": "follow",
    "direction": "outward"
}`)

			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
    "result": [],
    "info": {
        "count": 0
    }
}`)
		})

		Convey("query inward relation with count", func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			db := mock_skydb.NewMockTxDatabase(ctrl)
			db.EXPECT().
				Get(gomock.Any(), gomock.Any()).
				Do(func(recordID skydb.RecordID, record *skydb.Record) {
					if recordID.Type != "user" || recordID.Key != "101" {
						panic("Wrong RecordID")
					}
				}).
				SetArg(1, skydb.Record{
					ID: skydb.NewRecordID("user", "101"),
					Data: map[string]interface{}{
						"username": "user101",
						"email":    "user101@skygear.io",
					},
				}).
				Return(nil).
				AnyTimes()

			r := handlertest.NewSingleRouteRouter(&RelationQueryHandler{}, func(p *router.Payload) {
				p.DBConn = &conn
				p.Database = db
			})

			users := []skydb.AuthInfo{}
			users = append(users, skydb.AuthInfo{
				ID: "101",
			})
			conn.AuthInfo = users
			resp := r.POST(`{
    "name": "follow",
    "direction": "outward"
}`)

			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
    "result": [{
        "id": "101",
        "type": "user",
        "data":{
            "_id": "101",
            "email": "user101@skygear.io",
            "username": "user101"
        }
    }],
    "info": {
        "count": 1
    }
}`)
		})

		Convey("query relation with _follow", func() {
			r := handlertest.NewSingleRouteRouter(&RelationQueryHandler{}, func(p *router.Payload) {
				p.DBConn = &conn
			})

			resp := r.POST(`{
    "name": "_follow",
    "direction": "outward"
}`)

			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
    "result": [],
    "info": {
        "count": 0
    }
}`)
		})

		Convey("query relation with pagination", func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			db := mock_skydb.NewMockTxDatabase(ctrl)
			firstUser := db.EXPECT().
				Get(gomock.Any(), gomock.Any()).
				Do(func(recordID skydb.RecordID, record *skydb.Record) {
					if recordID.Type != "user" || recordID.Key != "101" {
						panic("Wrong RecordID")
					}
				}).
				SetArg(1, skydb.Record{
					ID: skydb.NewRecordID("user", "101"),
					Data: map[string]interface{}{
						"username": "user101",
						"email":    "user101@skygear.io",
					},
				}).
				Return(nil).
				Times(1)

			r := handlertest.NewSingleRouteRouter(&RelationQueryHandler{}, func(p *router.Payload) {
				p.DBConn = &conn
				p.Database = db
			})

			user1 := skydb.AuthInfo{
				ID: "101",
			}
			user2 := skydb.AuthInfo{
				ID: "102",
			}

			users := []skydb.AuthInfo{}
			users = append(users, user1)
			users = append(users, user2)
			conn.AuthInfo = users
			resp := r.POST(`{
    "name": "follow",
    "direction": "outward",
	"limit": 1
}`)

			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
    "result": [{
        "id": "101",
        "type": "user",
        "data":{
            "_id": "101",
            "email": "user101@skygear.io",
            "username": "user101"
        }
    }],
    "info": {
        "count": 2
    }
}`)

			db.EXPECT().
				Get(gomock.Any(), gomock.Any()).
				Do(func(recordID skydb.RecordID, record *skydb.Record) {
					if recordID.Type != "user" || recordID.Key != "102" {
						panic("Wrong RecordID")
					}
				}).
				SetArg(1, skydb.Record{
					ID: skydb.NewRecordID("user", "102"),
					Data: map[string]interface{}{
						"username": "user102",
						"email":    "user102@skygear.io",
					},
				}).
				Return(nil).
				After(firstUser)

			resp = r.POST(`{
    "name": "follow",
    "direction": "outward",
	"limit": 1,
	"offset": 1
}`)

			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
    "result": [{
        "id": "102",
        "type": "user",
        "data":{
            "_id": "102",
            "email": "user102@skygear.io",
            "username": "user102"
        }
    }],
    "info": {
        "count": 2
    }
}`)
		})

		Convey("query relation with wrong direction", func() {
			r := handlertest.NewSingleRouteRouter(&RelationQueryHandler{}, func(p *router.Payload) {
				p.DBConn = &conn
			})

			resp := r.POST(`{
    "name": "follow",
    "direction": "bidirection"
}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
    "error": {
        "code": 108,
        "message": "only outward, inward and mutual direction is allowed",
		"info": {"arguments": ["direction"]},
		"name": "InvalidArgument"
    }
}`)
		})
	})

	Convey("RelationRemoveHandler", t, func() {
		conn := testRelationConn{}
		r := handlertest.NewSingleRouteRouter(&RelationRemoveHandler{}, func(p *router.Payload) {
			p.DBConn = &conn
		})

		Convey("remove a relation", func() {
			resp := r.POST(`{
    "name": "friend",
    "targets": [
        "old-friend"
    ]
}`)

			So(conn.removeID, ShouldEqual, "old-friend")
			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
    "result": [{
        "id": "old-friend"
    }]
}`)
		})
	})
}
