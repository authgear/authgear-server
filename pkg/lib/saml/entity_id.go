package saml

import "encoding/base64"

func EncodeEntityIDURLComponent(id string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(id))
}

func DecodeEntityIDURLComponent(encodedId string) (string, error) {
	idBytes, err := base64.RawURLEncoding.DecodeString(encodedId)
	if err != nil {
		return "", err
	}
	return string(idBytes), nil
}
