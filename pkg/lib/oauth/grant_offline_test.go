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

func TestOfflineGrantScopesByResourceURI(t *testing.T) {
	Convey("OfflineGrant.ScopesByResourceURI", t, func() {
		const orders = "https://api.example.com/orders"
		const inventory = "https://api.example.com/inventory"

		Convey("a legacy grant with only grant-level scopes is unbound", func() {
			grant := &OfflineGrant{
				InitialClientID:   "mobile",
				Deprecated_Scopes: []string{"openid", "offline_access"},
			}
			So(grant.ScopesByResourceURI("mobile"), ShouldResemble, map[string][]string{
				"": {"openid", "offline_access"},
			})
		})

		Convey("refresh tokens bound to different resources are grouped", func() {
			grant := &OfflineGrant{
				InitialClientID: "mobile",
				RefreshTokens: []OfflineGrantRefreshToken{
					{ClientID: "mobile", Scopes: []string{"openid", "offline_access"}},
					{ClientID: "mobile", Scopes: []string{"openid", "read:orders"}, ResourceURI: orders},
					{ClientID: "mobile", Scopes: []string{"read:orders", "write:orders"}, ResourceURI: orders},
					{ClientID: "mobile", Scopes: []string{"read:stock"}, ResourceURI: inventory},
				},
			}
			So(grant.ScopesByResourceURI("mobile"), ShouldResemble, map[string][]string{
				"":        {"openid", "offline_access"},
				orders:    {"openid", "read:orders", "write:orders"},
				inventory: {"read:stock"},
			})
		})

		Convey("refresh tokens of another client are excluded", func() {
			grant := &OfflineGrant{
				InitialClientID: "mobile",
				RefreshTokens: []OfflineGrantRefreshToken{
					{ClientID: "mobile", Scopes: []string{"openid"}},
					{ClientID: "web", Scopes: []string{"read:orders"}, ResourceURI: orders},
				},
			}
			So(grant.ScopesByResourceURI("mobile"), ShouldResemble, map[string][]string{
				"": {"openid"},
			})
			So(grant.ScopesByResourceURI("other"), ShouldResemble, map[string][]string{})
		})
	})
}
