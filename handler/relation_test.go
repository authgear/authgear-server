package handler

import (
	. "github.com/oursky/ourd/ourtest"
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"github.com/oursky/ourd/oddb"
	"github.com/oursky/ourd/router"
)

type testRelationConn struct {
	addedID   string
	removeID  string
	addErr    error
	removeErr error
	oddb.Conn
}

func (conn *testRelationConn) QueryRelation(user string, name string, direction string) []oddb.UserInfo {
	panic("not implemented")
}

func (conn *testRelationConn) AddRelation(user string, name string, targetUser string) error {
	conn.addedID = targetUser
	return nil
}

func (conn *testRelationConn) RemoveRelation(user string, name string, targetUser string) error {
	conn.removeID = targetUser
	return nil
}

func TestRelationQueryHandler(t *testing.T) {
	Convey("RelationAddHandler", t, func() {
		conn := testRelationConn{}
		r := newSingleRouteRouter(RelationAddHandler, func(p *router.Payload) {
			p.DBConn = &conn
		})

		Convey("add new relation", func() {
			resp := r.POST(`{
    "name": "friend",
    "targets": [
        "user/some-friend"
    ]
}`)

			So(conn.addedID, ShouldEqual, "some-friend")
			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
    "result": [{
        "_id": "user/some-friend"
    }]
}`)
		})

		Convey("add new relation with mal-format userid", func() {
			resp := r.POST(`{
    "name": "friend",
    "targets": [
        "I am Wrong"
    ]
}`)

			So(resp.Code, ShouldEqual, 400)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
    "error": {
        "type": "RequestInvalid",
        "code": 101,
        "message": "\"targets\" should be of format 'user/{id}', got \"I am Wrong\""
    }
}`)
		})
	})

	Convey("RelationRemoveHandler", t, func() {
		conn := testRelationConn{}
		r := newSingleRouteRouter(RelationRemoveHandler, func(p *router.Payload) {
			p.DBConn = &conn
		})

		Convey("remove a relation", func() {
			resp := r.POST(`{
    "name": "friend",
    "targets": [
        "user/old-friend"
    ]
}`)

			So(conn.removeID, ShouldEqual, "old-friend")
			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
    "result": [{
        "_id": "user/old-friend"
    }]
}`)
		})
	})
}
