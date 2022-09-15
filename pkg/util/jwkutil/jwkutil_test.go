package jwkutil

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/lestrrat-go/jwx/jwk"
)

func TestPublicSetOf(t *testing.T) {
	Convey("PublicSetOf", t, func() {
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

		pkSet, err := jwk.PublicSetOf(set)
		So(err, ShouldBeNil)

		jj, err := json.Marshal(pkSet)
		So(err, ShouldBeNil)
		So(string(jj), ShouldEqualJSON, `
		{
			"keys": [
			{
				"e": "AQAB",
				"kty": "RSA",
				"n": "oDoW_ZqdK1BsZjLZ7hbtDKK6cp0cao9stOSIIdxxkWQsAwIG1VCpqSojC81EnbOAe6agqthozFCosJFjqO3ViQ",
				"alg": "RS256"
			}
			]
		}
		`)
	})
}

func TestExtractOctetKey(t *testing.T) {
	Convey("ExtractOctetKey", t, func() {
		key1, err := jwk.New([]byte("secret1"))
		if err != nil {
			panic(err)
		}
		_ = key1.Set("kid", "key-1")

		key2, err := jwk.New([]byte("secret2"))
		if err != nil {
			panic(err)
		}
		_ = key2.Set("kid", "key-2")

		set := jwk.NewSet()
		_ = set.Add(key1)
		_ = set.Add(key2)

		Convey("should match on key ID", func() {
			octetKey, err := ExtractOctetKey(set, "key-1")
			So(err, ShouldBeNil)
			So(octetKey, ShouldResemble, []byte("secret1"))

			octetKey, err = ExtractOctetKey(set, "key-2")
			So(err, ShouldBeNil)
			So(octetKey, ShouldResemble, []byte("secret2"))

			_, err = ExtractOctetKey(set, "key-3")
			So(err, ShouldBeError)
		})
		Convey("should match first key if key ID is not provided", func() {
			octetKey, err := ExtractOctetKey(set, "")
			So(err, ShouldBeNil)
			So(octetKey, ShouldResemble, []byte("secret1"))
		})
	})
}

func TestExtractOctetKeys(t *testing.T) {
	Convey("ExtractOctetKeys", t, func() {
		key1, err := jwk.New([]byte("secret1"))
		if err != nil {
			panic(err)
		}
		_ = key1.Set("kid", "key-1")

		key2, err := jwk.New([]byte("secret2"))
		if err != nil {
			panic(err)
		}
		_ = key2.Set("kid", "key-2")

		set := jwk.NewSet()
		_ = set.Add(key1)
		_ = set.Add(key2)

		Convey("should extract keys", func() {
			keys, err := ExtractOctetKeys(set)
			So(err, ShouldBeNil)
			So(keys, ShouldResemble, [][]byte{[]byte("secret1"), []byte("secret2")})
		})
	})
}
