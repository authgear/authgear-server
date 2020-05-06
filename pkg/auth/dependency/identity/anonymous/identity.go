package anonymous

import (
	"errors"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/lestrrat-go/jwx/jwk"
)

type Identity struct {
	ID     string
	UserID string
	KeyID  string
	Key    string
}

func (i *Identity) ParseJWT(token string, claims jwt.Claims) error {
	keys, err := jwk.ParseString(i.Key)
	if err != nil {
		return err
	}

	keyFunc := func(token *jwt.Token) (interface{}, error) {
		keyID, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("no kid provided")
		}
		if key := keys.LookupKeyID(keyID); len(key) == 1 {
			return key[0].Materialize()
		}
		return nil, errors.New("no matching key")
	}
	_, err = jwt.ParseWithClaims(token, claims, keyFunc)
	if err != nil {
		return errors.New("invalid JWT signature")
	}

	return nil
}
