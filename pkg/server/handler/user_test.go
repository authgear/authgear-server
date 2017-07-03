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
	"net/http"
	"testing"

	"github.com/skygeario/skygear-server/pkg/server/handler/handlertest"
	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skydb/skydbtest"
	. "github.com/skygeario/skygear-server/pkg/server/skytest"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/server/plugin/provider"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
)

type queryUserConn struct {
	skydb.Conn
}

func (userconn queryUserConn) QueryUser(emails []string, usernames []string) ([]skydb.AuthInfo, error) {
	myAuthInfos := []skydb.AuthInfo{
		skydb.AuthInfo{
			ID:             "user0",
			Email:          "john.doe@example.com",
			Username:       "johndoe",
			HashedPassword: []byte("password"),
			Roles: []string{
				"Programmer",
			},
		},
		skydb.AuthInfo{
			ID:             "user1",
			Username:       "johndoe1",
			Email:          "jane.doe+1@example.com",
			HashedPassword: []byte("password"),
		},
	}

	results := []skydb.AuthInfo{}
	for _, email := range emails {
		for _, authInfo := range myAuthInfos {
			if authInfo.Email == email {
				results = append(results, authInfo)
				break
			}
		}
	}
	for _, username := range usernames {
		for _, authInfo := range myAuthInfos {
			if authInfo.Username == username {
				results = append(results, authInfo)
				break
			}
		}
	}
	return results, nil
}

func (userconn queryUserConn) GetAdminRoles() ([]string, error) {
	return []string{
		"Admin",
	}, nil
}

func TestUserQueryHandler(t *testing.T) {
	Convey("UserQueryHandler", t, func() {
		adminAuthInfo := skydb.AuthInfo{
			ID:             "admin",
			Email:          "admin@example.com",
			Username:       "admin",
			HashedPassword: []byte("password"),
			Roles: []string{
				"Admin",
			},
		}

		adminRouter := handlertest.NewSingleRouteRouter(&UserQueryHandler{}, func(p *router.Payload) {
			p.DBConn = queryUserConn{}
			p.AuthInfo = &adminAuthInfo
		})

		Convey("query non-existent email", func() {
			resp := adminRouter.POST(`{
	"emails": ["peter.doe@example.com"]
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"result": [
	]
}`)
		})

		Convey("query single email", func() {
			resp := adminRouter.POST(`{
	"emails": ["john.doe@example.com"]
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"result": [
		{
			"id": "user0",
			"type": "user",
			"data": {
				"_id": "user0",
				"email": "john.doe@example.com",
				"username": "johndoe",
				"roles": [
					"Programmer"
				]
			}
		}
	]
}`)
		})

		Convey("query multiple email", func() {
			resp := adminRouter.POST(`{
	"emails": ["john.doe@example.com", "jane.doe+1@example.com"]
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"result": [
		{
			"id": "user0",
			"type": "user",
			"data": {"_id": "user0", "email": "john.doe@example.com", "username": "johndoe", "roles": ["Programmer"]}
		}, {
			"id": "user1",
			"type": "user",
			"data": {"_id": "user1", "email": "jane.doe+1@example.com", "username": "johndoe1"}
		}
	]
}`)
		})

		Convey("query with email and username", func() {
			resp := adminRouter.POST(`{
	"emails": ["john.doe@example.com"],
	"usernames": ["johndoe1"]
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"result": [
		{
			"id": "user0",
			"type": "user",
			"data": {"_id": "user0", "email": "john.doe@example.com", "username": "johndoe", "roles": ["Programmer"]}
		}, {
			"id": "user1",
			"type": "user",
			"data": {"_id": "user1", "email": "jane.doe+1@example.com", "username": "johndoe1"}
		}
	]
}`)
		})

		nonAdminAuthInfo := skydb.AuthInfo{
			ID:             "non-admin",
			Email:          "non-admin@example.com",
			Username:       "non-admin",
			HashedPassword: []byte("password"),
		}

		Convey("query non-existent email with non-admin", func() {
			nonAdminRouter := handlertest.NewSingleRouteRouter(&UserQueryHandler{}, func(p *router.Payload) {
				p.DBConn = queryUserConn{}
				p.AuthInfo = &nonAdminAuthInfo
			})
			resp := nonAdminRouter.POST(`{
				"emails": ["peter.doe@example.com"]
			}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"code": 102,
					"message": "No permission to query user",
					"name":"PermissionDenied"
				}
			}`)
		})

		Convey("query non-existent email with non-admin and master key", func() {
			nonAdminMasterKeyRouter := handlertest.NewSingleRouteRouter(
				&UserQueryHandler{},
				func(p *router.Payload) {
					p.DBConn = queryUserConn{}
					p.AuthInfo = &nonAdminAuthInfo
					p.AccessKey = router.MasterAccessKey
				},
			)

			resp := nonAdminMasterKeyRouter.POST(`{
				"emails": ["peter.doe@example.com"]
			}`)

			So(resp.Body.Bytes(), ShouldEqualJSON, `{"result":[]}`)
		})
	})
}

func TestUserUpdateHandler(t *testing.T) {
	Convey("UserUpdateHandler", t, func() {
		conn := skydbtest.NewMapConn()
		authInfo := skydb.AuthInfo{
			ID:             "user0",
			Username:       "john.doe",
			Email:          "john.doe@example.com",
			HashedPassword: []byte("password"),
		}
		conn.CreateUser(&authInfo)

		r := handlertest.NewSingleRouteRouter(&UserUpdateHandler{
			AccessModel: skydb.RoleBasedAccess,
		}, func(p *router.Payload) {
			p.DBConn = conn
			p.AuthInfo = &authInfo
		})

		Convey("update email", func() {
			resp := r.POST(`{
	"_id": "user0",
	"email": "peter.doe@example.com"
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"result": {
		"_id": "user0",
		"username": "john.doe",
		"email": "peter.doe@example.com"
	}
}`)

			newAuthInfo := skydb.AuthInfo{}
			So(conn.GetUser("user0", &newAuthInfo), ShouldBeNil)
			So(newAuthInfo.Email, ShouldEqual, "peter.doe@example.com")
		})

		Convey("admin can expand its roles", func() {
			conn := skydbtest.NewMapConn()
			authInfo := skydb.AuthInfo{
				ID:       "user0",
				Username: "username0",
				Roles:    []string{"admin"},
			}
			conn.CreateUser(&authInfo)

			r := handlertest.NewSingleRouteRouter(&UserUpdateHandler{
				AccessModel: skydb.RoleBasedAccess,
			}, func(p *router.Payload) {
				p.DBConn = conn
				p.AuthInfo = &authInfo
			})

			resp := r.POST(`{
	"_id": "user0",
	"roles": ["admin", "writer"]
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"result": {
		"_id": "user0",
		"username": "username0",
		"email": "",
		"roles": ["admin", "writer"]
	}
}`)
		})

		Convey("client with master key can expand its roles", func() {
			conn := skydbtest.NewMapConn()
			authInfo := skydb.AuthInfo{
				ID:       "user0",
				Username: "username0",
				Roles:    []string{},
			}
			conn.CreateUser(&authInfo)

			r := handlertest.NewSingleRouteRouter(&UserUpdateHandler{
				AccessModel: skydb.RoleBasedAccess,
			}, func(p *router.Payload) {
				p.DBConn = conn
				p.AuthInfo = &authInfo
				p.AccessKey = router.MasterAccessKey
			})

			resp := r.POST(`{
	"_id": "user0",
	"roles": ["admin", "writer"]
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"result": {
		"_id": "user0",
		"username": "username0",
		"email": "",
		"roles": ["admin", "writer"]
	}
}`)
		})

		Convey("missing roles key in payload will left the roles un-modified", func() {
			conn := skydbtest.NewMapConn()
			authInfo := skydb.AuthInfo{
				ID:       "user0",
				Username: "username0",
				Roles:    []string{"admin"},
			}
			conn.CreateUser(&authInfo)

			r := handlertest.NewSingleRouteRouter(&UserUpdateHandler{
				AccessModel: skydb.RoleBasedAccess,
			}, func(p *router.Payload) {
				p.DBConn = conn
				p.AuthInfo = &authInfo
			})

			resp := r.POST(`{
	"_id": "user0"
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"result": {
		"_id": "user0",
		"username": "username0",
		"email": "",
		"roles": ["admin"]
	}
}`)
		})
		Convey("prevent non-admin assign expand his roles", func() {
			conn := skydbtest.NewMapConn()
			authInfo := skydb.AuthInfo{
				ID:       "user0",
				Username: "username0",
			}
			conn.CreateUser(&authInfo)

			r := handlertest.NewSingleRouteRouter(&UserUpdateHandler{
				AccessModel: skydb.RoleBasedAccess,
			}, func(p *router.Payload) {
				p.DBConn = conn
				p.AuthInfo = &authInfo
			})

			resp := r.POST(`{
	"_id": "user0",
	"roles": ["admin", "writer"]
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"error": {
		"code": 102,
		"message": "no permission to add new roles",
		"name": "PermissionDenied"
	}
}`)
		})

		Convey("allow non-admin user to remove his roles", func() {
			conn := skydbtest.NewMapConn()
			authInfo := skydb.AuthInfo{
				ID:       "user0",
				Username: "username0",
				Roles:    []string{"user", "human"},
			}
			conn.CreateUser(&authInfo)

			r := handlertest.NewSingleRouteRouter(&UserUpdateHandler{
				AccessModel: skydb.RoleBasedAccess,
			}, func(p *router.Payload) {
				p.DBConn = conn
				p.AuthInfo = &authInfo
			})

			resp := r.POST(`{
	"_id": "user0",
	"roles": ["human"]
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"result": {
		"_id": "user0",
		"username": "username0",
		"email": "",
		"roles": ["human"]
	}
}`)
		})

		Convey("update roles at non role base configuration result in error", func() {
			conn := skydbtest.NewMapConn()
			authInfo := skydb.AuthInfo{
				ID:       "user0",
				Username: "username0",
			}
			conn.CreateUser(&authInfo)

			r := handlertest.NewSingleRouteRouter(&UserUpdateHandler{
				AccessModel: skydb.RelationBasedAccess,
			}, func(p *router.Payload) {
				p.DBConn = conn
				p.AuthInfo = &authInfo
			})

			resp := r.POST(`{
	"_id": "user0",
	"roles": ["admin", "writer"]
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"error": {
		"code": 108,
		"info": {
			"arguments": ["roles"]
		},
		"message": "Cannot assign user role on AcceesModel is not RoleBaseAccess",
		"name": "InvalidArgument"
	}
}`)
		})

		Convey("Regression #564, update roles with duplicated roles will result in InvalidaArgumnet", func() {
			conn := skydbtest.NewMapConn()
			authInfo := skydb.AuthInfo{
				ID:       "user0",
				Username: "username0",
				Roles:    []string{"admin"},
			}
			conn.CreateUser(&authInfo)

			r := handlertest.NewSingleRouteRouter(&UserUpdateHandler{
				AccessModel: skydb.RoleBasedAccess,
			}, func(p *router.Payload) {
				p.DBConn = conn
				p.AuthInfo = &authInfo
			})

			resp := r.POST(`{
	"username": "username0",
	"roles": ["admin", "admin"]
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"error": {
		"code": 108,
		"info": {
			"arguments": ["roles"]
		},
		"message": "duplicated roles in payload",
		"name": "InvalidArgument"
	}
}`)
		})
	})
}

func TestUserLinkHandler(t *testing.T) {
	Convey("UserLinkHandler", t, func() {
		conn := skydbtest.NewMapConn()
		authInfo := skydb.AuthInfo{
			ID:             "user0",
			Username:       "username0",
			Email:          "john.doe@example.com",
			HashedPassword: []byte("password"),
		}
		conn.CreateUser(&authInfo)
		authInfo2 := skydb.AuthInfo{
			ID:             "user1",
			Username:       "username1",
			Email:          "john.doe@example.org",
			HashedPassword: []byte("password"),
			Auth: skydb.ProviderInfo{
				"org.example:johndoe": map[string]interface{}{},
			},
		}
		conn.CreateUser(&authInfo2)

		providerRegistry := provider.NewRegistry()
		providerRegistry.RegisterAuthProvider("com.example", handlertest.NewSingleUserAuthProvider("com.example", "johndoe"))
		providerRegistry.RegisterAuthProvider("org.example", handlertest.NewSingleUserAuthProvider("org.example", "johndoe"))
		r := handlertest.NewSingleRouteRouter(&UserLinkHandler{
			ProviderRegistry: providerRegistry,
		}, func(p *router.Payload) {
			p.DBConn = conn
			p.AuthInfo = &authInfo
		})

		Convey("link account with non-existent provider", func() {
			resp := r.POST(`{"provider": "com.non-existent", "auth_data": {"name": "johndoe"}}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"error": {
		"code": 108,
		"name": "InvalidArgument",
		"info": {"arguments": ["provider"]},
		"message": "no auth provider of name \"com.non-existent\""
	}
}`)
			So(resp.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("link account", func() {
			resp := r.POST(`{"provider": "com.example", "auth_data": {"name": "johndoe"}}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{}`)

			newAuthInfo := skydb.AuthInfo{}
			So(conn.GetUser("user0", &newAuthInfo), ShouldBeNil)
			So(newAuthInfo.Auth["com.example:johndoe"], ShouldResemble, map[string]interface{}{
				"name": "johndoe",
			})
		})

		Convey("link account and unlink old", func() {
			resp := r.POST(`{"provider": "org.example", "auth_data": {"name": "johndoe"}}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{}`)

			newAuthInfo := skydb.AuthInfo{}
			So(conn.GetUser("user0", &newAuthInfo), ShouldBeNil)
			So(newAuthInfo.Auth["org.example:johndoe"], ShouldResemble, map[string]interface{}{
				"name": "johndoe",
			})
			So(conn.GetUser("user1", &newAuthInfo), ShouldBeNil)
			_, err := newAuthInfo.Auth["org.example:johndoe"]
			So(err, ShouldNotBeNil)
		})
	})
}
