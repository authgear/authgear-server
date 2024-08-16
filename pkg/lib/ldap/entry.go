package ldap

import "github.com/go-ldap/ldap/v3"

func init() {
	sensitiveAttributeList := []string{
		AttributeUserPassword.Name,
	}
	sensitiveAttributes = make(map[string]struct{})
	for _, attr := range sensitiveAttributeList {
		sensitiveAttributes[attr] = struct{}{}
	}
}

var sensitiveAttributes map[string]struct{}

type Entry struct {
	*ldap.Entry
}

func (e *Entry) ToJSON() map[string]interface{} {
	dict := map[string]interface{}{}
	dict["dn"] = e.DN
	for _, attr := range e.Attributes {
		dict[attr.Name] = attr.ByteValues
	}
	return dict
}
