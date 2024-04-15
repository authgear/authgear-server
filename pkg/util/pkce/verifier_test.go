package pkce

import (
	"encoding/base64"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGenerateS256Verifier(t *testing.T) {
	Convey("GenerateS256Verifier", t, func() {
		v := GenerateS256Verifier()
		So(v.CodeVerifier, ShouldHaveLength, 43)

		b, err := base64.RawURLEncoding.DecodeString(v.CodeVerifier)
		So(err, ShouldBeNil)

		zeros := make([]byte, 32)
		So(b, ShouldNotResemble, zeros)
	})
}
