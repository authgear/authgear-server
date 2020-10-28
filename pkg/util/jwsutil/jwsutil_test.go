package jwsutil

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/jwkutil"
	"github.com/authgear/authgear-server/pkg/util/jwtutil"
)

func TestVerifyWithSet(t *testing.T) {
	Convey("Verify: no alg in key", t, func() {
		payload := jwt.New()
		_ = payload.Set("foobar", 42)

		// nolint: gosec
		privKey, err := rsa.GenerateKey(rand.Reader, 512)
		So(err, ShouldBeNil)

		jwkKey, err := jwk.New(privKey)
		So(err, ShouldBeNil)
		_ = jwkKey.Set("kid", "mykey")

		alg := jwa.RS256

		token, err := jwtutil.Sign(payload, alg, jwkKey)
		So(err, ShouldBeNil)

		privateKeySet := &jwk.Set{
			Keys: []jwk.Key{jwkKey},
		}

		publicKeySet, err := jwkutil.PublicKeySet(privateKeySet)
		So(err, ShouldBeNil)

		_, _, err = VerifyWithSet(publicKeySet, token)
		So(err, ShouldBeNil)
	})
}
