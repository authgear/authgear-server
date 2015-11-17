package handler

import (
	. "github.com/oursky/skygear/ourtest"
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"github.com/oursky/skygear/handler/handlertest"
	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skydb"
)

type testRelationConn struct {
	addedID   string
	removeID  string
	addErr    error
	removeErr error
	skydb.Conn
}

func (conn *testRelationConn) QueryRelation(user string, name string, direction string) []skydb.UserInfo {
	return []skydb.UserInfo{}
}

func (conn *testRelationConn) AddRelation(user string, name string, targetUser string) error {
	conn.addedID = targetUser
	return nil
}

func (conn *testRelationConn) RemoveRelation(user string, name string, targetUser string) error {
	conn.removeID = targetUser
	return nil
}

func TestRelationHandler(t *testing.T) {
	Convey("RelationAddHandler", t, func() {
		conn := testRelationConn{}
		r := handlertest.NewSingleRouteRouter(RelationAddHandler, func(p *router.Payload) {
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
			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
    "result": [{
        "id": "some-friend"
    }]
}`)
		})
	})

	Convey("RelationQueryHandler", t, func() {
		conn := testRelationConn{}
		r := handlertest.NewSingleRouteRouter(RelationQueryHandler, func(p *router.Payload) {
			p.DBConn = &conn
		})

		Convey("query outward relation", func() {
			resp := r.POST(`{
    "name": "follow",
    "direction": "outward"
}`)

			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
    "result": []
}`)
		})

		Convey("query relation with wrong direction", func() {
			resp := r.POST(`{
    "name": "follow",
    "direction": "bidirection"
}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
    "error": {
        "code": 101,
        "message": "Only outward, inward and mutual direction is supported",
        "type": "RequestInvalid"
    }
}`)
		})
	})

	Convey("RelationRemoveHandler", t, func() {
		conn := testRelationConn{}
		r := handlertest.NewSingleRouteRouter(RelationRemoveHandler, func(p *router.Payload) {
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
