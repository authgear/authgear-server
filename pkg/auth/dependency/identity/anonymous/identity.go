package anonymous

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/lestrrat-go/jwx/jwk"
)

var KeyIDFormat = regexp.MustCompile(`^[-\w]{8,64}$`)

type Identity struct {
	ID     string
	UserID string
	KeyID  string
	Key    []byte
}

func (i *Identity) toJWK() (jwk.Key, error) {
	key := &jwk.RSAPublicKey{}
	var jwkMap map[string]interface{}
	if err := json.Unmarshal(i.Key, &jwkMap); err != nil {
		return nil, fmt.Errorf("invalid JWK: %w", err)
	}
	if err := key.ExtractMap(jwkMap); err != nil {
		return nil, fmt.Errorf("invalid JWK: %w", err)
	}
	return key, nil
}
