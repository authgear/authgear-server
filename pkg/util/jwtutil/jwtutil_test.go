package jwtutil

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/lestrrat-go/jwx/jwt"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSign(t *testing.T) {
	// testToken ignores signature because the private key is generated freshly every time.
	testToken := func(actual []byte, expected string) {
		actualHdr, actualPayload, _, err := jws.SplitCompact(actual)
		So(err, ShouldBeNil)

		expectedHdr, expectedPayload, _, err := jws.SplitCompact([]byte(expected))
		So(err, ShouldBeNil)

		So(string(actualHdr), ShouldEqual, string(expectedHdr))
		So(string(actualPayload), ShouldEqual, string(expectedPayload))
	}

	Convey("Sign with RSA jwk.Key with kid", t, func() {
		payload := jwt.New()
		_ = payload.Set("foobar", 42)

		alg := jwa.RS256
		// nolint: gosec
		privKey, err := rsa.GenerateKey(rand.Reader, 512)
		So(err, ShouldBeNil)

		jwkKey, err := jwk.New(privKey)
		So(err, ShouldBeNil)
		_ = jwkKey.Set("kid", "mykey")

		token, err := Sign(payload, alg, jwkKey)
		So(err, ShouldBeNil)

		testToken(token, "eyJhbGciOiJSUzI1NiIsImtpZCI6Im15a2V5IiwidHlwIjoiSldUIn0.eyJmb29iYXIiOjQyfQ.ViPT48rCUICElq8_9puYyDvKl_3X0Rg6jfSCPv-RsD1jmVsGMBKQYYS1CKEhv_ke3N_9MFMEK7aPR1GHlKOFPg")
	})

	Convey("Sign with RSA jwk.Key WITHOUT kid", t, func() {
		payload := jwt.New()
		_ = payload.Set("foobar", 42)

		alg := jwa.RS256
		// nolint: gosec
		privKey, err := rsa.GenerateKey(rand.Reader, 512)
		So(err, ShouldBeNil)

		jwkKey, err := jwk.New(privKey)
		So(err, ShouldBeNil)

		token, err := Sign(payload, alg, jwkKey)
		So(err, ShouldBeNil)

		testToken(token, "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJmb29iYXIiOjQyfQ.D6VFddNrBxi-fdrq8-44cJxebuy0u1KS0bViZBi8kVBWELFDdzMXw42l0W4bI-4h6FWyDWCj-xGTxaakHqSC9w")
	})

	Convey("Sign with ECDSA jwk.Key with kid", t, func() {
		payload := jwt.New()
		_ = payload.Set("foobar", 42)

		alg := jwa.ES256
		privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		So(err, ShouldBeNil)

		jwkKey, err := jwk.New(privKey)
		So(err, ShouldBeNil)

		token, err := Sign(payload, alg, jwkKey)
		So(err, ShouldBeNil)

		testToken(token, "eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJmb29iYXIiOjQyfQ.10fPqnmUD9mVVO_SGwLAKrQGijv5_mAYJV-mq6w6wM1h5DCgUPlofdKdMETUJIp-rhWwVzxzWM0u4MgpNPOixA")
	})

	Convey("Sign with raw key", t, func() {
		payload := jwt.New()
		_ = payload.Set("foobar", 42)

		alg := jwa.RS256
		// nolint: gosec
		privKey, err := rsa.GenerateKey(rand.Reader, 512)
		So(err, ShouldBeNil)

		token, err := Sign(payload, alg, privKey)
		So(err, ShouldBeNil)

		testToken(token, "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJmb29iYXIiOjQyfQ.invw9DyIBZOoTzQYZ8izM_cnLOsEJBrFAClHo36Fzv7OgV6uq25zXs3RJhRicmYO-_77Ck8LV0BZ_aC6pue67g")
	})

	Convey("Sign with octet key", t, func() {
		payload := jwt.New()
		_ = payload.Set("foobar", 42)

		alg := jwa.HS256
		key := []byte("secret")

		token, err := Sign(payload, alg, key)
		So(err, ShouldBeNil)

		testToken(token, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJmb29iYXIiOjQyfQ.ZrqY6Am7ejb_WrRBSP0EsWXyaFBHpfQTFWHoQSb_RNc")
	})
}

func TestSplitWithoutVerify(t *testing.T) {
	Convey("SplitWithoutVerify", t, func() {
		compact := []byte("eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJmb29iYXIiOjQyfQ.invw9DyIBZOoTzQYZ8izM_cnLOsEJBrFAClHo36Fzv7OgV6uq25zXs3RJhRicmYO-_77Ck8LV0BZ_aC6pue67g")

		hdr, payload, err := SplitWithoutVerify(compact)
		So(err, ShouldBeNil)

		hdrBytes, err := json.Marshal(hdr)
		So(err, ShouldBeNil)
		So(string(hdrBytes), ShouldEqualJSON, `
		{
			"typ": "JWT",
			"alg": "RS256"
		}
		`)

		payloadBytes, err := json.Marshal(payload)
		So(err, ShouldBeNil)
		So(string(payloadBytes), ShouldEqualJSON, `
		{
			"foobar": 42
		}
		`)
	})
}

func TestBuildFromMap(t *testing.T) {
	// testToken ignores signature because the private key is generated freshly every time.
	testToken := func(actual []byte, expected string) {
		actualHdr, actualPayload, _, err := jws.SplitCompact(actual)
		So(err, ShouldBeNil)

		expectedHdr, expectedPayload, _, err := jws.SplitCompact([]byte(expected))
		So(err, ShouldBeNil)

		So(string(actualHdr), ShouldEqual, string(expectedHdr))
		So(string(actualPayload), ShouldEqual, string(expectedPayload))
	}

	Convey("BuildFromMap", t, func() {
		m := map[string]interface{}{
			"https://example.com": map[string]interface{}{
				"foo": "bar",
			},
		}

		payload, err := BuildFromMap(m)
		So(err, ShouldBeNil)

		alg := jwa.HS256
		key := []byte("secret")

		token, err := Sign(payload, alg, key)
		So(err, ShouldBeNil)

		testToken(token, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJodHRwczovL2V4YW1wbGUuY29tIjp7ImZvbyI6ImJhciJ9fQ.HkXJRxSFoHUHvk1nZ36rcf4ZJDuKd2xExdjKIEv48dw")
	})

	Convey("ToMap", t, func() {
		b := jwt.NewBuilder()
		b.Expiration(time.Date(2006, 1, 2, 3, 4, 5, 6, time.UTC))
		payload, err := b.Build()
		So(err, ShouldBeNil)

		actual, err := ToMap(payload)
		So(err, ShouldBeNil)

		expected := map[string]interface{}{
			"exp": 1136171045.0,
		}

		So(actual, ShouldResemble, expected)
	})
}
