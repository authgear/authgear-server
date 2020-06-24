package jwkutil

import (
	"errors"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
)

func PublicKeySet(set *jwk.Set) (*jwk.Set, error) {
	jwks := &jwk.Set{}
	for _, key := range set.Keys {
		var ptrKey interface{}
		err := key.Raw(&ptrKey)
		if err != nil {
			return nil, err
		}

		pk, err := jwk.PublicKeyOf(ptrKey)
		if err != nil {
			return nil, err
		}

		key, err = jwk.New(pk)
		if err != nil {
			return nil, err
		}

		jwks.Keys = append(jwks.Keys, key)
	}
	return jwks, nil
}

func ExtractOctetKey(set *jwk.Set, id string) ([]byte, error) {
	for _, key := range set.Keys {
		if id != "" && key.KeyID() != id {
			continue
		}
		switch key.KeyType() {
		case jwa.OctetSeq:
			var bytes []byte
			err := key.Raw(&bytes)
			if err != nil {
				return nil, err
			}
			return bytes, nil
		default:
			return nil, errors.New("unexpected key type (key type should be octet)")
		}
	}

	return nil, errors.New("octet key not found")
}
