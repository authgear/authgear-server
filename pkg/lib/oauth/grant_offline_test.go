package oauth

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestOfflineGrantToSession(t *testing.T) {
	Convey("OfflineGrant.ToSession", t, func() {
		now := time.Now().UTC()
		newHash := "rotated-hash"
		token := OfflineGrantRefreshToken{
			InitialTokenHash: "initial-hash",
			ClientID:         "client",
			CreatedAt:        now,
			Scopes:           []string{"openid", "read:orders"},
			AuthorizationID:  "authz-id",
			ResourceURI:      "https://api.example.com/orders",
			RotatedTokenHash: &newHash,
			RotatedAt:        &now,
		}
		grant := &OfflineGrant{
			ID:              "grant-id",
			InitialClientID: "client",
			RefreshTokens:   []OfflineGrantRefreshToken{token},
		}

		Convey("looking up by the rotated (current) hash preserves ResourceURI", func() {
			session, ok := grant.ToSession(newHash)
			So(ok, ShouldBeTrue)
			So(session.ResourceURI, ShouldEqual, "https://api.example.com/orders")
		})

		Convey("looking up by the initial hash (pre-rotation) also preserves ResourceURI", func() {
			session, ok := grant.ToSession("initial-hash")
			So(ok, ShouldBeTrue)
			So(session.ResourceURI, ShouldEqual, "https://api.example.com/orders")
		})

		Convey("a token with no resource bound has an empty ResourceURI", func() {
			plainGrant := &OfflineGrant{
				ID:              "grant-id-2",
				InitialClientID: "client",
				RefreshTokens: []OfflineGrantRefreshToken{{
					InitialTokenHash: "plain-hash",
					ClientID:         "client",
					CreatedAt:        now,
				}},
			}
			session, ok := plainGrant.ToSession("plain-hash")
			So(ok, ShouldBeTrue)
			So(session.ResourceURI, ShouldEqual, "")
		})
	})
}
