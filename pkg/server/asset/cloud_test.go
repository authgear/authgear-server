package asset

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCloudStoreCreation(t *testing.T) {
	testServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			method := r.Method
			authHeader := r.Header.Get("Authorization")

			if method == http.MethodGet &&
				path == "/token/testapp" &&
				authHeader == "Bearer correct-auth-token" {

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
          "value": "correct-signer-token",
          "extra": "correct-signer-extra",
          "expired_at": "2016-08-25T10:19:27Z"
        }`))
			} else {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Bad Request"))
			}
		}),
	)

	Convey("Cloud Store Creation", t, func() {
		Convey("Success on normal flow (public)", func() {
			_store, err := NewCloudStore(
				"testapp",
				testServer.URL,
				"correct-auth-token",
				"http://localhost:12345/public",
				"http://localhost:12345/private",
				true,
			)
			store := _store.(*cloudStore)
			defer store.signer.refreshTicker.Stop()

			So(err, ShouldBeNil)
			So(store, ShouldNotBeNil)
			So(store.appName, ShouldEqual, "testapp")
			So(store.host, ShouldEqual, testServer.URL)
			So(store.authToken, ShouldEqual, "correct-auth-token")
			So(store.urlPrefix, ShouldEqual, "http://localhost:12345/public")
			So(store.public, ShouldBeTrue)

			time.Sleep(100 * time.Millisecond)
			So(store.signer.token, ShouldEqual, "correct-signer-token")
			So(store.signer.extra, ShouldEqual, "correct-signer-extra")
			So(store.signer.expiredAt.Unix(), ShouldEqual, 1472120367)
		})

		Convey("Success on normal flow (private)", func() {
			_store, err := NewCloudStore(
				"testapp",
				testServer.URL,
				"correct-auth-token",
				"http://localhost:12345/public",
				"http://localhost:12345/private",
				false,
			)
			store := _store.(*cloudStore)
			defer store.signer.refreshTicker.Stop()

			So(err, ShouldBeNil)
			So(store, ShouldNotBeNil)
			So(store.appName, ShouldEqual, "testapp")
			So(store.host, ShouldEqual, testServer.URL)
			So(store.authToken, ShouldEqual, "correct-auth-token")
			So(store.urlPrefix, ShouldEqual, "http://localhost:12345/private")
			So(store.public, ShouldBeFalse)

			time.Sleep(100 * time.Millisecond)
			So(store.signer.token, ShouldEqual, "correct-signer-token")
			So(store.signer.extra, ShouldEqual, "correct-signer-extra")
			So(store.signer.expiredAt.Unix(), ShouldEqual, 1472120367)
		})

		Convey("Fail when no app name", func() {
			store, err := NewCloudStore(
				"",
				testServer.URL,
				"correct-auth-token",
				"http://localhost:12345/public",
				"http://localhost:12345/private",
				true,
			)

			So(err, ShouldNotBeNil)
			So(store, ShouldBeNil)
		})

		Convey("Fail when no host", func() {
			store, err := NewCloudStore(
				"testapp",
				"",
				"correct-auth-token",
				"http://localhost:12345/public",
				"http://localhost:12345/private",
				true,
			)

			So(err, ShouldNotBeNil)
			So(store, ShouldBeNil)
		})

		Convey("Fail when no auth token", func() {
			store, err := NewCloudStore(
				"testapp",
				testServer.URL,
				"",
				"http://localhost:12345/public",
				"http://localhost:12345/private",
				true,
			)

			So(err, ShouldNotBeNil)
			So(store, ShouldBeNil)
		})

		Convey("Fail when no public url when needed", func() {
			store, err := NewCloudStore(
				"testapp",
				testServer.URL,
				"correct-auth-token",
				"",
				"http://localhost:12345/private",
				true,
			)

			So(err, ShouldNotBeNil)
			So(store, ShouldBeNil)
		})

		Convey("Fail when no private url when needed", func() {
			store, err := NewCloudStore(
				"testapp",
				testServer.URL,
				"correct-auth-token",
				"http://localhost:12345/public",
				"",
				false,
			)

			So(err, ShouldNotBeNil)
			So(store, ShouldBeNil)
		})
	})
}

func TestCloudStoreGetSignedURL(t *testing.T) {
	testServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			method := r.Method
			authHeader := r.Header.Get("Authorization")

			if method == http.MethodGet &&
				path == "/token/testapp" &&
				authHeader == "Bearer correct-auth-token" {

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
        "value": "correct-signer-token",
        "extra": "correct-signer-extra",
        "expired_at": "2016-08-25T10:19:27Z"
      }`))
			} else {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Bad Request"))
			}
		}),
	)

	Convey("Cloud Store Get Signed URL", t, func() {
		Convey("Success on public store", func() {
			_publicStore, _ := NewCloudStore(
				"testapp",
				testServer.URL,
				"correct-auth-token",
				"http://localhost:12345/public",
				"http://localhost:12345/private",
				true,
			)
			publicStore := _publicStore.(*cloudStore)
			defer publicStore.signer.refreshTicker.Stop()

			time.Sleep(100 * time.Millisecond)
			signedURL, err := publicStore.SignedURL("file:%251")
			So(err, ShouldBeNil)
			So(
				signedURL,
				ShouldEqual,
				"http://localhost:12345/public/testapp/file:%25251",
			)
		})

		Convey("Success on private store", func() {
			_publicStore, _ := NewCloudStore(
				"testapp",
				testServer.URL,
				"correct-auth-token",
				"http://localhost:12345/public",
				"http://localhost:12345/private",
				false,
			)
			publicStore := _publicStore.(*cloudStore)
			defer publicStore.signer.refreshTicker.Stop()

			time.Sleep(100 * time.Millisecond)
			signedURL, err := publicStore.SignedURL("file:%251")
			So(err, ShouldBeNil)

			parsed, err := url.Parse(signedURL)
			So(err, ShouldBeNil)
			So(parsed, ShouldNotBeNil)
			So(parsed.Host, ShouldEqual, "localhost:12345")
			So(parsed.Path, ShouldEqual, "/private/testapp/file:%251")

			expiredAtString := parsed.Query().Get("expired_at")
			So(expiredAtString, ShouldNotBeEmpty)

			targetExpiredAtUnix := time.Now().Add(15 * time.Minute).Unix()
			expiredAtUnix, err := strconv.ParseInt(expiredAtString, 10, 64)
			So(err, ShouldBeNil)
			So(
				expiredAtUnix,
				ShouldBeBetween,
				targetExpiredAtUnix-100,
				targetExpiredAtUnix+100,
			)

			signature := parsed.Query().Get("signature")
			So(signature, ShouldNotBeEmpty)
		})
	})
}

func TestCloudStoreGeneratePostFileRequest(t *testing.T) {
	Convey("Generate Post File Request", t, func(c C) {
		testServer := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("Bad Request"))
					return
				}

				reqBody, err := ioutil.ReadAll(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("Fails on reading request body"))
					return
				}

				reqeustData := struct {
					ContentType string `json:"content-type"`
					ContentSize int    `json:"content-size"`
				}{}

				err = json.Unmarshal(reqBody, &reqeustData)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("Fails when unmarshaling data"))
					return
				}

				responseData := struct {
					Action      string      `json:"action"`
					ExtraFields interface{} `json:"extra-fields"`
				}{
					Action: "http://skygear.dev/files/file001",
					ExtraFields: struct {
						ExtraField1 string `json:"X-Extra-Field-1"`
						ExtraField2 string `json:"X-Extra-Field-2"`
						RequestPath string `json:"X-Request-Path"`
						AuthHeader  string `json:"X-Auth-Header"`
						ContentType string `json:"X-Content-Type"`
						ContentSize int    `json:"X-Content-Size"`
					}{
						ExtraField1: "extra-value-1",
						ExtraField2: "extra-value-2",
						RequestPath: r.URL.Path,
						AuthHeader:  r.Header.Get("Authorization"),
						ContentType: reqeustData.ContentType,
						ContentSize: reqeustData.ContentSize,
					},
				}

				responseBytes, err := json.Marshal(responseData)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("Fails when marshaling data"))
					return
				}

				w.WriteHeader(http.StatusOK)
				w.Write(responseBytes)
			}),
		)

		Convey("Success on normal flow", func() {
			store := &cloudStore{
				appName:   "testapp",
				host:      testServer.URL,
				authToken: "correct-auth-token",
				urlPrefix: "http://localhost:12345/public",
				public:    true,
			}
			postRequest, err := store.GeneratePostFileRequest(
				"file:%251",
				"application/pdf",
				24,
			)

			So(err, ShouldBeNil)
			So(postRequest, ShouldNotBeNil)

			So(postRequest.Action, ShouldEqual, "http://skygear.dev/files/file001")
			So(
				postRequest.ExtraFields["X-Extra-Field-1"],
				ShouldEqual,
				"extra-value-1",
			)
			So(
				postRequest.ExtraFields["X-Extra-Field-2"],
				ShouldEqual,
				"extra-value-2",
			)
			So(
				postRequest.ExtraFields["X-Request-Path"],
				ShouldEqual,
				"/asset/testapp/file:%251",
			)
			So(
				postRequest.ExtraFields["X-Auth-Header"],
				ShouldEqual,
				"Bearer correct-auth-token",
			)
			So(
				postRequest.ExtraFields["X-Content-Type"],
				ShouldEqual,
				"application/pdf",
			)
			So(
				postRequest.ExtraFields["X-Content-Size"],
				ShouldEqual,
				24,
			)
		})
	})
}
