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

	Convey("RelationRemoveHandler", t, func() {
		conn := testRelationConn{}
		r := newSingleRouteRouter(RelationRemoveHandler, func(p *router.Payload) {
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
