package jwkutil

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/lestrrat-go/jwx/jwk"

	. "github.com/smartystreets/goconvey/convey"
)

const PrivateKeyPEM = `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC89eQDeH8icj6j
1DUHTXKyhFkYOVrOVLA4xflDwqAuw5IrJQNgIjTsBZXrR1rh4BSBsjoE0ToH+/Da
MfyAicQpv7QPI4pM8a/a3SY+rlr4j4LzFtchUvBMcGbSZZqKINBtxpAsFLPGFnwF
NrxXIwrxE79cgY+g1KcmF8twqDmmash6fMoOeU8MTa8Q9Z7wTzhySeeZlBVFtvJp
79Wqe75dtp0pe6E6ujavVjPifj2Msdl9RW7KhJsttgGhMGR2Jp07nAIBT150qX0G
3gu0G5ILbgxcrhYZYK5fk/u6MQ0sAyXwS+fmppsPmYw6UVYlS2UGnaJlCE7Ml0e2
yyEyrbmnAgMBAAECggEARX7NsDUV1O5deVVnd1sVjvA78DvP2Miu0wKErVYcIXbO
AE4pkqah/hgDzjc9BouqHxUUX4cvp5YSO71cl02TtqMJrvOsPqY4ve7NzQnE7Vui
lpLU5i2hsQs51bGGh7yPy3/WsE+g2n6UeDpsREPgF0/i9ju0PjtXihwAN1u3cCt9
t9CsSGliHqQX9uO7o92yN+aROKEbw3x3gKpRJ/Gv3fQcVR01cXvaBrtdEb6kEVEB
WBlCA0kmRc/H7jVYGcWqalLDjj99Pox47PLUigyJsNxJmMD881Ihah4zEQMpX7pW
eRuyISTAA+i0MXO8+bypE6trglF8YQH6JTcVLTz70QKBgQDngFYD0gAqB41vMpmQ
TGSr16qs63Q9QD0Ot9ZkSvYY745HvK7syLq6FLZl5Qz/f45NQ99BQtkGDZjE8sn9
W4V7/yA8xzNP+xmvdqsoOAcIO4j8W34dA6gS+z4h2u98LpqV9Q6ehbrZCPB8/MSn
1QTnbINGw1ZCxfj6olN7ppaZaQKBgQDQ9RIHxHVhDHWpa83Fgf0oDve2UXD5YDsZ
Axu6cYQOGCM7h0WxwDViIUuieWortYvGq8K1IlqfaDWlo5BHRXozmakMpJ4K8sBW
F8TWn7PYw9cPH2XuZZHPnPiYkkhe0SoAifa3tk4bgyR5txOjdCr5L3ZFWfM7Vmkp
hL2M7JTIjwKBgDXyShkJzs/8gpDvEan2o18IGtXA6I19cr0DSgqFDWQyLs24wmqb
PCgwu3BzN9wyNU78CgKDOV+Xu4npqfhIY4rJoRGIugRhV1L0LF5q7/iTJxDnoTPR
rlD+CzSIeFZP5eYb/RQjxa7dzmzR2mHh2gqz1sOesXNN/v8o5Jtj7qRBAoGAEibH
yy7wt156th3sQRT6pckvEYJfmvoWCCUx+m8z9nl4TgqBLmCxAnY7+MAtTeC2ZKq0
/kEeuCw4RMxBkz9gzyyw960xIWhW9uOXsMEswU651tF2bFAca3mKSs6iRMJMsMFL
Ukge3tr0hzI1HYTQ2taZooqey2/FMNscECrY/dcCgYEAuhBfEof+DCeuLmKgvks+
Idv5Ky2ZIR49L8VxCy7K+BXhr2vnKX6itlVDQVVpNIphdLHXQK6CNr8Ko5WinZHu
gouLseU4p4zh8vYZcgPyqlLEdkygMCN0b0+HVaBTs0jlLGbvTC0Oiz69umYMe+5g
eZDnqWNf7mYPdP5mO5iTtMw=
-----END PRIVATE KEY-----
`

const PublicKeyPEM = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvPXkA3h/InI+o9Q1B01y
soRZGDlazlSwOMX5Q8KgLsOSKyUDYCI07AWV60da4eAUgbI6BNE6B/vw2jH8gInE
Kb+0DyOKTPGv2t0mPq5a+I+C8xbXIVLwTHBm0mWaiiDQbcaQLBSzxhZ8BTa8VyMK
8RO/XIGPoNSnJhfLcKg5pmrIenzKDnlPDE2vEPWe8E84cknnmZQVRbbyae/Vqnu+
XbadKXuhOro2r1Yz4n49jLHZfUVuyoSbLbYBoTBkdiadO5wCAU9edKl9Bt4LtBuS
C24MXK4WGWCuX5P7ujENLAMl8Evn5qabD5mMOlFWJUtlBp2iZQhOzJdHtsshMq25
pwIDAQAB
-----END PUBLIC KEY-----
`

func TestPublicPEM(t *testing.T) {
	Convey("TestPublicPEM", t, func() {
		jwkSet, err := jwk.Parse([]byte(PrivateKeyPEM), jwk.WithPEM(true))
		So(err, ShouldBeNil)

		Convey("PublicPEM outputs PUBLIC KEY PEM block", func() {
			pemBytes, err := PublicPEM(jwkSet)
			So(err, ShouldBeNil)
			So(string(pemBytes), ShouldEqual, PublicKeyPEM)
		})

		Convey("The output of PublicPEM can be parsed with stdlib", func() {
			block, _ := pem.Decode([]byte(PublicKeyPEM))
			rsaPublicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
			So(err, ShouldBeNil)
			var rsaPublicKeyType *rsa.PublicKey
			So(rsaPublicKey, ShouldHaveSameTypeAs, rsaPublicKeyType)
		})
	})
}

func TestPrivatePublicPEM(t *testing.T) {
	Convey("TestPrivatePublicPEM", t, func() {
		jwkSet, err := jwk.Parse([]byte(PrivateKeyPEM), jwk.WithPEM(true))
		So(err, ShouldBeNil)

		Convey("PrivatePublicPEM outputs PRIVATE KEY PEM block", func() {
			bytes, err := PrivatePublicPEM(jwkSet)
			So(err, ShouldBeNil)
			So(string(bytes), ShouldEqual, PrivateKeyPEM)
		})

		Convey("The output of PrivatePublicPEM can be parsed with stdlib", func() {
			block, _ := pem.Decode([]byte(PrivateKeyPEM))
			rsaPrivateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			So(err, ShouldBeNil)
			var rsaPrivateKeyType *rsa.PrivateKey
			So(rsaPrivateKey, ShouldHaveSameTypeAs, rsaPrivateKeyType)
		})
	})
}
