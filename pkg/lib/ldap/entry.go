package ldap

import "github.com/go-ldap/ldap/v3"

const (
	AttributeNameEmail    = "mail"
	AttributeNameMobile   = "mobile"
	AttributeNameUsername = "uid"
)

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
