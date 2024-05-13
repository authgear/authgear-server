package model_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"

	"github.com/ethereum/go-ethereum/crypto"
)

func TestSIWEPubKey(t *testing.T) {
	Convey("SIWEPublicKey", t, func() {
		Convey("should encode and decode key", func() {

			Convey("valid curve", func() {
				key, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
				So(err, ShouldBeNil)
				So(key, ShouldNotBeNil)

				hex, err := model.NewSIWEPublicKey(&key.PublicKey)
				So(err, ShouldBeNil)
				So(hex, ShouldNotBeEmpty)

				decodedKey, err := hex.ECDSA()
				So(err, ShouldBeNil)

				So(key.PublicKey, ShouldResemble, *decodedKey)
			})

			Convey("invalid curve", func() {
				key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
				So(err, ShouldBeNil)
				So(key, ShouldNotBeNil)

				hex, err := model.NewSIWEPublicKey(&key.PublicKey)
				So(err, ShouldBeError)
				So(hex, ShouldBeEmpty)
			})
		})

	})
}
