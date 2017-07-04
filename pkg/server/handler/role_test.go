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
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/server/handler/handlertest"
	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
	. "github.com/skygeario/skygear-server/pkg/server/skytest"

	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skydb/mock_skydb"
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

func TestRoleAssignHandler(t *testing.T) {
	Convey("RoleAssignHandler", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		conn := mock_skydb.NewMockConn(ctrl)

		Convey("should set role successfully", func() {
			payloadString := `{
	"roles": ["god", "buddha"],
	"users": ["johndoe", "janedoe"]
}`
			responseString := `{
	"result": "OK"
}`
			conn.EXPECT().AssignRoles(
				gomock.Eq([]string{"johndoe", "janedoe"}),
				gomock.Eq([]string{"god", "buddha"}),
			).Return(nil)

			mockRouter := handlertest.NewSingleRouteRouter(&RoleAssignHandler{}, func(p *router.Payload) {
				p.DBConn = conn
				p.AccessKey = router.MasterAccessKey
				p.AuthInfo = &skydb.AuthInfo{}
			})
			resp := mockRouter.POST(payloadString)
			So(resp.Body.Bytes(), ShouldEqualJSON, responseString)
		})

		Convey("should handle error", func() {
			conn.EXPECT().AssignRoles(
				gomock.Eq([]string{"johndoe", "janedoe"}),
				gomock.Eq([]string{"god", "buddha"}),
			).Return(skyerr.NewError(skyerr.UnexpectedError, "unexpected error"))

			mockRouter := handlertest.NewSingleRouteRouter(&RoleAssignHandler{}, func(p *router.Payload) {
				p.DBConn = conn
				p.AccessKey = router.MasterAccessKey
				p.AuthInfo = &skydb.AuthInfo{}
			})

			resp := mockRouter.POST(`{
		"roles": ["god", "buddha"],
		"users": ["johndoe", "janedoe"]
	}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
		"error":{
			"code":10000,
			"message":"unexpected error",
			"name":"UnexpectedError"
		}
	}`)
		})
	})
}

func TestRoleRevokeHandler(t *testing.T) {
	Convey("RoleRevokeHandler", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		conn := mock_skydb.NewMockConn(ctrl)

		Convey("should set role successfully", func() {
			payloadString := `{
	"roles": ["god", "buddha"],
	"users": ["johndoe", "janedoe"]
}`
			responseString := `{
	"result": "OK"
}`

			conn.EXPECT().RevokeRoles(
				gomock.Eq([]string{"johndoe", "janedoe"}),
				gomock.Eq([]string{"god", "buddha"}),
			).Return(nil)

			mockRouter := handlertest.NewSingleRouteRouter(&RoleRevokeHandler{}, func(p *router.Payload) {
				p.DBConn = conn
				p.AccessKey = router.MasterAccessKey
				p.AuthInfo = &skydb.AuthInfo{}
			})
			resp := mockRouter.POST(payloadString)
			So(resp.Body.Bytes(), ShouldEqualJSON, responseString)
		})

		Convey("should handle error", func() {
			conn.EXPECT().RevokeRoles(
				gomock.Eq([]string{"johndoe", "janedoe"}),
				gomock.Eq([]string{"god", "buddha"}),
			).Return(skyerr.NewError(skyerr.UnexpectedError, "unexpected error"))

			mockRouter := handlertest.NewSingleRouteRouter(&RoleRevokeHandler{}, func(p *router.Payload) {
				p.DBConn = conn
				p.AccessKey = router.MasterAccessKey
				p.AuthInfo = &skydb.AuthInfo{}
			})

			resp := mockRouter.POST(`{
		"roles": ["god", "buddha"],
		"users": ["johndoe", "janedoe"]
	}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
		"error":{
			"code":10000,
			"message":"unexpected error",
			"name":"UnexpectedError"
		}
	}`)
		})
	})
}
