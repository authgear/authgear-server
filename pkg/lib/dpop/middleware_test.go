package dpop

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

// failingProofReplayStore stands in for the store being unreachable.
type failingProofReplayStore struct {
	calls int
}

func (s *failingProofReplayStore) MarkProofUsed(ctx context.Context, proof *DPoPProof) (bool, error) {
	s.calls++
	return false, errors.New("redis is down")
}

type memoryProofReplayStore struct {
	seen map[string]struct{}
}

func (s *memoryProofReplayStore) MarkProofUsed(ctx context.Context, proof *DPoPProof) (bool, error) {
	if s.seen == nil {
		s.seen = make(map[string]struct{})
	}
	key := proof.JKT + ":" + proof.JTI
	if _, ok := s.seen[key]; ok {
		return true, nil
	}
	s.seen[key] = struct{}{}
	return false, nil
}

func TestMiddlewareProofReplay(t *testing.T) {
	Convey("DPoP proof replay", t, func() {
		const origin = "http://example.com"
		const htm = "POST"
		const htu = origin + "/path"

		mockClock := clock.NewMockClockAt("2006-01-02T03:04:05Z")

		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		So(err, ShouldBeNil)
		privateJWK, err := jwk.FromRaw(privateKey)
		So(err, ShouldBeNil)
		publicJWK, err := privateJWK.PublicKey()
		So(err, ShouldBeNil)
		_ = publicJWK.Set(jwk.AlgorithmKey, jwa.ES256)

		signProof := func(jti string) string {
			header := jws.NewHeaders()
			_ = header.Set("typ", DPoPJWTTyp)
			_ = header.Set("alg", "ES256")
			_ = header.Set("jwk", publicJWK)

			payload := jwt.New()
			_ = payload.Set("jti", jti)
			_ = payload.Set("htm", htm)
			_ = payload.Set("htu", htu)
			_ = payload.Set("iat", mockClock.NowUTC().Unix())

			payloadBytes, err := jwt.NewSerializer().Serialize(payload)
			So(err, ShouldBeNil)
			signed, err := jws.Sign(payloadBytes, jws.WithKey(jwa.ES256, privateJWK, jws.WithProtectedHeaders(header)))
			So(err, ShouldBeNil)
			return string(signed)
		}

		store := &memoryProofReplayStore{}
		m := &Middleware{
			DPoPProvider: &Provider{
				Clock:      mockClock,
				HTTPOrigin: httputil.HTTPOrigin(origin),
			},
			ProofReplayStore: store,
		}

		// A proof that fails validation is replaced in the context with
		// InvalidDPoPProofWithError rather than rejected outright, so what the
		// handler receives is what distinguishes accepted from rejected.
		var lastProof any
		var nextCalled int
		next := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			nextCalled++
			lastProof = GetDPoPProof(r.Context())
		})
		h := m.Handle(next)

		serve := func(proofJWT string) {
			r, _ := http.NewRequest(htm, htu, nil)
			r.Header.Set("DPoP", proofJWT)
			h.ServeHTTP(httptest.NewRecorder(), r)
		}

		Convey("accepts a proof once and rejects the replay", func() {
			proofJWT := signProof("jti-1")

			serve(proofJWT)
			So(nextCalled, ShouldEqual, 1)
			proof, ok := lastProof.(*DPoPProof)
			So(ok, ShouldBeTrue)
			So(proof.JTI, ShouldEqual, "jti-1")

			// The very same proof, presented again inside its validity window.
			serve(proofJWT)
			So(nextCalled, ShouldEqual, 2)
			invalid, ok := lastProof.(*InvalidDPoPProofWithError)
			So(ok, ShouldBeTrue)
			So(invalid.Error, ShouldBeError, ErrProofReplayed)
		})

		Convey("accepts distinct proofs from the same key", func() {
			serve(signProof("jti-1"))
			_, ok := lastProof.(*DPoPProof)
			So(ok, ShouldBeTrue)

			serve(signProof("jti-2"))
			proof, ok := lastProof.(*DPoPProof)
			So(ok, ShouldBeTrue)
			So(proof.JTI, ShouldEqual, "jti-2")
		})

		Convey("still serves the request when the replay store is unavailable", func() {
			// A Redis outage must not take down every DPoP-authenticated
			// endpoint. Replay becomes possible; nothing else is relaxed.
			failing := &failingProofReplayStore{}
			m.ProofReplayStore = failing

			proofJWT := signProof("jti-1")

			serve(proofJWT)
			So(failing.calls, ShouldEqual, 1)
			proof, ok := lastProof.(*DPoPProof)
			So(ok, ShouldBeTrue)
			So(proof.JTI, ShouldEqual, "jti-1")

			// The same proof again also succeeds: this is the replay that the
			// store would have caught, and it is the accepted cost.
			serve(proofJWT)
			proof, ok = lastProof.(*DPoPProof)
			So(ok, ShouldBeTrue)
			So(proof.JTI, ShouldEqual, "jti-1")
		})

		Convey("still applies the other checks when the store is unavailable", func() {
			m.ProofReplayStore = &failingProofReplayStore{}

			// Wrong method: the proof is not bound to this request, and that
			// must still be rejected even with the replay check degraded.
			r, _ := http.NewRequest("GET", htu, nil)
			r.Header.Set("DPoP", signProof("jti-1"))
			h.ServeHTTP(httptest.NewRecorder(), r)
			invalid, ok := lastProof.(*InvalidDPoPProofWithError)
			So(ok, ShouldBeTrue)
			So(invalid.Error, ShouldBeError, ErrUnmatchedMethod)

			// A garbage proof is still rejected too.
			r, _ = http.NewRequest(htm, htu, nil)
			r.Header.Set("DPoP", "not-a-jwt")
			h.ServeHTTP(httptest.NewRecorder(), r)
			_, ok = lastProof.(*InvalidDPoPProofWithError)
			So(ok, ShouldBeTrue)
		})

		Convey("does not consume a jti for a proof that fails validation", func() {
			proofJWT := signProof("jti-1")

			// Wrong method, so the proof is not bound to this request.
			r, _ := http.NewRequest("GET", htu, nil)
			r.Header.Set("DPoP", proofJWT)
			h.ServeHTTP(httptest.NewRecorder(), r)
			_, ok := lastProof.(*InvalidDPoPProofWithError)
			So(ok, ShouldBeTrue)

			// The legitimate holder can still use it.
			serve(proofJWT)
			proof, ok := lastProof.(*DPoPProof)
			So(ok, ShouldBeTrue)
			So(proof.JTI, ShouldEqual, "jti-1")
		})
	})
}
