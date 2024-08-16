package ldap

type AttributeNameAndValue struct {
	Name  string
	Value string
}

func (a AttributeNameAndValue) String() string {
	return EncodeAttributeName(a.Name) + "=" + EncodeAttributeValue(a.Value)
}
