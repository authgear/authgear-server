package anonymous

import (
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
	return jwk.ParseKey(i.Key)
}
