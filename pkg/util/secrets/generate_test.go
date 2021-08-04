package secrets

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/jwk"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/jwkutil"
	"github.com/authgear/authgear-server/pkg/util/rand"
)

func TestGenerateOctetKey(t *testing.T) {
	isInAlphabet := func(r rune) bool {
		return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r == '_')
	}

	Convey("GenerateOctetKey", t, func() {
		createdAt := time.Date(2006, 1, 2, 3, 4, 5, 0, time.UTC)
		jwkKey := GenerateOctetKey(createdAt, rand.InsecureRand)

		sKey, ok := jwkKey.(jwk.SymmetricKey)
		So(ok, ShouldBeTrue)

		octets := sKey.Octets()
		So(octets, ShouldHaveLength, 32)

		for _, r := range string(octets) {
			So(isInAlphabet(r), ShouldBeTrue)
		}

		// Able to get created_at just after fresh creation.
		createdAtIface, ok := jwkKey.Get(jwkutil.KeyCreatedAt)
		So(ok, ShouldBeTrue)
		var float64Type float64
		So(createdAtIface, ShouldHaveSameTypeAs, float64Type)
		So(createdAtIface.(float64), ShouldEqual, 1136171045)

		// Able to get created_at after marshaling and unmarshaling.
		jwkSet := jwk.NewSet()
		jwkSet.Add(jwkKey)
		jwkSetJSON, err := json.Marshal(jwkSet)
		So(err, ShouldBeNil)
		newSet := jwk.NewSet()
		err = json.Unmarshal(jwkSetJSON, &newSet)
		So(err, ShouldBeNil)
		So(newSet.Len(), ShouldEqual, 1)
		key, ok := newSet.Get(0)
		So(ok, ShouldBeTrue)
		createdAtIface, ok = key.Get(jwkutil.KeyCreatedAt)
		So(ok, ShouldBeTrue)
		So(createdAtIface, ShouldHaveSameTypeAs, float64Type)
		So(createdAtIface.(float64), ShouldEqual, 1136171045)
	})
}
