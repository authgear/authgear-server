package handler

import (
	"github.com/oursky/skygear/handler/handlertest"
	"github.com/oursky/skygear/router"
	. "github.com/oursky/skygear/skytest"
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"github.com/oursky/skygear/skydb"
)

type roleConn struct {
	skydb.Conn
	adminRoles   []string
	defaultRoles []string
}

func (conn *roleConn) SetAdminRoles(roles []string) error {
	conn.adminRoles = roles
	return nil
}

func (conn *roleConn) SetDefaultRoles(roles []string) error {
	conn.defaultRoles = roles
	return nil
}

func TestRoleDefaultHandler(t *testing.T) {
	Convey("RoleDefaultHandler", t, func() {
		mockConn := &roleConn{}
		router := handlertest.NewSingleRouteRouter(&RoleDefaultHandler{}, func(p *router.Payload) {
			p.DBConn = mockConn
		})

		Convey("set role successfully", func() {
			resp := router.POST(`{
    "roles": ["human", "chinese"]
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
    "result": [
        "human",
        "chinese"
    ]
}`)
			So(mockConn.defaultRoles, ShouldResemble, []string{
				"human",
				"chinese",
			})
		})

		Convey("accept empty role request", func() {
			resp := router.POST(`{
    "roles": []
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
    "result": []
}`)
			So(mockConn.defaultRoles, ShouldResemble, []string{})
		})

		Convey("reject request without roles single email", func() {
			resp := router.POST(`{
    "no_roles": []
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
    "error": {
        "code": 107, 
        "message": "Missing roles key in request", 
        "name": "BadRequest"
    }
}`)
		})
	})
}

func TestRoleAdminHandler(t *testing.T) {
	Convey("RoleAdminHandler", t, func() {
		mockConn := &roleConn{}
		router := handlertest.NewSingleRouteRouter(&RoleAdminHandler{}, func(p *router.Payload) {
			p.DBConn = mockConn
		})

		Convey("set role successfully", func() {
			resp := router.POST(`{
    "roles": ["god", "buddha"]
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
    "result": [
        "god",
        "buddha"
    ]
}`)
			So(mockConn.adminRoles, ShouldResemble, []string{
				"god",
				"buddha",
			})
		})

		Convey("accept empty role request", func() {
			resp := router.POST(`{
    "roles": []
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
    "result": []
}`)
			So(mockConn.adminRoles, ShouldResemble, []string{})
		})

		Convey("reject request without roles single email", func() {
			resp := router.POST(`{
    "no_roles": []
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
    "error": {
        "code": 107, 
        "message": "Missing roles key in request", 
        "name": "BadRequest"
    }
}`)
		})
	})
}
