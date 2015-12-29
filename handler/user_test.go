package handler

import (
	"github.com/oursky/skygear/handler/handlertest"
	. "github.com/oursky/skygear/skytest"
	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skydb/skydbtest"
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"github.com/oursky/skygear/provider"
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
		router := handlertest.NewSingleRouteRouter(UserQueryHandler, func(p *router.Payload) {
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
			Email:          "john.doe@example.com",
			HashedPassword: []byte("password"),
		}
		conn.CreateUser(&userInfo)

		router := handlertest.NewSingleRouteRouter(UserUpdateHandler, func(p *router.Payload) {
			p.DBConn = conn
			p.UserInfo = &userInfo
		})

		Convey("update email", func() {
			resp := router.POST(`{"email": "peter.doe@example.com"}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{}`)

			newUserInfo := skydb.UserInfo{}
			So(conn.GetUser("user0", &newUserInfo), ShouldBeNil)
			So(newUserInfo.Email, ShouldEqual, "peter.doe@example.com")
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
		r := handlertest.NewSingleRouteRouter(UserLinkHandler, func(p *router.Payload) {
			p.DBConn = conn
			p.UserInfo = &userInfo
			p.ProviderRegistry = providerRegistry
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
