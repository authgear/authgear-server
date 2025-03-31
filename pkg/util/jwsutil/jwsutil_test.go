package jwsutil

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/jwtutil"
)

func TestVerifyWithSet(t *testing.T) {
	Convey("Verify: no alg in key", t, func() {
		payload := jwt.New()
		_ = payload.Set("foobar", 42)

		privKey, err := rsa.GenerateKey(rand.Reader, 2048)
		So(err, ShouldBeNil)

		jwkKey, err := jwk.FromRaw(privKey)
		So(err, ShouldBeNil)
		_ = jwkKey.Set("kid", "c240e6bd-0083-4ee7-98c0-518d65bb0dd8")

		alg := jwa.RS256

		token, err := jwtutil.Sign(payload, alg, jwkKey)
		So(err, ShouldBeNil)

		privateKeySet := jwk.NewSet()
		_ = privateKeySet.AddKey(jwkKey)

		publicKeySet, err := jwk.PublicSetOf(privateKeySet)
		So(err, ShouldBeNil)

		_, _, err = VerifyWithSet(publicKeySet, token)
		So(err, ShouldBeNil)
	})
}
