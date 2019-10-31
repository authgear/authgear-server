package handler

import (
	"bytes"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/core/cloudstorage"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func TestSignHandler(t *testing.T) {
	Convey("SignHandler", t, func() {
		h := &SignHandler{}
		validator := validation.NewValidator("http://v2.skygear.io")
		validator.AddSchemaFragments(
			SignRequestSchema,
		)
		provider := &cloudstorage.MockProvider{}
		h.CloudStorageProvider = provider
		h.Validator = validator

		Convey("assets is required", func() {
			requestBody := []byte(`{}`)
			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/_asset/sign", bytes.NewReader(requestBody))
			r.Header.Add("content-type", "application/json")
			h.ServeHTTP(w, r)

			So(w.Code, ShouldEqual, 400)
			So(w.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"name": "Invalid",
					"reason": "ValidationFailed",
					"message": "invalid request body",
					"code": 400,
					"info": {
						"causes": [
							{ "kind": "Required", "message": "assets is required", "pointer":"/assets" }
						]
					}
				}
			}`)
		})

		Convey("asset_name is required and non-empty", func() {
			requestBody := []byte(`
			{
				"assets": [
					{},
					{ "asset_name": "" }
				]
			}
			`)
			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/_asset/sign", bytes.NewReader(requestBody))
			r.Header.Add("content-type", "application/json")
			h.ServeHTTP(w, r)

			So(w.Code, ShouldEqual, 400)
			So(w.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"name": "Invalid",
					"reason": "ValidationFailed",
					"message": "invalid request body",
					"code": 400,
					"info": {
						"causes": [
							{ "kind": "Required", "message": "asset_name is required", "pointer":"/assets/0/asset_name" },
							{ "kind": "StringLength", "message": "String length must be greater than or equal to 1", "pointer":"/assets/1/asset_name", "details": { "gte": 1 } }
						]
					}
				}
			}`)
		})

		Convey("success", func() {
			requestBody := []byte(`
			{
				"assets": [
					{ "asset_name": "myimage.png" }
				]
			}
			`)
			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/_asset/sign", bytes.NewReader(requestBody))
			r.Header.Add("content-type", "application/json")
			h.ServeHTTP(w, r)

			So(w.Code, ShouldEqual, 200)
			So(w.Body.Bytes(), ShouldEqualJSON, `
{"result":{"assets":[{"asset_name":"myimage.png","url":"http://example.com/_asset/get/myimage.png"}]}}
			`)
		})
	})
}
