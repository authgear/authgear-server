package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"

	. "github.com/skygeario/skygear-server/pkg/server/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetRoleHandler(t *testing.T) {
	Convey("Test GetRoleHandler", t, func() {
		authInfoStore := authinfo.NewMockStoreWithAuthInfoMap(
			map[string]authinfo.AuthInfo{
				"chima.cat.id": authinfo.AuthInfo{
					ID: "chima.cat.id",
					Roles: []string{
						"admin",
						"user",
					},
				},
				"faseng.cat.id": authinfo.AuthInfo{
					ID: "faseng.cat.id",
					Roles: []string{
						"user",
					},
				},
			},
		)

		rh := &GetRoleHandler{}
		rh.TxContext = db.NewMockTxContext()
		rh.AuthInfoStore = authInfoStore

		Convey("get role of current user", func() {
			rh.AuthContext = auth.NewMockContextGetterWithDefaultUser()
			h := handler.APIHandlerToHandler(rh, rh.TxContext)

			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"users": ["faseng.cat.id"]
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {
					"faseng.cat.id": ["user"]
				}
			}`)
		})

		Convey("should not get role of other user", func() {
			rh.AuthContext = auth.NewMockContextGetterWithDefaultUser()
			h := handler.APIHandlerToHandler(rh, rh.TxContext)

			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"users": ["chima.cat.id", "faseng.cat.id"]
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 403)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"code": 102,
					"message": "unable to get roles of other users",
					"name": "PermissionDenied"
				}
			}`)
		})

		Convey("get role of all users for admin", func() {
			rh.AuthContext = auth.NewMockContextGetterWithAdminUser()
			h := handler.APIHandlerToHandler(rh, rh.TxContext)

			req, _ := http.NewRequest("POST", "", strings.NewReader(`{
				"users": ["chima.cat.id", "faseng.cat.id"]
			}`))
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"result": {
					"faseng.cat.id": ["user"],
					"chima.cat.id": ["admin", "user"]
				}
			}`)
		})
	})
}
