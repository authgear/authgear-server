package declarative

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ldap"
)

func createIdentitySpecFromLDAPEntry(deps *authflow.Dependencies, serverConfig *config.LDAPServerConfig, entry *ldap.Entry) (*identity.Spec, error) {
	userIDAttributeName := serverConfig.UserIDAttributeName
	userIDAttributeValue := entry.GetRawAttributeValue(userIDAttributeName)

	return &identity.Spec{
		Type: model.IdentityTypeLDAP,
		LDAP: &identity.LDAPSpec{
			ServerName:           serverConfig.Name,
			UserIDAttributeName:  userIDAttributeName,
			UserIDAttributeValue: userIDAttributeValue,
			Claims:               extractLDAPEntryClaims(entry),
			RawEntryJSON:         entry.ToJSON(),
		},
	}, nil
}

func extractLDAPEntryClaims(entry *ldap.Entry) map[string]interface{} {
	claims := map[string]interface{}{}
	if v := entry.GetAttributeValue(ldap.AttributeMail.Name); v != "" {
		claims[string(model.ClaimEmail)] = v
	}
	if v := entry.GetAttributeValue(ldap.AttributeMobile.Name); v != "" {
		claims[string(model.ClaimPhoneNumber)] = v
	}
	if v := entry.GetAttributeValue(ldap.AttributeUID.Name); v != "" {
		claims[string(model.ClaimPreferredUsername)] = v
	}
	return claims
}
