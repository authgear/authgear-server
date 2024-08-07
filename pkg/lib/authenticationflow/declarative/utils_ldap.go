package declarative

import (
	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ldap"
)

func createIdentitySpecFromLDAPEntry(deps *authflow.Dependencies, serverName string, entry *ldap.Entry) (*identity.Spec, error) {
	var ldapServerConfig *config.LDAPServerConfig
	for _, serverConfig := range deps.Config.Identity.LDAP.Servers {
		if serverConfig.Name == serverName {
			ldapServerConfig = serverConfig
		}
	}
	if ldapServerConfig == nil {
		return nil, api.ErrLDAPServerNotFound
	}

	userIDAttributeName := ldapServerConfig.UserIDAttributeName
	userIDAttributeValue := entry.GetAttributeValue(userIDAttributeName)

	return &identity.Spec{
		Type: model.IdentityTypeLDAP,
		LDAP: &identity.LDAPSpec{
			ServerName:           serverName,
			UserIDAttributeName:  userIDAttributeName,
			UserIDAttributeValue: userIDAttributeValue,
			Claims:               extractLDAPEntryClaims(entry),
			RawEntryJSON:         entry.ToJSON(),
		},
	}, nil
}

func extractLDAPEntryClaims(entry *ldap.Entry) map[string]interface{} {
	claims := map[string]interface{}{}
	if v := entry.GetAttributeValue(ldap.AttributeNameEmail); v != "" {
		claims[string(model.ClaimEmail)] = v
	}
	if v := entry.GetAttributeValue(ldap.AttributeNameMobile); v != "" {
		claims[string(model.ClaimPhoneNumber)] = v
	}
	if v := entry.GetAttributeValue(ldap.AttributeNameUsername); v != "" {
		claims[string(model.ClaimPreferredUsername)] = v
	}
	return claims
}
