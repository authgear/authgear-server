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

	"github.com/oursky/ourd/asset"
	"github.com/oursky/ourd/oddb"
	. "github.com/oursky/ourd/ourtest"
	"github.com/oursky/ourd/router"
	. "github.com/smartystreets/goconvey/convey"
)

// a thin wrapper on router.Gateway with helper methods to invoke request more
// easily
type modGateway router.Gateway

func newmodGateway(pattern string) *modGateway {
	return (*modGateway)(router.NewGateway(pattern))
}

func (g *modGateway) Handle(method string, handler router.Handler, prepareFunc func(*router.Payload)) {
	(*router.Gateway)(g).Handle(method, handler, func(p *router.Payload, _ *router.Response) int {
		prepareFunc(p)
		return 200
	})
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
	path = "http://ourd.test/" + path
	req, _ := http.NewRequest(method, path, strings.NewReader(body))
	resp := httptest.NewRecorder()

	(*router.Gateway)(g).ServeHTTP(resp, req)
	return resp
}

type naiveAssetConn struct {
	asset oddb.Asset
	oddb.Conn
}

func (c *naiveAssetConn) GetAsset(name string, asset *oddb.Asset) error {
	*asset = c.asset
	return nil
}

func (c *naiveAssetConn) SaveAsset(asset *oddb.Asset) error {
	c.asset = *asset
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

func TestAssetUploadURLHandler(t *testing.T) {
	Convey("AssetUploadURLHandler", t, func() {
		assetConn := &naiveAssetConn{}
		store := newBufferedStore()

		r := newmodGateway("(.+)")
		r.Handle("PUT", AssetUploadURLHandler, func(p *router.Payload) {
			p.DBConn = assetConn
			p.AssetStore = store
		})

		Convey("uploads a file", func() {
			uuidNew = func() string {
				return "f28c4037-uuid-4d0a-94d6-2206ab371d6c"
			}

			req, _ := http.NewRequest("PUT", "http://ourd.test/asset", strings.NewReader(``))
			req.Header.Set("Content-Type", "plain/text")
			req.Body = ioutil.NopCloser(strings.NewReader(`I am a boy`))

			resp := r.Do(req)

			So(assetConn.asset, ShouldResemble, oddb.Asset{
				Name:        "f28c4037-uuid-4d0a-94d6-2206ab371d6c-asset",
				ContentType: "plain/text",
				Size:        10,
			})

			So(store.buf.String(), ShouldEqual, "I am a boy")
			So(store.name, ShouldEqual, "f28c4037-uuid-4d0a-94d6-2206ab371d6c-asset")
			So(store.length, ShouldEqual, 10)
			So(store.contentType, ShouldEqual, "plain/text")

			So(resp.Body.String(), ShouldEqualJSON, `{
				"result": {
					"$type": "asset",
					"$name": "f28c4037-uuid-4d0a-94d6-2206ab371d6c-asset"
				}
			}`)
		})

		Convey("errors missing content-type", func() {
			resp := r.PUT("asset", ``)
			So(resp.Body.String(), ShouldEqualJSON, `{
				"error": {
					"type": "RequestInvalid",
					"code": 101,
					"message": "Content-Type cannot be empty"
				}
			}`)
		})

		Convey("errors reading zero-byte body", func() {
			req, _ := http.NewRequest("PUT", "http://ourd.test/asset", strings.NewReader(``))
			req.Header.Set("Content-Type", "plain/text")
			resp := r.Do(req)

			So(resp.Body.String(), ShouldEqualJSON, `{
				"error": {
					"type": "RequestInvalid",
					"code": 101,
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

func (p *naiveStoreSignatureParser) ParseSignature(signed string, name string, expiredAt time.Time) (valid bool, err error) {
	p.signed = signed
	p.name = name
	p.expiredAt = expiredAt

	return p.valid, nil
}

func TestAssetGetURLHandler(t *testing.T) {
	Convey("AssetGetURLHandler", t, func() {
		assetConn := &naiveAssetConn{}
		store := newBufferedStore()
		signparser := newNaiveStoreSignatureParser(store)

		r := newmodGateway("(.+)")
		r.Handle("GET", AssetGetURLHandler, func(p *router.Payload) {
			p.DBConn = assetConn
			p.AssetStore = signparser
		})

		Convey("GET a signed URL", func() {
			timeNow = func() time.Time {
				return time.Unix(1436431129, 999)
			}
			defer func() {
				timeNow = time.Now
			}()
			signparser.valid = true
			assetConn.asset = oddb.Asset{
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
				timeNow = time.Now
			}()

			resp := r.GET("assetName?signature=signedSignature&expiredAt=1436431130")
			So(resp.Body.String(), ShouldEqualJSON, `{
				"error": {
					"type": "RequestInvalid",
					"code": 101,
					"message": "Access denied"
				}
			}`)
		})

		Convey("errors on invalid signature", func() {
			timeNow = func() time.Time {
				return time.Unix(1436431129, 999)
			}
			defer func() {
				timeNow = time.Now
			}()

			resp := r.GET("assetName?signature=signedSignature&expiredAt=1436431130")
			So(resp.Body.String(), ShouldEqualJSON, `{
				"error": {
					"type": "RequestInvalid",
					"code": 101,
					"message": "Invalid signature"
				}
			}`)
		})
	})
}
