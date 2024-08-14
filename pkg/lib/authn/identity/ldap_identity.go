package identity

import (
	"encoding/base64"
	"time"
	"unicode/utf8"

	goldap "github.com/go-ldap/ldap/v3"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/ldap"
)

type LDAP struct {
	ID                   string                 `json:"id"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
	UserID               string                 `json:"user_id"`
	ServerName           string                 `json:"server_name"`
	UserIDAttributeName  string                 `json:"user_id_attribute_name"`
	UserIDAttributeValue []byte                 `json:"user_id_attribute_value"`
	Claims               map[string]interface{} `json:"claims,omitempty"`
	RawEntryJSON         map[string]interface{} `json:"raw_entry_json,omitempty"`
}

func (i *LDAP) UserIDAttributeValueDisplayValue() string {
	ldapAttribute, ok := ldap.DefaultAttributeRegistry.Get(i.UserIDAttributeName)
	// We try to decode with known attribute first
	if ok {
		stringValues, err := ldapAttribute.Type.Decoder().DecodeToStringRepresentable([][]byte{i.UserIDAttributeValue})
		if err == nil {
			return stringValues[0]
		}
	}
	// If the attribute is unknown, we return its in string
	// format if it is a valid utf8 bytes
	if utf8.Valid(i.UserIDAttributeValue) {
		return string(i.UserIDAttributeValue)
	}
	// Otherise, we encode the bytes in base64
	str := base64.StdEncoding.EncodeToString(i.UserIDAttributeValue)
	return str
}

// EntryJSON returns a map that with attributes known by us
func (i *LDAP) EntryJSON() map[string]interface{} {
	result := map[string]interface{}{}
	if dn, ok := i.RawEntryJSON["dn"].(string); ok {
		result["dn"] = dn
	}
	for name, values := range i.RawEntryJSON {
		ldapAttribute, ok := ldap.DefaultAttributeRegistry.Get(name)
		if !ok {
			continue
		}
		var byteValues [][]byte
		for _, byteStr := range values.([]interface{}) {
			bytes, err := base64.StdEncoding.DecodeString(byteStr.(string))
			if err == nil {
				byteValues = append(byteValues, bytes)
			}
		}
		stringValues, err := ldapAttribute.Type.Decoder().DecodeToStringRepresentable(byteValues)
		if err != nil {
			continue
		}
		result[name] = stringValues
	}
	return result
}

func (i *LDAP) DisplayID() string {
	dn, ok := i.RawEntryJSON["dn"].(string)
	if !ok {
		ldapAttribute, ok := ldap.DefaultAttributeRegistry.Get(i.UserIDAttributeName)
		if !ok {
			return (&goldap.AttributeTypeAndValue{
				Type:  i.UserIDAttributeName,
				Value: string(i.UserIDAttributeValue),
			}).String()
		}
		stringValues, err := ldapAttribute.Type.Decoder().DecodeToStringRepresentable([][]byte{i.UserIDAttributeValue})
		if err != nil {
			return (&goldap.AttributeTypeAndValue{
				Type:  i.UserIDAttributeName,
				Value: string(i.UserIDAttributeValue),
			}).String()
		}
		return (&goldap.AttributeTypeAndValue{
			Type:  i.UserIDAttributeName,
			Value: stringValues[0],
		}).String()
	}
	return dn
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
		UserIDAttributeName:  i.UserIDAttributeName,
		UserIDAttributeValue: i.UserIDAttributeValue,
		Claims:               i.Claims,
		RawEntryJSON:         i.RawEntryJSON,
	}
}

// TODO(DEV-1668): Support attributes mapping in LDAP
// We need to convert ldap entry attribute to identity aware standard claims
// Expected to return ClaimEmail or ClaimPhoneNumber or ClaimPreferredUsername
func (i *LDAP) IdentityAwareStandardClaims() map[model.ClaimName]string {
	claims := map[model.ClaimName]string{}
	if email, ok := i.Claims[string(model.ClaimEmail)].(string); ok {
		claims[model.ClaimEmail] = email
	}
	if phoneNumber, ok := i.Claims[string(model.ClaimPhoneNumber)].(string); ok {
		claims[model.ClaimPhoneNumber] = phoneNumber
	}
	if username, ok := i.Claims[string(model.ClaimPreferredUsername)].(string); ok {
		claims[model.ClaimPreferredUsername] = username
	}
	return claims
}
