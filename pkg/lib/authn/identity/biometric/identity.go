package biometric

import (
	"regexp"
	"time"

	"github.com/lestrrat-go/jwx/jwk"
)

var KeyIDFormat = regexp.MustCompile(`^[-\w]{8,64}$`)

type Identity struct {
	ID         string
	Labels     map[string]interface{}
	CreatedAt  time.Time
	UpdatedAt  time.Time
	UserID     string
	KeyID      string
	Key        []byte
	DeviceInfo map[string]interface{}
}

func (i *Identity) toJWK() (jwk.Key, error) {
	return jwk.ParseKey(i.Key)
}
