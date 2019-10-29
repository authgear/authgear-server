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

func TestPresignUploadHandler(t *testing.T) {
	Convey("PresignUploadHandler", t, func() {
		h := &PresignUploadHandler{}
		validator := validation.NewValidator("http://v2.skygear.io")
		validator.AddSchemaFragments(
			PresignUploadRequestSchema,
		)
		provider := &cloudstorage.MockProvider{}
		h.CloudStorageProvider = provider
		h.Validator = validator

		Convey("headers is required", func() {
			requestBody := []byte(`{}`)
			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/_asset/presign_upload", bytes.NewReader(requestBody))
			r.Header.Add("content-type", "application/json")
			h.ServeHTTP(w, r)

			So(w.Code, ShouldEqual, 400)
			// TODO(error): validation
			So(w.Body.Bytes(), ShouldEqualJSON, `
{"error":{"code":400,"message":"Validation Error","name":"Invalid","reason":"Invalid"}}
			`)
		})

		Convey("content-length is required", func() {
			requestBody := []byte(`{
				"headers": {}
			}`)
			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/_asset/presign_upload", bytes.NewReader(requestBody))
			r.Header.Add("content-type", "application/json")
			h.ServeHTTP(w, r)

			So(w.Code, ShouldEqual, 400)
			// TODO(error): validation
			So(w.Body.Bytes(), ShouldEqualJSON, `
{"error":{"code":400,"message":"Validation Error","name":"Invalid","reason":"Invalid"}}
			`)
		})

		Convey("prefix must be safe", func() {
			requestBody := []byte(`{
				"prefix": "/",
				"headers": {
					"content-length": "123"
				}
			}`)
			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/_asset/presign_upload", bytes.NewReader(requestBody))
			r.Header.Add("content-type", "application/json")
			h.ServeHTTP(w, r)

			So(w.Code, ShouldEqual, 400)
			// TODO(error): validation
			So(w.Body.Bytes(), ShouldEqualJSON, `
{"error":{"code":400,"message":"Validation Error","name":"Invalid","reason":"Invalid"}}
			`)
		})

		Convey("success", func() {
			requestBody := []byte(`{
				"prefix": "-_.azAZ09",
				"headers": {
					"content-length": "123"
				}
			}`)
			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/_asset/presign_upload", bytes.NewReader(requestBody))
			r.Header.Add("content-type", "application/json")
			provider.PresignUploadResponse = &cloudstorage.PresignUploadResponse{
				AssetName: "myimage.png",
				URL:       "http://example.com/app/myimage.png",
				Method:    "PUT",
				Headers: []cloudstorage.HeaderField{
					cloudstorage.HeaderField{
						Name:  "Content-Length",
						Value: "123",
					},
				},
			}
			h.ServeHTTP(w, r)

			So(w.Code, ShouldEqual, 200)
			So(w.Body.Bytes(), ShouldEqualJSON, `
{"result":{"asset_name":"myimage.png","headers":[{"name":"Content-Length","value":"123"}],"method":"PUT","url":"http://example.com/app/myimage.png"}}
			`)
		})
	})
}
