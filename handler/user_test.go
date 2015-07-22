package handler

import (
	"github.com/oursky/ourd/handler/handlertest"
	"github.com/oursky/ourd/oddb/oddbtest"
	. "github.com/oursky/ourd/ourtest"
	"github.com/oursky/ourd/router"
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"github.com/oursky/ourd/oddb"
	"github.com/oursky/ourd/provider"
)

type queryUserConn struct {
	oddb.Conn
}

func (userconn queryUserConn) QueryUser(emails []string) ([]oddb.UserInfo, error) {
	results := []oddb.UserInfo{}
	for _, email := range emails {
		if email == "john.doe@example.com" {
			results = append(results, oddb.UserInfo{
				ID:             "user0",
				Email:          "john.doe@example.com",
				HashedPassword: []byte("password"),
			})
		}
		if email == "jane.doe@example.com" {
			results = append(results, oddb.UserInfo{
				ID:             "user1",
				Email:          "jane.doe@example.com",
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
		{"id": "user0", "type": "user", "data": {"_id": "user0", "email": "john.doe@example.com"}}
	]
}`)
		})

		Convey("query multiple email", func() {
			resp := router.POST(`{
	"emails": ["john.doe@example.com", "jane.doe@example.com"]
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"result": [
		{"id": "user0", "type": "user", "data": {"_id": "user0", "email": "john.doe@example.com"}},
		{"id": "user1", "type": "user", "data": {"_id": "user1", "email": "jane.doe@example.com"}}
	]
}`)
		})
	})
}

func TestUserUpdateHandler(t *testing.T) {
	Convey("UserUpdateHandler", t, func() {
		conn := oddbtest.NewMapConn()
		userInfo := oddb.UserInfo{
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

			newUserInfo := oddb.UserInfo{}
			So(conn.GetUser("user0", &newUserInfo), ShouldBeNil)
			So(newUserInfo.Email, ShouldEqual, "peter.doe@example.com")
		})
	})
}

func TestUserLinkHandler(t *testing.T) {
	Convey("UserLinkHandler", t, func() {
		conn := oddbtest.NewMapConn()
		userInfo := oddb.UserInfo{
			ID:             "user0",
			Email:          "john.doe@example.com",
			HashedPassword: []byte("password"),
		}
		conn.CreateUser(&userInfo)

		providerRegistry := provider.NewRegistry()
		providerRegistry.RegisterAuthProvider("com.example", handlertest.NewSingleUserAuthProvider("com.example", "johndoe"))
		r := handlertest.NewSingleRouteRouter(UserLinkHandler, func(p *router.Payload) {
			p.DBConn = conn
			p.UserInfo = &userInfo
			p.ProviderRegistry = providerRegistry
		})

		Convey("link account", func() {
			resp := r.POST(`{"provider": "com.example", "auth_data": {"name": "johndoe"}}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{}`)

			newUserInfo := oddb.UserInfo{}
			So(conn.GetUser("user0", &newUserInfo), ShouldBeNil)
			So(newUserInfo.Auth["com.example:johndoe"], ShouldResemble, map[string]interface{}{
				"name": "johndoe",
			})
		})
	})
}
