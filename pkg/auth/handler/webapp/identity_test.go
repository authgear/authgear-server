package webapp_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/api/oauthrelyingparty"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

func TestIdentitiesDisplayName(t *testing.T) {
	emailIdentity := &identity.Info{
		Type: model.IdentityTypeLoginID,
		LoginID: &identity.LoginID{
			OriginalLoginID: "user@example.com",
		},
	}

	oauthProviderIdentity := &identity.Info{
		Type: model.IdentityTypeOAuth,
		OAuth: &identity.OAuth{
			ProviderID: oauthrelyingparty.ProviderID{
				Type: "provider",
			},
			Claims: map[string]interface{}{
				"email": "user@oauth-provider.com",
			},
		},
	}

	oauthProviderIdentityWithStandardClaims := &identity.Info{
		Type: model.IdentityTypeOAuth,
		OAuth: &identity.OAuth{
			ProviderID: oauthrelyingparty.ProviderID{
				Type: "provider2",
			},
			Claims: map[string]interface{}{},
		},
	}

	anonymousIdentity := &identity.Info{
		Type: model.IdentityTypeAnonymous,
	}

	biometricIdentity := &identity.Info{
		Type: model.IdentityTypeBiometric,
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
