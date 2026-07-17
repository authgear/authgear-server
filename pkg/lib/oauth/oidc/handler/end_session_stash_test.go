package handler

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc/protocol"
)

func TestEndSessionStash(t *testing.T) {
	Convey("sealEndSessionRequest and openEndSessionRequest", t, func() {
		req := protocol.EndSessionRequest{
			"id_token_hint":            "some-id-token",
			"post_logout_redirect_uri": "https://example.com/after-logout",
			"state":                    "some-state",
		}

		Convey("should round-trip to the original request", func() {
			key, sealed, err := sealEndSessionRequest(req)
			So(err, ShouldBeNil)

			opened, err := openEndSessionRequest(key, sealed)
			So(err, ShouldBeNil)
			So(opened, ShouldResemble, req)
		})

		Convey("should generate fresh key and sealed value on every call", func() {
			key1, sealed1, err := sealEndSessionRequest(req)
			So(err, ShouldBeNil)

			key2, sealed2, err := sealEndSessionRequest(req)
			So(err, ShouldBeNil)

			So(key1, ShouldNotEqual, key2)
			So(sealed1, ShouldNotEqual, sealed2)
		})

		Convey("should fail if the key does not match the sealed value", func() {
			key1, _, err := sealEndSessionRequest(req)
			So(err, ShouldBeNil)

			_, sealed2, err := sealEndSessionRequest(req)
			So(err, ShouldBeNil)

			_, err = openEndSessionRequest(key1, sealed2)
			So(err, ShouldEqual, ErrEndSessionStashInvalid)
		})

		Convey("should fail if the sealed value is truncated", func() {
			key, sealed, err := sealEndSessionRequest(req)
			So(err, ShouldBeNil)

			truncated := sealed[:len(sealed)-4]
			_, err = openEndSessionRequest(key, truncated)
			So(err, ShouldEqual, ErrEndSessionStashInvalid)
		})

		Convey("should fail if the key is malformed base64", func() {
			_, sealed, err := sealEndSessionRequest(req)
			So(err, ShouldBeNil)

			_, err = openEndSessionRequest("not-valid-base64!!!", sealed)
			So(err, ShouldEqual, ErrEndSessionStashInvalid)
		})

		Convey("should fail if the sealed value is malformed base64", func() {
			key, _, err := sealEndSessionRequest(req)
			So(err, ShouldBeNil)

			_, err = openEndSessionRequest(key, "not-valid-base64!!!")
			So(err, ShouldEqual, ErrEndSessionStashInvalid)
		})

		Convey("should fail if the key has the wrong length for AES", func() {
			_, sealed, err := sealEndSessionRequest(req)
			So(err, ShouldBeNil)

			_, err = openEndSessionRequest("dG9vc2hvcnQ", sealed)
			So(err, ShouldEqual, ErrEndSessionStashInvalid)
		})
	})
}
