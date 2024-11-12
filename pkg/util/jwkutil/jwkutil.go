package jwkutil

import (
	"errors"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

const (
	KeyCreatedAt = "created_at"
)

func ExtractOctetKey(set jwk.Set, id string) ([]byte, error) {
	for it := set.Keys(contextForTheUnusedContextArgumentInJWXV2API); it.Next(contextForTheUnusedContextArgumentInJWXV2API); {
		pair := it.Pair()
		key := pair.Value.(jwk.Key)

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

func ExtractOctetKeys(set jwk.Set) ([][]byte, error) {
	keys := [][]byte{}
	for it := set.Keys(contextForTheUnusedContextArgumentInJWXV2API); it.Next(contextForTheUnusedContextArgumentInJWXV2API); {
		pair := it.Pair()
		key := pair.Value.(jwk.Key)
		switch key.KeyType() {
		case jwa.OctetSeq:
			var bytes []byte
			err := key.Raw(&bytes)
			if err != nil {
				return nil, err
			}
			keys = append(keys, bytes)
		default:
			return nil, errors.New("unexpected key type (key type should be octet)")
		}
	}
	return keys, nil
}
