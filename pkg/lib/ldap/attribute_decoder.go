package ldap

import "github.com/google/uuid"

type AttributeDecoder interface {
	DecodeToStringRepresentable(byteValues [][]byte) ([]string, error)
}

type StringAttributeDecoder struct{}

var _ AttributeDecoder = StringAttributeDecoder{}

func (StringAttributeDecoder) DecodeToStringRepresentable(byteValues [][]byte) ([]string, error) {
	result := make([]string, 0, len(byteValues))
	for _, bytes := range byteValues {
		result = append(result, string(bytes))
	}
	return result, nil
}

type UUIDAttributeDecoder struct{}

var _ AttributeDecoder = UUIDAttributeDecoder{}

func (UUIDAttributeDecoder) DecodeToStringRepresentable(byteValues [][]byte) ([]string, error) {
	result := make([]string, 0, len(byteValues))
	for _, bytes := range byteValues {
		UUID, err := uuid.FromBytes(bytes)
		if err != nil {
			return nil, err
		}
		result = append(result, UUID.String())
	}
	return result, nil
}
