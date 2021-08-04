package jwkutil

import (
	"github.com/lestrrat-go/jwx/jwk"
)

func PublicPEM(set jwk.Set) (bytes []byte, err error) {
	set, err = jwk.PublicSetOf(set)
	if err != nil {
		return
	}

	bytes, err = jwk.Pem(set)
	if err != nil {
		return
	}

	return
}

func PrivatePublicPEM(set jwk.Set) (bytes []byte, err error) {
	bytes, err = jwk.Pem(set)
	if err != nil {
		return
	}

	return
}
