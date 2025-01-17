package apple

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

	"github.com/authgear/authgear-server/pkg/util/clock"
)

func TestAppleImpl(t *testing.T) {
	Convey("AppleImpl.GetAuthorizationURL", t, func() {
		deps := oauthrelyingparty.Dependencies{
			ProviderConfig: oauthrelyingparty.ProviderConfig{
				"client_id": "client_id",
				"type":      Type,
			},
		}
		g := Apple{}

		ctx := context.Background()
		u, err := g.GetAuthorizationURL(ctx, deps, oauthrelyingparty.GetAuthorizationURLOptions{
			RedirectURI:  "https://localhost/",
			ResponseMode: oauthrelyingparty.ResponseModeFormPost,
			Nonce:        "nonce",
			State:        "state",
			Prompt:       []string{"login"},
		})
		So(err, ShouldBeNil)
		So(u, ShouldEqual, "https://appleid.apple.com/auth/authorize?client_id=client_id&nonce=nonce&redirect_uri=https%3A%2F%2Flocalhost%2F&response_mode=form_post&response_type=code&scope=name+email&state=state")
	})

	Convey("AppleImpl.createClientSecret", t, func() {
		deps := oauthrelyingparty.Dependencies{
			ProviderConfig: oauthrelyingparty.ProviderConfig{
				"client_id": "the_client_id",
				"type":      Type,
				"team_id":   "the_team_id",
				"key_id":    "the_key_id",
			},
			// Generated with the following command
			//   openssl genpkey -algorithm EC -pkeyopt ec_paramgen_curve:P-256 -out -
			//
			// In case you wonder why it is a P-256 key, it is observed that Apple generates such keys.
			// See https://developer.apple.com/help/account/manage-keys/create-a-private-key/
			ClientSecret: `-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgnDWXkNs9pRnFZwkm
miwAePJd5JPUey25Bo8yNPPTovihRANCAARk0V61v/iATyYj3Qbj9ayQzDEVMAwp
UyS+h/UyCBBNs4aRFSL76tZaeGAmGa62GINnZ4UH4etxoLa4PvNnc77t
-----END PRIVATE KEY-----
`,
			Clock: clock.NewMockClockAt("2025-01-17T13:32:00+08:00"),
		}
		g := Apple{}

		clientSecret, err := g.createClientSecret(deps)
		So(err, ShouldBeNil)
		// The signature algorithm is ES256, which is ECDSA with P-256 and SHA256, according to https://datatracker.ietf.org/doc/html/rfc7518#section-3.1
		// ECDSA, by definition, use a cryptographically secure random number.
		// You can see this nature by looking at the signature of https://pkg.go.dev/crypto/ecdsa#Sign
		// So the signature is different in every generation.
		// What we want to assert here is the header and the payload is of an expected shape.
		So(clientSecret, ShouldStartWith, "eyJhbGciOiJFUzI1NiIsImtpZCI6InRoZV9rZXlfaWQiLCJ0eXAiOiJKV1QifQ.eyJhdWQiOiJodHRwczovL2FwcGxlaWQuYXBwbGUuY29tIiwiZXhwIjoxNzM3MDkyMjIwLCJpYXQiOjE3MzcwOTE5MjAsImlzcyI6InRoZV90ZWFtX2lkIiwic3ViIjoidGhlX2NsaWVudF9pZCJ9.")
	})
}
