package webapp

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

// identitiesDisplayNamePriorities indicates which identity type will be shown
// as the identitiesDisplayName
// Higher level means more willing to be shown
var identitiesDisplayNamePriorities = map[authn.IdentityType]int{
	authn.IdentityTypeLoginID: 2,
	authn.IdentityTypeOAuth:   1,
}

func IdentitiesDisplayName(identities []*identity.Info) string {
	level := 0
	var i *identity.Info
	for _, perIdentity := range identities {
		l := identitiesDisplayNamePriorities[perIdentity.Type]
		if l >= level {
			level = l
			i = perIdentity
		}
	}

	if i == nil {
		return ""
	}

	switch i.Type {
	case authn.IdentityTypeLoginID:
		return i.DisplayID()
	case authn.IdentityTypeOAuth:
		providerType, _ := i.Claims[identity.IdentityClaimOAuthProviderType].(string)
		displayID := i.DisplayID()
		if displayID != "" {
			return fmt.Sprintf("%s:%s", providerType, i.DisplayID())
		}
		return providerType
	case authn.IdentityTypeAnonymous:
		return "anonymous"
	case authn.IdentityTypeBiometric:
		return "biometric"
	default:
		return ""
	}
}
