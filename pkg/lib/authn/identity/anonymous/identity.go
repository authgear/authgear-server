package anonymous

import (
	"regexp"
	"time"

	"github.com/lestrrat-go/jwx/jwk"
)

var KeyIDFormat = regexp.MustCompile(`^[-\w]{8,64}$`)

type Identity struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	UserID    string
	KeyID     string
	Key       []byte
}

func (i *Identity) toJWK() (jwk.Key, error) {
	return jwk.ParseKey(i.Key)
}
