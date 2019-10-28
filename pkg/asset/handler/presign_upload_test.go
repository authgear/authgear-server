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
			So(w.Body.Bytes(), ShouldEqualJSON, `
{"error":{"code":107,"info":{"arguments":["#: headers is required"],"causes":[{"message":"headers is required","pointer":"#"}]},"message":"Validation Error","name":"InvalidArgument"}}
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
			So(w.Body.Bytes(), ShouldEqualJSON, `
{"error":{"code":107,"info":{"arguments":["#/headers: content-length is required"],"causes":[{"message":"content-length is required","pointer":"#/headers"}]},"message":"Validation Error","name":"InvalidArgument"}}
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
			So(w.Body.Bytes(), ShouldEqualJSON, `
{"error":{"code":107,"info":{"arguments":["#/prefix: Does not match pattern '^[-_.a-zA-Z0-9]*$'"],"causes":[{"message":"Does not match pattern '^[-_.a-zA-Z0-9]*$'","pointer":"#/prefix"}]},"message":"Validation Error","name":"InvalidArgument"}}
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
