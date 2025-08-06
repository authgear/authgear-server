package config_test

import (
	"encoding/json"
	"testing"

	"github.com/lestrrat-go/jwx/v2/jwa"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestOAuthClientCredentialsItemMarshalUnmarshalJSON(t *testing.T) {
	Convey("OAuthClientCredentialsItemMarshalUnmarshalJSON", t, func() {
		// nolint: gosec
		secretJSON := `{"client_id":"confidential-client","keys":[{"created_at":1136171045,"k":"c2VjcmV0MQ","kid":"9dc0e72c-bf34-4ab3-a616-393093bdae0b","kty":"oct"}]}`

		// Test Unmarshal
		var item *config.OAuthClientCredentialsItem
		err := json.Unmarshal([]byte(secretJSON), &item)
		So(err, ShouldBeNil)
		So(item.ClientID, ShouldEqual, "confidential-client")
		k, ok := item.Key(0)
		So(ok, ShouldBeTrue)
		So(k.KeyID(), ShouldEqual, "9dc0e72c-bf34-4ab3-a616-393093bdae0b")
		So(k.KeyType(), ShouldEqual, jwa.OctetSeq)
		var key []byte
		err = k.Raw(&key)
		So(err, ShouldBeNil)
		So(string(key), ShouldEqual, "secret1")

		// Test Marshal
		actualJSON, err := json.Marshal(item)
		So(err, ShouldBeNil)
		So(string(actualJSON), ShouldEqualJSON, secretJSON)

		So(item.SensitiveStrings(), ShouldResemble, []string{"secret1"})
	})
}

func TestOAuthClientCredentialsOctetKey_Mask(t *testing.T) {
	Convey("OAuthClientCredentialsOctetKey.Mask", t, func() {
		testCases := []struct {
			name     string
			key      []byte
			expected string
		}{
			{
				name:     "empty string",
				key:      []byte(""),
				expected: "******",
			},
			{
				name:     "1 character string",
				key:      []byte("a"),
				expected: "******",
			},
			{
				name:     "2 character string",
				key:      []byte("ab"),
				expected: "a******",
			},
			{
				name:     "3 character string",
				key:      []byte("abc"),
				expected: "a******",
			},
			{
				name:     "4 character string",
				key:      []byte("abcd"),
				expected: "ab******",
			},
			{
				name:     "5 character string",
				key:      []byte("abcde"),
				expected: "ab******",
			},
			{
				name:     "6 character string",
				key:      []byte("abcdef"),
				expected: "abc******",
			},
			{
				name:     "7 character string",
				key:      []byte("abcdefg"),
				expected: "abc******",
			},
			{
				name:     "8 character string",
				key:      []byte("abcdefgh"),
				expected: "abcd******",
			},
			{
				name:     "long string",
				key:      []byte("thisisalongkeythatneedstobemasked"),
				expected: "this******",
			},
		}

		for _, tc := range testCases {
			Convey(tc.name, func() {
				k := config.OAuthClientCredentialsOctetKey{Key: tc.key}
				So(k.Mask(), ShouldEqual, tc.expected)
			})
		}
	})
}
