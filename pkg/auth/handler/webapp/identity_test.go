package webapp_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

func TestIdentitiesDisplayName(t *testing.T) {
	emailIdentity := &identity.Info{
		Type: authn.IdentityTypeLoginID,
		Claims: map[string]interface{}{
			identity.IdentityClaimLoginIDOriginalValue: "user@example.com",
		},
	}

	oauthProviderIdentity := &identity.Info{
		Type: authn.IdentityTypeOAuth,
		Claims: map[string]interface{}{
			identity.IdentityClaimOAuthProviderType: "provider",
			identity.StandardClaimEmail:             "user@oauth-provider.com",
		},
	}

	oauthProviderIdentityWithStandardClaims := &identity.Info{
		Type: authn.IdentityTypeOAuth,
		Claims: map[string]interface{}{
			identity.IdentityClaimOAuthProviderType: "provider2",
		},
	}

	anonymousIdentity := &identity.Info{
		Type: authn.IdentityTypeAnonymous,
	}

	biometricIdentity := &identity.Info{
		Type: authn.IdentityTypeBiometric,
	}

	Convey("identitiesDisplayName", t, func() {
		displayName := webapp.IdentitiesDisplayName([]*identity.Info{
			anonymousIdentity,
			oauthProviderIdentity,
			biometricIdentity,
			emailIdentity,
		})
		So(displayName, ShouldEqual, "user@example.com")

		displayName = webapp.IdentitiesDisplayName([]*identity.Info{
			anonymousIdentity,
			oauthProviderIdentity,
			biometricIdentity,
		})
		So(displayName, ShouldEqual, "provider:user@oauth-provider.com")

		displayName = webapp.IdentitiesDisplayName([]*identity.Info{
			oauthProviderIdentityWithStandardClaims,
		})
		So(displayName, ShouldEqual, "provider2")

	})
}
