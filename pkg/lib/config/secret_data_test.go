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
