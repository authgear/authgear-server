package ldap

import "github.com/go-ldap/ldap/v3"

func init() {
	sensitiveAttributeList := []string{
		AttributeNamePassword,
	}
	sensitiveAttributes = make(map[string]struct{})
	for _, attr := range sensitiveAttributeList {
		sensitiveAttributes[attr] = struct{}{}
	}
}

// Here are some well-known attributes' names.
const (
	// Email is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.16
	AttributeNameEmail = "mail"
	// Mobile phone is from https://datatracker.ietf.org/doc/html/rfc4524#section-2.18
	AttributeNameMobile = "mobile"
	// Username is from https://datatracker.ietf.org/doc/html/rfc4519#section-2.39
	AttributeNameUsername = "uid"
	// Password is from From https://datatracker.ietf.org/doc/html/rfc4519#section-2.41
	AttributeNamePassword = "userPassword"
)

var sensitiveAttributes map[string]struct{}

type Entry struct {
	*ldap.Entry
}

func (e *Entry) ToJSON() map[string]interface{} {
	dict := map[string]interface{}{}
	dict["dn"] = e.DN
	for _, attr := range e.Attributes {
		dict[attr.Name] = attr.Values
	}
	return dict
}
