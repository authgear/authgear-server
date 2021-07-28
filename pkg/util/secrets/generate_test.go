package secrets

import (
	"testing"

	"github.com/lestrrat-go/jwx/jwk"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/rand"
)

func TestGenerateOctetKey(t *testing.T) {
	isInAlphabet := func(r rune) bool {
		return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r == '_')
	}

	Convey("GenerateOctetKey", t, func() {
		jwkKey := GenerateOctetKey(rand.InsecureRand)

		sKey, ok := jwkKey.(jwk.SymmetricKey)
		So(ok, ShouldBeTrue)

		octets := sKey.Octets()
		So(octets, ShouldHaveLength, 32)

		for _, r := range string(octets) {
			So(isInAlphabet(r), ShouldBeTrue)
		}
	})
}
