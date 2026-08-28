package dcr

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/crypto"
)

func TestGenerateInitialAccessToken(t *testing.T) {
	Convey("GenerateInitialAccessToken", t, func() {
		Convey("third-party token has iat_tp_ prefix and matching hash", func() {
			plaintext, hash := GenerateInitialAccessToken(InitialAccessTokenTypeThirdParty)
			So(strings.HasPrefix(plaintext, IATPrefixThirdParty), ShouldBeTrue)
			So(len(strings.TrimPrefix(plaintext, IATPrefixThirdParty)), ShouldEqual, tokenRandomLength)
			So(hash, ShouldEqual, crypto.SHA256String(plaintext))
		})

		Convey("first-party token has iat_fp_ prefix and matching hash", func() {
			plaintext, hash := GenerateInitialAccessToken(InitialAccessTokenTypeFirstParty)
			So(strings.HasPrefix(plaintext, IATPrefixFirstParty), ShouldBeTrue)
			So(len(strings.TrimPrefix(plaintext, IATPrefixFirstParty)), ShouldEqual, tokenRandomLength)
			So(hash, ShouldEqual, crypto.SHA256String(plaintext))
		})

		Convey("two generated tokens are not equal", func() {
			plaintext1, _ := GenerateInitialAccessToken(InitialAccessTokenTypeThirdParty)
			plaintext2, _ := GenerateInitialAccessToken(InitialAccessTokenTypeThirdParty)
			So(plaintext1, ShouldNotEqual, plaintext2)
		})
	})
}

func TestHashInitialAccessToken(t *testing.T) {
	Convey("HashInitialAccessToken is deterministic and matches crypto.SHA256String", t, func() {
		plaintext := "iat_tp_Xf2kLmNpQrStUvWx"
		So(HashInitialAccessToken(plaintext), ShouldEqual, crypto.SHA256String(plaintext))
		So(HashInitialAccessToken(plaintext), ShouldEqual, HashInitialAccessToken(plaintext))
	})
}
