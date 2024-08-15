package identity

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
)

type LDAP struct {
	ID                   string                 `json:"id"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
	UserID               string                 `json:"user_id"`
	ServerName           string                 `json:"server_name"`
	UserIDAttributeOID   string                 `json:"user_id_attribute_oid"`
	UserIDAttributeValue string                 `json:"user_id_attribute_value"`
	Claims               map[string]interface{} `json:"claims,omitempty"`
	RawEntryJSON         map[string]interface{} `json:"raw_entry_json,omitempty"`
}

func (i *LDAP) ToInfo() *Info {
	return &Info{
		ID:        i.ID,
		UserID:    i.UserID,
		CreatedAt: i.CreatedAt,
		UpdatedAt: i.UpdatedAt,
		Type:      model.IdentityTypeLDAP,

		LDAP: i,
	}
}

func (i *LDAP) ToLDAPSpec() *LDAPSpec {
	return &LDAPSpec{
		ServerName:           i.ServerName,
		UserIDAttributeOID:   i.UserIDAttributeOID,
		UserIDAttributeValue: i.UserIDAttributeValue,
		Claims:               i.Claims,
		RawEntryJSON:         i.RawEntryJSON,
	}
}

// TODO(DEV-1668)
// We need to convert ldap entry attribute to identity aware standard claims
func (i *LDAP) IdentityAwareStandardClaims() map[model.ClaimName]string {
	return make(map[model.ClaimName]string)
}
