package resourcescope

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestCommandsAddResourceToClientID(t *testing.T) {
	Convey("Commands.AddResourceToClientID", t, func() {
		ctx := context.Background()

		newConfig := func(clientID string, appType config.OAuthClientApplicationType) *config.OAuthConfig {
			return &config.OAuthConfig{
				Clients: []config.OAuthClientConfig{
					{
						ClientID:        clientID,
						ApplicationType: appType,
					},
				},
			}
		}

		Convey("an unknown client ID returns ErrClientNotFound before any client-type check", func() {
			c := &Commands{OAuthConfig: newConfig("known-client", config.OAuthClientApplicationTypeConfidential)}
			err := c.AddResourceToClientID(ctx, "https://api.example.com", "unknown-client")
			So(err, ShouldEqual, ErrClientNotFound)
		})

		disallowed := []config.OAuthClientApplicationType{
			config.OAuthClientApplicationTypeSPA,
			config.OAuthClientApplicationTypeTraditionalWeb,
			config.OAuthClientApplicationTypeNative,
			config.OAuthClientApplicationTypeThirdPartyApp,
			config.OAuthClientApplicationTypeDynamicThirdParty,
		}
		for _, appType := range disallowed {
			Convey("a "+string(appType)+" client returns ErrClientCannotBeAssociatedWithResource, distinct from ErrClientNotFound", func() {
				c := &Commands{OAuthConfig: newConfig("the-client", appType)}
				err := c.AddResourceToClientID(ctx, "https://api.example.com", "the-client")
				So(err, ShouldEqual, ErrClientCannotBeAssociatedWithResource)
				So(err, ShouldNotEqual, ErrClientNotFound)
			})
		}

		allowed := []config.OAuthClientApplicationType{
			config.OAuthClientApplicationTypeConfidential,
			config.OAuthClientApplicationTypeM2M,
		}
		for _, appType := range allowed {
			Convey("a "+string(appType)+" client passes the client-type check and reaches the store", func() {
				// Store is nil here on purpose: reaching past both checks
				// means dereferencing it, so the panic (instead of one of
				// the two sentinel errors) proves the gate let it through.
				c := &Commands{OAuthConfig: newConfig("the-client", appType)}
				So(func() {
					_ = c.AddResourceToClientID(ctx, "https://api.example.com", "the-client")
				}, ShouldPanic)
			})
		}
	})
}
