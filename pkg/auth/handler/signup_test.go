package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/db"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	"github.com/skygeario/skygear-server/pkg/core/validation"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSignupHandler(t *testing.T) {
	Convey("Test SignupHandler", t, func() {
		sh := &SignupHandler{}
		validator := validation.NewValidator("http://v2.skygear.io")
		validator.AddSchemaFragments(
			SignupRequestSchema,
		)
		sh.Validator = validator
		sh.TxContext = db.NewMockTxContext()

		Convey("should reject request without login ID", func() {
			req, _ := http.NewRequest("POST", "", strings.NewReader(`
			{
				"login_ids": [],
				"password": "123456"
			}`))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			sh.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, 400)
			So(resp.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"name": "Invalid",
					"reason": "ValidationFailed",
					"message": "invalid request body",
					"code": 400,
					"info": {
						"causes": [
							{
								"kind": "EntryAmount",
								"pointer": "/login_ids",
								"message": "Array must have at least 1 items",
								"details": { "gte": 1 }
							}
						]
					}
				}
			}`)
		})
	})
}
