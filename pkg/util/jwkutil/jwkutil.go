package jwkutil

import (
	"context"
	"errors"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
)

const (
	KeyCreatedAt = "created_at"
)

func ExtractOctetKey(set jwk.Set, id string) ([]byte, error) {
	for it := set.Iterate(context.Background()); it.Next(context.Background()); {
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
