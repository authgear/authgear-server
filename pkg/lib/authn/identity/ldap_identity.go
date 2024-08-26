package identity

import (
	"encoding/base64"
	"fmt"
	"time"
	"unicode"
	"unicode/utf8"

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
	LastLoginUserName    string                 `json:"last_login_username"`
}

func (i *LDAP) UserIDAttributeValueDisplayValue() string {
	return RenderAttribute(i.UserIDAttributeName, i.UserIDAttributeValue)
}

// EntryJSON returns a map that with attributes rendered.
func (i *LDAP) EntryJSON() map[string]interface{} {
	result := map[string]interface{}{}
	for name, values := range i.RawEntryJSON {
		switch name {
		case "dn":
			if dn, ok := values.(string); ok {
				result["dn"] = dn
			}
		default:
			var stringValues []string
			for _, byteStr := range values.([]interface{}) {
				bytes, err := base64.StdEncoding.DecodeString(byteStr.(string))
				if err != nil {
					panic(fmt.Errorf("ldap: unexpected malformed base64 encoded string: %w", err))
				}
				str := RenderAttribute(name, bytes)
				stringValues = append(stringValues, str)
			}
			result[name] = stringValues
		}
	}
	return result
}

func (i *LDAP) DisplayID() string {
	dn, ok := i.RawEntryJSON["dn"].(string)
	if !ok {
		return (ldap.AttributeNameAndValue{
			Name:  i.UserIDAttributeName,
			Value: i.UserIDAttributeValueDisplayValue(),
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
		LastLoginUserName:    i.LastLoginUserName,
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

func RenderAttribute(attributeName string, attributeValue []byte) string {
	attribute, ok := ldap.DefaultAttributeRegistry.Get(attributeName)
	if ok {
		// We try to decode with known attribute first
		str, err := attribute.Type.Decoder().DecodeToStringRepresentable(attributeValue)
		if err == nil {
			return str
		}
	}

	// If the attribute is unknown or decode failed, we return its in string
	// format if it looks like a printable string.
	if printableString, ok := ToPrintable(attributeValue); ok {
		return printableString
	}

	// Otherise, we encode the bytes in base64
	return base64.StdEncoding.EncodeToString(attributeValue)
}

func ToPrintable(b []byte) (str string, ok bool) {
	validUTF8 := utf8.Valid(b)
	if !validUTF8 {
		return
	}

	str = string(b)
	for _, r := range str {
		isGraphic := unicode.IsGraphic(r)
		isSpace := unicode.IsSpace(r)
		isPrintable := isGraphic || isSpace
		if !isPrintable {
			return
		}
	}

	ok = true
	return
}
