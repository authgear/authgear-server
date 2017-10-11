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
		ctrl := gomock.NewController(handlertest.NewGoroutineAwareTestReporter(t))
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
		ctrl := gomock.NewController(handlertest.NewGoroutineAwareTestReporter(t))
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

func TestRoleGetHandler(t *testing.T) {
	Convey("RoleGetHandler", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		conn := mock_skydb.NewMockConn(ctrl)
		conn.EXPECT().GetAdminRoles().Return([]string{"admin"}, nil).AnyTimes()
		conn.EXPECT().GetRoles(gomock.Eq([]string{"user1", "user2", "user3"})).
			Return(map[string][]string{
				"user1": []string{
					"developer",
					"project-manager",
				},
				"user2": []string{},
				"user3": []string{
					"designer",
				},
			}, nil).AnyTimes()
		conn.EXPECT().GetRoles(gomock.Eq([]string{"user1"})).
			Return(map[string][]string{
				"user1": []string{
					"developer",
					"project-manager",
				},
			}, nil).AnyTimes()
		conn.EXPECT().GetRoles(gomock.Eq([]string{"error-user"})).
			Return(
				nil,
				skyerr.NewError(skyerr.UnexpectedError, "unexpected error"),
			).AnyTimes()

		Convey("should get roles successfully with master key", func() {
			handler := RoleGetHandler{}
			authInfo := skydb.AuthInfo{}
			mockRouter := handlertest.NewSingleRouteRouter(&handler, func(p *router.Payload) {
				p.DBConn = conn
				p.AuthInfo = &authInfo
				p.AccessKey = router.MasterAccessKey
			})
			resp := mockRouter.POST(
				`{ "users": ["user1", "user2", "user3"] }`,
			)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {
					"user1": ["developer", "project-manager"],
					"user2": [],
					"user3": ["designer"]
				}
			}`)
		})

		Convey("should get roles successfully with admin role", func() {
			handler := RoleGetHandler{}
			authInfo := skydb.AuthInfo{
				Roles: []string{"admin"},
			}
			mockRouter := handlertest.NewSingleRouteRouter(&handler, func(p *router.Payload) {
				p.DBConn = conn
				p.AuthInfo = &authInfo
				p.AccessKey = router.ClientAccessKey
			})
			resp := mockRouter.POST(`{ "users": ["user1", "user2", "user3"]}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {
					"user1": ["developer", "project-manager"],
					"user2": [],
					"user3": ["designer"]
				}
			}`)
		})

		Convey("should get roles successfully get my role", func() {
			handler := RoleGetHandler{}
			authInfo := skydb.AuthInfo{
				ID: "user1",
			}
			mockRouter := handlertest.NewSingleRouteRouter(&handler, func(p *router.Payload) {
				p.DBConn = conn
				p.AuthInfo = &authInfo
				p.AccessKey = router.ClientAccessKey
			})
			resp := mockRouter.POST(`{ "users": [ "user1" ] }`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {
					"user1": ["developer", "project-manager"]
				}
			}`)
		})

		Convey("should fail if no permission", func() {
			handler := RoleGetHandler{}
			authInfo := skydb.AuthInfo{
				ID: "user2",
			}
			mockRouter := handlertest.NewSingleRouteRouter(&handler, func(p *router.Payload) {
				p.DBConn = conn
				p.AuthInfo = &authInfo
				p.AccessKey = router.ClientAccessKey
			})
			resp := mockRouter.POST(`{ "users": [ "user1" ] }`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"code": 102,
					"name": "PermissionDenied",
					"message": "no permission to get other users' roles"
				}
			}`)
		})

		Convey("should handle error", func() {
			handler := RoleGetHandler{}
			authInfo := skydb.AuthInfo{}
			mockRouter := handlertest.NewSingleRouteRouter(&handler, func(p *router.Payload) {
				p.DBConn = conn
				p.AuthInfo = &authInfo
				p.AccessKey = router.MasterAccessKey
			})
			resp := mockRouter.POST(`{ "users": [ "error-user" ] }`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"code": 10000,
					"message": "unexpected error",
					"name": "UnexpectedError"
				}
			}`)
		})
	})
}
