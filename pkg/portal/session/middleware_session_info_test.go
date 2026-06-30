package session

import (
	"net/http"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/jwtutil"
)

const testPrivateKeyPEM = `-----BEGIN PRIVATE KEY-----
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

// buildTestJWKSet parses testPrivateKeyPEM into a JWK set usable for both
// signing (private key) and verification (public key).
func buildTestJWKSet(t *testing.T) (privateJWKSet jwk.Set, publicJWKSet jwk.Set) {
	t.Helper()
	set, err := jwk.Parse([]byte(testPrivateKeyPEM), jwk.WithPEM(true))
	if err != nil {
		t.Fatal(err)
	}
	privKey, _ := set.Key(0)
	_ = privKey.Set(jwk.KeyIDKey, "test-key-1")
	_ = privKey.Set(jwk.AlgorithmKey, jwa.RS256.String())

	pubSet, err := jwk.PublicSetOf(set)
	if err != nil {
		t.Fatal(err)
	}
	return set, pubSet
}

// signToken builds and signs a JWT with the given typ header and claims.
func signToken(t *testing.T, privJWKSet jwk.Set, typ string, sub string, issuedAt time.Time, expiry time.Time) string {
	t.Helper()

	claims := jwt.New()
	_ = claims.Set(jwt.SubjectKey, sub)
	_ = claims.Set(jwt.IssuedAtKey, issuedAt.Unix())
	_ = claims.Set(jwt.ExpirationKey, expiry.Unix())
	_ = claims.Set(string(model.ClaimUserIsAnonymous), false)
	_ = claims.Set(string(model.ClaimUserIsVerified), true)
	_ = claims.Set(string(model.ClaimUserCanReauthenticate), true)
	_ = claims.Set(string(model.ClaimAuthgearRoles), []any{})

	privKey, _ := privJWKSet.Key(0)

	hdr := jws.NewHeaders()
	_ = hdr.Set("typ", typ)

	signed, err := jwtutil.SignWithHeader(claims, hdr, jwa.RS256, privKey)
	if err != nil {
		t.Fatal(err)
	}
	return string(signed)
}

func makeHeader(token string) http.Header {
	h := http.Header{}
	h.Set("Authorization", "Bearer "+token)
	return h
}

func TestJWTToSessionInfo(t *testing.T) {
	Convey("jwtToSessionInfo", t, func() {
		privJWKSet, pubJWKSet := buildTestJWKSet(t)

		now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		m := &SessionInfoMiddleware{
			Clock: clock.NewMockClockAtTime(now),
		}

		Convey("valid at+jwt access token returns valid session", func() {
			token := signToken(t, privJWKSet, "at+jwt", "user123", now.Add(-1*time.Minute), now.Add(1*time.Hour))
			info := m.jwtToSessionInfo(pubJWKSet, makeHeader(token))
			So(info.IsValid, ShouldBeTrue)
			So(info.UserID, ShouldEqual, "user123")
			So(info.UserVerified, ShouldBeTrue)
			So(info.UserCanReauthenticate, ShouldBeTrue)
		})

		Convey("id_token with typ JWT is rejected", func() {
			token := signToken(t, privJWKSet, "JWT", "user123", now.Add(-1*time.Minute), now.Add(1*time.Hour))
			info := m.jwtToSessionInfo(pubJWKSet, makeHeader(token))
			So(info.IsValid, ShouldBeFalse)
		})

		Convey("missing Authorization header returns invalid session", func() {
			info := m.jwtToSessionInfo(pubJWKSet, http.Header{})
			So(info.IsValid, ShouldBeFalse)
		})

		Convey("malformed Authorization header returns invalid session", func() {
			h := http.Header{}
			h.Set("Authorization", "notabearer")
			info := m.jwtToSessionInfo(pubJWKSet, h)
			So(info.IsValid, ShouldBeFalse)
		})

		Convey("expired token returns invalid session", func() {
			token := signToken(t, privJWKSet, "at+jwt", "user123", now.Add(-2*time.Hour), now.Add(-1*time.Hour))
			info := m.jwtToSessionInfo(pubJWKSet, makeHeader(token))
			So(info.IsValid, ShouldBeFalse)
		})

		Convey("token signed with a different key returns invalid session", func() {
			// Sign with our private key but verify against an empty JWK set — no matching key.
			emptyPubSet := jwk.NewSet()
			token := signToken(t, privJWKSet, "at+jwt", "user123", now.Add(-1*time.Minute), now.Add(1*time.Hour))
			info := m.jwtToSessionInfo(emptyPubSet, makeHeader(token))
			So(info.IsValid, ShouldBeFalse)
		})
	})
}
