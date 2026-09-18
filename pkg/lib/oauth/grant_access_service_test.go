package oauth

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
)

func TestIssueAccessGrantResultWriteTo(t *testing.T) {
	Convey("IssueAccessGrantResult.WriteTo", t, func() {
		Convey("always writes scope, matching what client_credentials already does", func() {
			r := &IssueAccessGrantResult{
				Token:     "a-token",
				TokenType: "Bearer",
				ExpiresIn: 1800,
				Scopes:    []string{"openid", "read:orders"},
			}
			resp := protocol.TokenResponse{}
			r.WriteTo(resp)
			So(resp["scope"], ShouldEqual, "openid read:orders")
			So(resp["access_token"], ShouldEqual, "a-token")
		})

		Convey("an empty scope list writes an empty string, not an omitted key", func() {
			r := &IssueAccessGrantResult{Token: "a-token", TokenType: "Bearer"}
			resp := protocol.TokenResponse{}
			r.WriteTo(resp)
			So(resp["scope"], ShouldEqual, "")
		})

		Convey("a nil receiver or nil response is a no-op", func() {
			var r *IssueAccessGrantResult
			So(func() { r.WriteTo(protocol.TokenResponse{}) }, ShouldNotPanic)

			r = &IssueAccessGrantResult{Scopes: []string{"openid"}}
			So(func() { r.WriteTo(nil) }, ShouldNotPanic)
		})
	})
}
