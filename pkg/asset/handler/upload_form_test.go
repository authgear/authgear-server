package handler

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/h2non/gock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/asset/dependency/presign"
	"github.com/skygeario/skygear-server/pkg/core/cloudstorage"
	. "github.com/skygeario/skygear-server/pkg/core/skytest"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

func TestUploadFormHandler(t *testing.T) {
	Convey("UploadFormHandler", t, func() {
		h := &UploadFormHandler{}
		validator := validation.NewValidator("http://v2.skygear.io")
		validator.AddSchemaFragments(
			PresignUploadRequestSchema,
		)
		provider := &cloudstorage.MockProvider{}
		h.CloudStorageProvider = provider
		h.Validator = validator
		h.PresignProvider = &presign.MockProvider{}

		Convey("Content-Type must be multipart/form-data", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			recorder := httptest.NewRecorder()
			h.ServeHTTP(recorder, req)

			So(recorder.Result().StatusCode, ShouldEqual, 400)
			So(recorder.Body.Bytes(), ShouldEqualJSON, `
{"error":{"code":400,"message":"invalid content-type","name":"BadRequest","reason":"BadAssetUploadForm"}}
			`)
		})

		Convey("Content-Type must have boundary", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			req.Header.Set("Content-Type", "multipart-form")
			recorder := httptest.NewRecorder()
			h.ServeHTTP(recorder, req)

			So(recorder.Result().StatusCode, ShouldEqual, 400)
			So(recorder.Body.Bytes(), ShouldEqualJSON, `
{"error":{"code":400,"message":"invalid content-type","name":"BadRequest","reason":"BadAssetUploadForm"}}
			`)
		})

		Convey("Reject repeated field", func() {
			buf := &bytes.Buffer{}
			w := multipart.NewWriter(buf)
			w.WriteField("content-type", "image/png")
			w.WriteField("content-type", "image/jpeg")
			w.Close()

			req, _ := http.NewRequest("POST", "/", buf)
			req.Header.Set("Content-Type", w.FormDataContentType())
			recorder := httptest.NewRecorder()
			h.ServeHTTP(recorder, req)

			So(recorder.Result().StatusCode, ShouldEqual, 400)
			So(recorder.Body.Bytes(), ShouldEqualJSON, `
{"error":{"code":400,"message":"repeated field: content-type","name":"BadRequest","reason":"BadAssetUploadForm"}}
			`)
		})

		Convey("Require exactly 1 file part", func() {
			buf := &bytes.Buffer{}
			w := multipart.NewWriter(buf)
			w.Close()

			req, _ := http.NewRequest("POST", "/", buf)
			req.Header.Set("Content-Type", w.FormDataContentType())
			recorder := httptest.NewRecorder()
			h.ServeHTTP(recorder, req)

			So(recorder.Result().StatusCode, ShouldEqual, 400)
			So(recorder.Body.Bytes(), ShouldEqualJSON, `
{"error":{"code":400,"message":"expected exactly 1 file part","name":"BadRequest","reason":"BadAssetUploadForm"}}
			`)
		})

		Convey("Reject unknown field", func() {
			buf := &bytes.Buffer{}
			w := multipart.NewWriter(buf)
			w.WriteField("unknown", "value")
			fileW, _ := w.CreateFormFile("file", "filename")
			w.Close()
			fileW.Write([]byte("Hello, World\b"))

			req, _ := http.NewRequest("POST", "/", buf)
			req.Header.Set("Content-Type", w.FormDataContentType())
			recorder := httptest.NewRecorder()
			h.ServeHTTP(recorder, req)

			So(recorder.Result().StatusCode, ShouldEqual, 400)
			So(recorder.Body.Bytes(), ShouldEqualJSON, `{
				"error": {
					"name": "Invalid",
					"reason": "ValidationFailed",
					"message": "invalid pre-signed request",
					"code": 400,
					"info": {
						"causes": [
							{ "kind": "ExtraEntry", "message": "Additional property unknown is not allowed", "pointer":"/headers/unknown" }
						]
					}
				}
			}`)
		})

		Convey("Success", func() {
			gock.InterceptClient(http.DefaultClient)
			defer gock.Off()
			defer gock.RestoreClient(http.DefaultClient)

			body := "Hello, World\n"

			provider.PresignUploadResponse = &cloudstorage.PresignUploadResponse{
				AssetName: "myimage.png",
				URL:       "http://example.com/app/myimage.png",
				Method:    "PUT",
				Headers:   []cloudstorage.HeaderField{},
			}

			gock.New("http://example.com").
				Put("/app/myimage.png").
				Reply(200)

			buf := &bytes.Buffer{}
			w := multipart.NewWriter(buf)
			fileW, _ := w.CreateFormFile("file", "filename")
			w.Close()
			fileW.Write([]byte(body))

			req, _ := http.NewRequest("POST", "/", buf)
			req.Header.Set("Content-Type", w.FormDataContentType())
			recorder := httptest.NewRecorder()
			h.ServeHTTP(recorder, req)

			So(recorder.Result().StatusCode, ShouldEqual, 200)
			So(recorder.Body.Bytes(), ShouldEqualJSON, `
{"result":{"asset_name":"myimage.png"}}
			`)
		})
	})
}
