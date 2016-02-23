package handler

import (
	"github.com/oursky/skygear/handler/handlertest"
	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skydb/skydbtest"
	. "github.com/oursky/skygear/skytest"
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"github.com/oursky/skygear/plugin/provider"
	"github.com/oursky/skygear/skydb"
)

type queryUserConn struct {
	skydb.Conn
}

func (userconn queryUserConn) QueryUser(emails []string) ([]skydb.UserInfo, error) {
	results := []skydb.UserInfo{}
	for _, email := range emails {
		if email == "john.doe@example.com" {
			results = append(results, skydb.UserInfo{
				ID:             "user0",
				Email:          "john.doe@example.com",
				Username:       "johndoe",
				HashedPassword: []byte("password"),
			})
		}
		if email == "jane.doe+1@example.com" {
			results = append(results, skydb.UserInfo{
				ID:             "user1",
				Username:       "johndoe1",
				Email:          "jane.doe+1@example.com",
				HashedPassword: []byte("password"),
			})
		}
	}
	return results, nil
}

func TestUserQueryHandler(t *testing.T) {
	Convey("UserQueryHandler", t, func() {
		router := handlertest.NewSingleRouteRouter(&UserQueryHandler{}, func(p *router.Payload) {
			p.DBConn = queryUserConn{}
		})

		Convey("query non-existent email", func() {
			resp := router.POST(`{
	"emails": ["peter.doe@example.com"]
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"result": [
	]
}`)
		})

		Convey("query single email", func() {
			resp := router.POST(`{
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
				"username": "johndoe"
			}
		}
	]
}`)
		})

		Convey("query multiple email", func() {
			resp := router.POST(`{
	"emails": ["john.doe@example.com", "jane.doe+1@example.com"]
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"result": [
		{
			"id": "user0",
			"type": "user",
			"data": {"_id": "user0", "email": "john.doe@example.com", "username": "johndoe"}
		}, {
			"id": "user1",
			"type": "user",
			"data": {"_id": "user1", "email": "jane.doe+1@example.com", "username": "johndoe1"}
		}
	]
}`)
		})
	})
}

func TestUserUpdateHandler(t *testing.T) {
	Convey("UserUpdateHandler", t, func() {
		conn := skydbtest.NewMapConn()
		userInfo := skydb.UserInfo{
			ID:             "user0",
			Username:       "john.doe",
			Email:          "john.doe@example.com",
			HashedPassword: []byte("password"),
		}
		conn.CreateUser(&userInfo)

		r := handlertest.NewSingleRouteRouter(&UserUpdateHandler{
			AccessModel: skydb.RoleBasedAccess,
		}, func(p *router.Payload) {
			p.DBConn = conn
			p.UserInfo = &userInfo
		})

		Convey("update email", func() {
			resp := r.POST(`{"email": "peter.doe@example.com"}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"result": {
		"_id": "user0",
		"username": "john.doe",
		"email": "peter.doe@example.com"
	}
}`)

			newUserInfo := skydb.UserInfo{}
			So(conn.GetUser("user0", &newUserInfo), ShouldBeNil)
			So(newUserInfo.Email, ShouldEqual, "peter.doe@example.com")
		})

		Convey("admin can expand its roles", func() {
			conn := skydbtest.NewMapConn()
			userInfo := skydb.UserInfo{
				ID:       "user0",
				Username: "username0",
				Roles:    []string{"admin"},
			}
			conn.CreateUser(&userInfo)

			r := handlertest.NewSingleRouteRouter(&UserUpdateHandler{
				AccessModel: skydb.RoleBasedAccess,
			}, func(p *router.Payload) {
				p.DBConn = conn
				p.UserInfo = &userInfo
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

		Convey("prevent non-admin assign expand his roles", func() {
			conn := skydbtest.NewMapConn()
			userInfo := skydb.UserInfo{
				ID:       "user0",
				Username: "username0",
			}
			conn.CreateUser(&userInfo)

			r := handlertest.NewSingleRouteRouter(&UserUpdateHandler{
				AccessModel: skydb.RoleBasedAccess,
			}, func(p *router.Payload) {
				p.DBConn = conn
				p.UserInfo = &userInfo
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
			userInfo := skydb.UserInfo{
				ID:       "user0",
				Username: "username0",
				Roles:    []string{"user", "human"},
			}
			conn.CreateUser(&userInfo)

			r := handlertest.NewSingleRouteRouter(&UserUpdateHandler{
				AccessModel: skydb.RoleBasedAccess,
			}, func(p *router.Payload) {
				p.DBConn = conn
				p.UserInfo = &userInfo
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
			userInfo := skydb.UserInfo{
				ID:       "user0",
				Username: "username0",
			}
			conn.CreateUser(&userInfo)

			r := handlertest.NewSingleRouteRouter(&UserUpdateHandler{
				AccessModel: skydb.RelationBasedAccess,
			}, func(p *router.Payload) {
				p.DBConn = conn
				p.UserInfo = &userInfo
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
			userInfo := skydb.UserInfo{
				ID:       "user0",
				Username: "username0",
				Roles:    []string{"admin"},
			}
			conn.CreateUser(&userInfo)

			r := handlertest.NewSingleRouteRouter(&UserUpdateHandler{
				AccessModel: skydb.RoleBasedAccess,
			}, func(p *router.Payload) {
				p.DBConn = conn
				p.UserInfo = &userInfo
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
		userInfo := skydb.UserInfo{
			ID:             "user0",
			Username:       "username0",
			Email:          "john.doe@example.com",
			HashedPassword: []byte("password"),
		}
		conn.CreateUser(&userInfo)
		userInfo2 := skydb.UserInfo{
			ID:             "user1",
			Username:       "username1",
			Email:          "john.doe@example.org",
			HashedPassword: []byte("password"),
			Auth: skydb.AuthInfo{
				"org.example:johndoe": map[string]interface{}{},
			},
		}
		conn.CreateUser(&userInfo2)

		providerRegistry := provider.NewRegistry()
		providerRegistry.RegisterAuthProvider("com.example", handlertest.NewSingleUserAuthProvider("com.example", "johndoe"))
		providerRegistry.RegisterAuthProvider("org.example", handlertest.NewSingleUserAuthProvider("org.example", "johndoe"))
		r := handlertest.NewSingleRouteRouter(&UserLinkHandler{
			ProviderRegistry: providerRegistry,
		}, func(p *router.Payload) {
			p.DBConn = conn
			p.UserInfo = &userInfo
		})

		Convey("link account", func() {
			resp := r.POST(`{"provider": "com.example", "auth_data": {"name": "johndoe"}}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{}`)

			newUserInfo := skydb.UserInfo{}
			So(conn.GetUser("user0", &newUserInfo), ShouldBeNil)
			So(newUserInfo.Auth["com.example:johndoe"], ShouldResemble, map[string]interface{}{
				"name": "johndoe",
			})
		})

		Convey("link account and unlink old", func() {
			resp := r.POST(`{"provider": "org.example", "auth_data": {"name": "johndoe"}}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{}`)

			newUserInfo := skydb.UserInfo{}
			So(conn.GetUser("user0", &newUserInfo), ShouldBeNil)
			So(newUserInfo.Auth["org.example:johndoe"], ShouldResemble, map[string]interface{}{
				"name": "johndoe",
			})
			So(conn.GetUser("user1", &newUserInfo), ShouldBeNil)
			_, err := newUserInfo.Auth["org.example:johndoe"]
			So(err, ShouldNotBeNil)
		})
	})
}
