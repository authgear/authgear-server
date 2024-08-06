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

const (
	AttributeNameEmail    = "mail"
	AttributeNameMobile   = "mobile"
	AttributeNameUsername = "uid"
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
