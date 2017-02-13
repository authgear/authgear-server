// Copyright 2015-present Oursky Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handler

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/skygeario/skygear-server/pkg/server/asset"
	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	. "github.com/skygeario/skygear-server/pkg/server/skytest"
	. "github.com/smartystreets/goconvey/convey"
)

// a thin wrapper on router.Gateway with helper methods to invoke request more
// easily
type modGateway router.Gateway

func newmodGateway(pattern string) *modGateway {
	return (*modGateway)(router.NewGateway(pattern, "/", nil))
}

func (g *modGateway) Handle(method string, handler router.Handler, prepareFunc func(*router.Payload)) {
	(*router.Gateway)(g).Handle(method, handler, mockProcessor{prepareFunc})
}

func (g *modGateway) Do(req *http.Request) *httptest.ResponseRecorder {
	resp := httptest.NewRecorder()
	(*router.Gateway)(g).ServeHTTP(resp, req)
	return resp
}

func (g *modGateway) GET(path string) *httptest.ResponseRecorder {
	return g.makeRequest("GET", path, "")
}

func (g *modGateway) PUT(path, body string) *httptest.ResponseRecorder {
	return g.makeRequest("PUT", path, body)
}

func (g *modGateway) makeRequest(method, path, body string) *httptest.ResponseRecorder {
	path = "http://skygear.test/" + path
	req, _ := http.NewRequest(method, path, strings.NewReader(body))
	resp := httptest.NewRecorder()

	(*router.Gateway)(g).ServeHTTP(resp, req)
	return resp
}

type mockProcessor struct {
	Mockfunc func(*router.Payload)
}

func (p mockProcessor) Preprocess(payload *router.Payload, _ *router.Response) int {
	p.Mockfunc(payload)
	return http.StatusOK
}

type naiveAssetConn struct {
	skydb.Conn
	savedAsset map[string]*skydb.Asset
}

func (c *naiveAssetConn) GetAsset(name string, asset *skydb.Asset) error {
	saved := c.savedAsset[name]
	if saved != nil {
		*asset = *saved
	}
	return nil
}

func (c *naiveAssetConn) SaveAsset(asset *skydb.Asset) error {
	c.savedAsset[asset.Name] = asset
	return nil
}

// an asset store that writes to and reads from the same buf
type bufferedAssetStore struct {
	buf         *bytes.Buffer
	name        string
	length      int64
	contentType string
}

func newBufferedStore() *bufferedAssetStore {
	return &bufferedAssetStore{
		buf: &bytes.Buffer{},
	}
}

func (store *bufferedAssetStore) GetFileReader(name string) (io.ReadCloser, error) {
	return ioutil.NopCloser(store.buf), nil
}

func (store *bufferedAssetStore) PutFileReader(name string, src io.Reader, length int64, contentType string) error {
	store.name = name
	store.length = length
	store.contentType = contentType

	written, err := io.Copy(store.buf, src)
	if err != nil {
		return err
	}

	if written != length {
		return fmt.Errorf("got bytes written = %v, want %v", written, length)
	}

	return nil
}

func (store *bufferedAssetStore) GeneratePostFileRequest(name string) (*asset.PostFileRequest, error) {
	return &asset.PostFileRequest{
		Action: "http://skygear.test/files/" + name,
		ExtraFields: map[string]interface{}{
			"X-Extra-Field-1": "extra-field-1-value",
			"X-Extra-Field-2": "extra-field-2-value",
		},
	}, nil
}

func (store *bufferedAssetStore) SignedURL(name string) (string, error) {
	return name + "?signedurl=true", nil
}

func (store *bufferedAssetStore) IsSignatureRequired() bool {
	return false
}

func TestUploadFileHandler(t *testing.T) {
	Convey("UploadFileHandler", t, func() {
		assetConn := &naiveAssetConn{}
		assetConn.savedAsset = map[string]*skydb.Asset{}

		store := newBufferedStore()

		r := newmodGateway("(.+)")
		r.Handle("PUT", &UploadFileHandler{
			AssetStore: store,
		}, func(p *router.Payload) {
			p.DBConn = assetConn
		})

		Convey("uploads a file", func() {
			assetConn.savedAsset["c34e739e-ac82-44c0-b36b-28d226edb237-asset"] = &skydb.Asset{
				Name:        "c34e739e-ac82-44c0-b36b-28d226edb237-asset",
				ContentType: "plain/text",
				Size:        0,
			}

			req, _ := http.NewRequest(
				"PUT",
				"http://skygear.test/c34e739e-ac82-44c0-b36b-28d226edb237-asset",
				strings.NewReader(``),
			)
			req.Header.Set("Content-Type", "plain/text")
			req.Body = ioutil.NopCloser(strings.NewReader(`I am a boy`))

			resp := r.Do(req)

			savedAsset := assetConn.savedAsset["c34e739e-ac82-44c0-b36b-28d226edb237-asset"]
			So(savedAsset, ShouldNotBeNil)
			So(savedAsset.Name, ShouldEqual, "c34e739e-ac82-44c0-b36b-28d226edb237-asset")
			So(savedAsset.ContentType, ShouldEqual, "plain/text")
			So(savedAsset.Size, ShouldEqual, 10)

			So(store.name, ShouldEqual, "c34e739e-ac82-44c0-b36b-28d226edb237-asset")
			So(store.length, ShouldEqual, 10)
			So(store.contentType, ShouldEqual, "plain/text")
			So(store.buf.String(), ShouldEqual, "I am a boy")

			So(resp.Body.String(), ShouldEqualJSON, `{
				"result": {
					"$type": "asset",
					"$name": "c34e739e-ac82-44c0-b36b-28d226edb237-asset",
					"$url": "c34e739e-ac82-44c0-b36b-28d226edb237-asset?signedurl=true",
					"$content_type":"plain/text"
				}
			}`)
		})

		Convey("refs #426 uploads a file with + character", func() {
			assetConn.savedAsset["78640a1a-25c7-45a1-8c1f-cf8d3b162f9e-helloworld"] = &skydb.Asset{
				Name:        "78640a1a-25c7-45a1-8c1f-cf8d3b162f9e-helloworld",
				ContentType: "plain/text",
				Size:        0,
			}

			req, _ := http.NewRequest(
				"PUT",
				"http://skygear.test/78640a1a-25c7-45a1-8c1f-cf8d3b162f9e-hello+world",
				strings.NewReader(``),
			)
			req.Header.Set("Content-Type", "plain/text")
			req.Body = ioutil.NopCloser(strings.NewReader(`I am a boy`))

			resp := r.Do(req)

			savedAsset := assetConn.savedAsset["78640a1a-25c7-45a1-8c1f-cf8d3b162f9e-helloworld"]
			So(savedAsset, ShouldNotBeNil)
			So(savedAsset.Name, ShouldEqual, "78640a1a-25c7-45a1-8c1f-cf8d3b162f9e-helloworld")
			So(savedAsset.ContentType, ShouldEqual, "plain/text")
			So(savedAsset.Size, ShouldEqual, 10)

			So(store.name, ShouldEqual, "78640a1a-25c7-45a1-8c1f-cf8d3b162f9e-helloworld")
			So(store.length, ShouldEqual, 10)
			So(store.contentType, ShouldEqual, "plain/text")
			So(store.buf.String(), ShouldEqual, "I am a boy")

			So(resp.Body.String(), ShouldEqualJSON, `{
				"result": {
					"$type": "asset",
					"$name": "78640a1a-25c7-45a1-8c1f-cf8d3b162f9e-helloworld",
					"$url": "78640a1a-25c7-45a1-8c1f-cf8d3b162f9e-helloworld?signedurl=true",
					"$content_type":"plain/text"
				}
			}`)
		})

		Convey("errors missing content-type", func() {
			resp := r.PUT("asset", ``)
			So(resp.Body.String(), ShouldEqualJSON, `{
				"error": {
					"code": 108,
					"name": "InvalidArgument",
					"message": "Content-Type cannot be empty"
				}
			}`)
		})

		Convey("errors reading zero-byte body", func() {
			req, _ := http.NewRequest("PUT", "http://skygear.test/asset", strings.NewReader(``))
			req.Header.Set("Content-Type", "plain/text")
			resp := r.Do(req)

			So(resp.Body.String(), ShouldEqualJSON, `{
				"error": {
					"code": 108,
					"name": "InvalidArgument",
					"message": "Zero-byte content"
				}
			}`)
		})
	})
}

type naiveStoreSignatureParser struct {
	valid     bool
	signed    string
	name      string
	expiredAt time.Time
	asset.Store
}

func newNaiveStoreSignatureParser(assetStore asset.Store) *naiveStoreSignatureParser {
	return &naiveStoreSignatureParser{
		Store: assetStore,
	}
}

func (p *naiveStoreSignatureParser) SignedURL(name string) (string, error) {
	panic("this should not be called")
}

func (p *naiveStoreSignatureParser) IsSignatureRequired() bool {
	return true
}

func (p *naiveStoreSignatureParser) ParseSignature(signed string, name string, expiredAt time.Time) (valid bool, err error) {
	p.signed = signed
	p.name = name
	p.expiredAt = expiredAt

	return p.valid, nil
}

func TestGetFileHandler(t *testing.T) {
	Convey("GetFileHandler", t, func() {
		assetConn := &naiveAssetConn{}
		assetConn.savedAsset = map[string]*skydb.Asset{}

		store := newBufferedStore()
		signparser := newNaiveStoreSignatureParser(store)

		r := newmodGateway("(.+)")
		r.Handle("GET", &GetFileHandler{
			AssetStore: signparser,
		}, func(p *router.Payload) {
			p.DBConn = assetConn
		})

		Convey("GET a signed URL", func() {
			timeNow = func() time.Time {
				return time.Unix(1436431129, 999)
			}
			defer func() {
				timeNow = timeNowUTC
			}()
			signparser.valid = true
			assetConn.savedAsset["assetName"] = &skydb.Asset{
				Name:        "assetName",
				ContentType: "plain/text",
				Size:        10,
			}
			io.WriteString(store.buf, "I am a boy")

			resp := r.GET("assetName?signature=signedSignature&expiredAt=1436431130")
			So(resp.Body.String(), ShouldEqual, "I am a boy")
			So(resp.Header().Get("Content-Type"), ShouldEqual, "plain/text")
			So(resp.Header().Get("Content-Length"), ShouldEqual, "10")

			So(signparser.signed, ShouldEqual, "signedSignature")
			So(signparser.name, ShouldEqual, "assetName")
			So(signparser.expiredAt.Unix(), ShouldEqual, 1436431130)
		})

		Convey("errors if signature expired", func() {
			timeNow = func() time.Time {
				return time.Unix(1436431130, 1)
			}
			defer func() {
				timeNow = timeNowUTC
			}()

			resp := r.GET("assetName?signature=signedSignature&expiredAt=1436431130")
			So(resp.Body.String(), ShouldEqualJSON, `{
				"error": {
					"code": 102,
					"name": "PermissionDenied",
					"message": "Access denied"
				}
			}`)
		})

		Convey("errors on invalid signature", func() {
			timeNow = func() time.Time {
				return time.Unix(1436431129, 999)
			}
			defer func() {
				timeNow = timeNowUTC
			}()

			resp := r.GET("assetName?signature=signedSignature&expiredAt=1436431130")
			So(resp.Body.String(), ShouldEqualJSON, `{
				"error": {
					"code": 106,
					"name": "InvalidSignature",
					"message": "Invalid signature"
				}
			}`)
		})
	})
}
