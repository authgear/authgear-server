package handler

import (
	"testing"

	"github.com/oursky/skygear/handler/handlertest"
	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skyerr"
	. "github.com/oursky/skygear/skytest"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/oursky/skygear/skydb"
)

func TestRolePayload(t *testing.T) {
	Convey("rolePaylod", t, func() {
		Convey("valid data", func() {
			payload := rolePayload{}
			payload.Decode(map[string]interface{}{
				"roles": []string{
					"admin",
					"system",
				},
			})
			So(payload.Validate(), ShouldBeNil)
			So(payload.Roles, ShouldResemble, []string{
				"admin",
				"system",
			})
		})
		Convey("missing roles", func() {
			payload := rolePayload{}
			payload.Decode(map[string]interface{}{})
			So(
				payload.Validate(),
				ShouldResemble,
				skyerr.NewInvalidArgument("unspecified roles in request", []string{"roles"}),
			)
		})
	})
}

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
        "code": 108,
        "message": "unspecified roles in request",
        "info": {
            "arguments": [
                "roles"
            ]
        },
        "name": "InvalidArgument"
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
        "code": 108,
        "message": "unspecified roles in request",
        "info": {
            "arguments": [
                "roles"
            ]
        },
        "name": "InvalidArgument"
    }
}`)
		})
	})
}
