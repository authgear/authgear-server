package handler

import (
	. "github.com/oursky/ourd/ourtest"
	"github.com/oursky/ourd/router"
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"github.com/oursky/ourd/oddb"
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
		router := newSingleRouteRouter(UserQueryHandler, func(p *router.Payload) {
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
		{"id": "user0", "type": "user", "data": {"id": "user0", "email": "john.doe@example.com"}}
	]
}`)
		})

		Convey("query multiple email", func() {
			resp := router.POST(`{
	"emails": ["john.doe@example.com", "jane.doe@example.com"]
}`)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
	"result": [
		{"id": "user0", "type": "user", "data": {"id": "user0", "email": "john.doe@example.com"}},
		{"id": "user1", "type": "user", "data": {"id": "user1", "email": "jane.doe@example.com"}}
	]
}`)
		})
	})
}
