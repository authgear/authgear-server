package handler

import (
	"sort"

	. "github.com/oursky/skygear/skytest"
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"github.com/oursky/skygear/handler/handlertest"
	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skydb"
)

type testRelationConn struct {
	UserInfo     []skydb.UserInfo
	RelationName string
	addedID      string
	removeID     string
	addErr       error
	removeErr    error
	skydb.Conn
}

func (conn *testRelationConn) GetUser(id string, userinfo *skydb.UserInfo) error {
	userinfo.ID = id
	userinfo.Username = "testRelationConn"
	return nil
}

type sortableUserInfo []skydb.UserInfo

func (a sortableUserInfo) Len() int           { return len(a) }
func (a sortableUserInfo) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a sortableUserInfo) Less(i, j int) bool { return a[i].ID < a[j].ID }

func (conn *testRelationConn) QueryRelation(user string, name string, direction string, config skydb.QueryConfig) []skydb.UserInfo {
	conn.RelationName = name
	if conn.UserInfo == nil {
		return []skydb.UserInfo{}
	}

	sort.Sort(sortableUserInfo(conn.UserInfo))
	if config.Limit == 0 {
		return conn.UserInfo[config.Offset:]
	}

	if config.Offset+config.Limit-1 > uint64(len(conn.UserInfo)) {
		return conn.UserInfo[config.Offset:]
	}
	return conn.UserInfo[config.Offset : config.Offset+config.Limit]
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
	if conn.UserInfo == nil {
		return 0, nil
	}
	count := uint64(len(conn.UserInfo))
	return count, nil
}

func TestRelationHandler(t *testing.T) {
	Convey("RelationAddHandler", t, func() {
		conn := testRelationConn{}
		r := handlertest.NewSingleRouteRouter(&RelationAddHandler{}, func(p *router.Payload) {
			p.DBConn = &conn
		})

		Convey("add new relation", func() {
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
		r := handlertest.NewSingleRouteRouter(&RelationQueryHandler{}, func(p *router.Payload) {
			p.DBConn = &conn
		})

		Convey("query outward relation", func() {
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
			users := []skydb.UserInfo{}
			users = append(users, skydb.UserInfo{
				ID:       "101",
				Username: "user101",
				Email:    "user101@skygear.io",
			})
			conn.UserInfo = users
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
			user1 := skydb.UserInfo{
				ID:       "101",
				Username: "user101",
				Email:    "user101@skygear.io",
			}
			user2 := skydb.UserInfo{
				ID:       "102",
				Username: "user102",
				Email:    "user102@skygear.io",
			}

			users := []skydb.UserInfo{}
			users = append(users, user1)
			users = append(users, user2)
			conn.UserInfo = users
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
