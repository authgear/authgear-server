package pkce

import (
	"encoding/base64"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewS256Verifier(t *testing.T) {
	Convey("NewS256Verifier", t, func() {
		_, err := NewS256Verifier("")
		So(err, ShouldBeError, "code_verifier must be a string between 43 and 128 characters long using A-Z, a-z, 0-9, -, ., _, ~")

		_, err = NewS256Verifier("0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789")
		So(err, ShouldBeError, "code_verifier must be a string between 43 and 128 characters long using A-Z, a-z, 0-9, -, ., _, ~")

		_, err = NewS256Verifier("012345678901234567890123456789012345678912$")
		So(err, ShouldBeError, "code_verifier must be a string between 43 and 128 characters long using A-Z, a-z, 0-9, -, ., _, ~")

		_, err = NewS256Verifier("0123456789012345678901234567890123456789123")
		So(err, ShouldBeNil)
	})
}

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

func TestVerify(t *testing.T) {
	Convey("Verify", t, func() {
		v, err := NewS256Verifier("jwEcs37HXgmZafciKLgzfsHtgJ3b714dhEop8Lzq_n4")
		So(err, ShouldBeNil)

		So(v.Verify(""), ShouldBeFalse)
		So(v.Verify("1"), ShouldBeFalse)
		So(v.Verify("JO1aklO5DCWCdDvEvCWJIKSuZmfQ25kyrNIEeSiahz4"), ShouldBeTrue)
	})
}
