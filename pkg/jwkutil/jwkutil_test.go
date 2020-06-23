package jwkutil

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/lestrrat-go/jwx/jwk"

	. "github.com/skygeario/skygear-server/pkg/core/skytest"
)

func TestPublicKeySet(t *testing.T) {
	Convey("PublicKeySet", t, func() {
		j := `
		{
			"keys": [
			{
				"p": "1HVBnIsDJdCfU7MjjlwDST6SeOkxsQMZf5zX3yatr5M",
				"kty": "RSA",
				"q": "wRB-wbjCYJR44QNxUTDfcbwDVht3rPRGF8KCrMeZP_M",
				"d": "Ead04wqHKL-awgP2WTJzegHwEzLWb7NP4fqQOVLlvIQlWrXCUg8HmrdzEDa_hAPMzzTANNRYGDBbnRwy1uEO1Q",
				"e": "AQAB",
				"qi": "Dwqq0Uyscdkr6kvPA2BhG3rFaUXNu5hhyDgrCGooX8o",
				"dp": "yOT_X5kLJuy4W5remjRjXxTtx6sps6msqMCUV4vpXEU",
				"alg": "RS256",
				"dq": "g9rXJ0ke_8UHFW47ex7szAmDIdDamDWwlVOT2ZrsMD8",
				"n": "oDoW_ZqdK1BsZjLZ7hbtDKK6cp0cao9stOSIIdxxkWQsAwIG1VCpqSojC81EnbOAe6agqthozFCosJFjqO3ViQ"
			}
			]
		}
		`

		set, err := jwk.ParseString(j)
		So(err, ShouldBeNil)

		pkSet, err := PublicKeySet(set)
		So(err, ShouldBeNil)

		jj, err := json.Marshal(pkSet)
		So(err, ShouldBeNil)
		So(jj, ShouldEqualJSON, `
		{
			"keys": [
			{
				"e": "AQAB",
				"kty": "RSA",
				"n": "oDoW_ZqdK1BsZjLZ7hbtDKK6cp0cao9stOSIIdxxkWQsAwIG1VCpqSojC81EnbOAe6agqthozFCosJFjqO3ViQ"
			}
			]
		}
		`)
	})
}

func TestExtractOctetKey(t *testing.T) {
	Convey("ExtractOctetKey", t, func() {
		secret := []byte("secret")
		key, err := jwk.New(secret)
		So(err, ShouldBeNil)
		set := &jwk.Set{
			Keys: []jwk.Key{key},
		}
		octetKey, err := ExtractOctetKey(set, "")
		So(err, ShouldBeNil)
		So(octetKey, ShouldResemble, secret)
	})
}
