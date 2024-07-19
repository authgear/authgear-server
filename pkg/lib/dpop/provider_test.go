package dpop

import (
	"encoding/json"
	"net/url"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/clock"
)

func TestDPoPProvider(t *testing.T) {
	Convey("validateProofJWT", t, func() {

		mockClock := clock.NewMockClockAt("2006-01-02T03:04:05Z")
		provider := &Provider{
			Clock: mockClock,
		}

		jwkJson, _ := json.Marshal(map[string]string{
			"kty": "EC",
			"alg": "ES256",
			"kid": "4cd07b3b-7290-42e8-9dff-c8097ad3c4cb",
			"crv": "P-256",
			"x":   "5XlfPlkbRwz7p3tuj0D7mE0AMP25a_1tfZHQH9jeC-o",
			"y":   "ZbkkKtJ4oFuz0HOjTH3R0JI_ccJdXwWwu6yeXePRS0s",
		})
		jwkKey, _ := jwk.ParseKey(jwkJson)

		Convey("ok", func() {
			header := jws.NewHeaders()
			header.Set("typ", "dpop+jwt")
			_ = header.Set("alg", "ES256")
			_ = header.Set("jwk", jwkKey)
			payload := jwt.New()
			_ = payload.Set("jti", "df352b68-9d3d-4006-a7fa-cf222a5f46b9")
			_ = payload.Set("htm", "POST")
			_ = payload.Set("htu", "http://example.com/path")
			_ = payload.Set("iat", mockClock.NowUTC().Unix())

			_, proof, err := provider.validateProofJWT(header, payload)
			So(err, ShouldBeNil)
			So(proof, ShouldNotBeNil)
			So(proof.JTI, ShouldEqual, "df352b68-9d3d-4006-a7fa-cf222a5f46b9")
			So(proof.HTM, ShouldEqual, "POST")
			expectedURL, _ := url.Parse("http://example.com/path")
			So(proof.HTU, ShouldEqual, expectedURL)
			So(proof.JKT, ShouldEqual, "1ePpVXfpPbb3o-3KXTh-T-a9Nc6mQ2qconuT5wq0OT4")
		})

		Convey("expired", func() {
			header := jws.NewHeaders()
			header.Set("typ", "dpop+jwt")
			_ = header.Set("alg", "ES256")
			_ = header.Set("jwk", jwkKey)
			payload := jwt.New()
			_ = payload.Set("jti", "df352b68-9d3d-4006-a7fa-cf222a5f46b9")
			_ = payload.Set("htm", "POST")
			_ = payload.Set("htu", "http://example.com/path")
			_ = payload.Set("iat", mockClock.NowUTC().Add(-301*time.Second).Unix())

			_, _, err := provider.validateProofJWT(header, payload)
			So(err, ShouldEqual, ErrProofExpired)
		})

		Convey("future jwt", func() {
			header := jws.NewHeaders()
			header.Set("typ", "dpop+jwt")
			_ = header.Set("alg", "ES256")
			_ = header.Set("jwk", jwkKey)
			payload := jwt.New()
			_ = payload.Set("jti", "df352b68-9d3d-4006-a7fa-cf222a5f46b9")
			_ = payload.Set("htm", "POST")
			_ = payload.Set("htu", "http://example.com/path")
			_ = payload.Set("iat", mockClock.NowUTC().Add(301*time.Second).Unix())

			_, _, err := provider.validateProofJWT(header, payload)
			So(err, ShouldEqual, ErrInvalidJwt)
		})

		Convey("invalid jwk", func() {
			header := jws.NewHeaders()
			header.Set("typ", "dpop+jwt")
			_ = header.Set("alg", "ES256")
			payload := jwt.New()
			_ = payload.Set("jti", "df352b68-9d3d-4006-a7fa-cf222a5f46b9")
			_ = payload.Set("htm", "POST")
			_ = payload.Set("htu", "http://example.com/path")
			_ = payload.Set("iat", mockClock.NowUTC().Unix())

			_, _, err := provider.validateProofJWT(header, payload)
			So(err, ShouldEqual, ErrInvalidJwk)
		})

		Convey("invalid typ", func() {
			header := jws.NewHeaders()
			header.Set("typ", "invalid")
			_ = header.Set("alg", "ES256")
			_ = header.Set("jwk", jwkKey)
			payload := jwt.New()
			_ = payload.Set("jti", "df352b68-9d3d-4006-a7fa-cf222a5f46b9")
			_ = payload.Set("htm", "POST")
			_ = payload.Set("htu", "http://example.com/path")
			_ = payload.Set("iat", mockClock.NowUTC().Unix())

			_, _, err := provider.validateProofJWT(header, payload)
			So(err, ShouldEqual, ErrInvalidJwtType)
		})

		Convey("no jti", func() {
			header := jws.NewHeaders()
			header.Set("typ", "dpop+jwt")
			_ = header.Set("alg", "ES256")
			_ = header.Set("jwk", jwkKey)
			payload := jwt.New()
			_ = payload.Set("htm", "POST")
			_ = payload.Set("htu", "http://example.com/path")
			_ = payload.Set("iat", mockClock.NowUTC().Unix())

			_, _, err := provider.validateProofJWT(header, payload)
			So(err, ShouldEqual, ErrInvalidJwtPayload)
		})

		Convey("no htm", func() {
			header := jws.NewHeaders()
			header.Set("typ", "dpop+jwt")
			_ = header.Set("alg", "ES256")
			_ = header.Set("jwk", jwkKey)
			payload := jwt.New()
			_ = payload.Set("jti", "df352b68-9d3d-4006-a7fa-cf222a5f46b9")
			_ = payload.Set("htu", "http://example.com/path")
			_ = payload.Set("iat", mockClock.NowUTC().Unix())

			_, _, err := provider.validateProofJWT(header, payload)
			So(err, ShouldEqual, ErrInvalidJwtPayload)
		})

		Convey("no htu", func() {
			header := jws.NewHeaders()
			header.Set("typ", "dpop+jwt")
			_ = header.Set("alg", "ES256")
			_ = header.Set("jwk", jwkKey)
			payload := jwt.New()
			_ = payload.Set("jti", "df352b68-9d3d-4006-a7fa-cf222a5f46b9")
			_ = payload.Set("htm", "POST")
			_ = payload.Set("iat", mockClock.NowUTC().Unix())

			_, _, err := provider.validateProofJWT(header, payload)
			So(err, ShouldEqual, ErrInvalidJwtPayload)
		})

		Convey("unsupported alg", func() {
			header := jws.NewHeaders()
			header.Set("typ", "dpop+jwt")
			_ = header.Set("jwk", jwkKey)
			_ = header.Set("alg", "UNKNOWN")
			payload := jwt.New()
			_ = payload.Set("jti", "df352b68-9d3d-4006-a7fa-cf222a5f46b9")
			_ = payload.Set("htm", "POST")
			_ = payload.Set("iat", mockClock.NowUTC().Unix())

			_, _, err := provider.validateProofJWT(header, payload)
			So(err, ShouldEqual, ErrUnsupportedAlg)
		})

	})
}
