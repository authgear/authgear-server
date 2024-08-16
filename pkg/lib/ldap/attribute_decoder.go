package ldap

import "github.com/google/uuid"

type AttributeDecoder interface {
	DecodeToStringRepresentable(bytes []byte) (string, error)
}

type StringAttributeDecoder struct{}

var _ AttributeDecoder = StringAttributeDecoder{}

func (StringAttributeDecoder) DecodeToStringRepresentable(bytes []byte) (string, error) {
	return string(bytes), nil
}

type UUIDAttributeDecoder struct{}

var _ AttributeDecoder = UUIDAttributeDecoder{}

func (UUIDAttributeDecoder) DecodeToStringRepresentable(bytes []byte) (string, error) {
	UUID, err := uuid.FromBytes(bytes)
	if err != nil {
		return "", err
	}
	return UUID.String(), nil
}
