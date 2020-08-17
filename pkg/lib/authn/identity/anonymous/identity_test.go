package anonymous

import (
	"crypto/rsa"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestToJWK(t *testing.T) {
	Convey("IdentityToJWK", t, func() {
		iden := Identity{
			Key: []byte(`
			{
				"kty": "RSA",
				"e": "AQAB",
				"kid": "foobar",
				"alg": "RS256",
				"n": "ump6Mewd7SRkmp1IaxVFUj3kaWYVKrgSdxC3awpHutlALFcbQogObB5xmGNJ7wb5yLbhO9opnVprFyANbGCArw"
			}
			`),
		}
		key, err := iden.toJWK()
		So(err, ShouldBeNil)
		So(key.KeyID(), ShouldEqual, "foobar")
	})
}

func TestRaw(t *testing.T) {
	Convey("IdentityToJWK", t, func() {
		iden := Identity{
			Key: []byte(`
			{
				"kty": "RSA",
				"e": "AQAB",
				"kid": "foobar",
				"alg": "RS256",
				"n": "ump6Mewd7SRkmp1IaxVFUj3kaWYVKrgSdxC3awpHutlALFcbQogObB5xmGNJ7wb5yLbhO9opnVprFyANbGCArw"
			}
			`),
		}
		key, err := iden.toJWK()
		So(err, ShouldBeNil)

		var ptrKey interface{}
		err = key.Raw(&ptrKey)
		So(err, ShouldBeNil)
		So(ptrKey, ShouldNotBeNil)
		So(ptrKey, ShouldHaveSameTypeAs, &rsa.PublicKey{})
	})
}
