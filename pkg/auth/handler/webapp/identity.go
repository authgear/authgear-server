package webapp

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

// identitiesDisplayNamePriorities indicates which identity type will be shown
// as the identitiesDisplayName
// Higher level means more willing to be shown
var identitiesDisplayNamePriorities = map[model.IdentityType]int{
	model.IdentityTypeLoginID: 2,
	model.IdentityTypeOAuth:   1,
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
	case model.IdentityTypeLoginID:
		return i.DisplayID()
	case model.IdentityTypeOAuth:
		providerType := i.OAuth.ProviderID.Type
		displayID := i.DisplayID()
		if displayID != "" {
			return fmt.Sprintf("%s:%s", providerType, i.DisplayID())
		}
		return providerType
	case model.IdentityTypeAnonymous:
		return "anonymous"
	case model.IdentityTypeBiometric:
		return "biometric"
	case model.IdentityTypePasskey:
		return "passkey"
	case model.IdentityTypeSIWE:
		return "siwe"
	case model.IdentityTypeLDAP:
		return "ldap"
	default:
		return ""
	}
}
