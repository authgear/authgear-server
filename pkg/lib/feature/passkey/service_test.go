package passkey

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/protocol/webauthncose"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
)

const packedAttestationResponseES256 = `{
	"rawId": "hUf7WI3IZmoLOzYhHFe7U-df4QD17lQBMi9iS-z3dWFlr79MXOoTR8dJzb_Y7sAstHBrcC1nv8pOr6aFz50K65juYXWt8k26bKu-Hu4CulPo53bIStJ4kpOr2Dlr6Z4D",
	"id": "hUf7WI3IZmoLOzYhHFe7U-df4QD17lQBMi9iS-z3dWFlr79MXOoTR8dJzb_Y7sAstHBrcC1nv8pOr6aFz50K65juYXWt8k26bKu-Hu4CulPo53bIStJ4kpOr2Dlr6Z4D",
	"response": {
	  "clientDataJSON": "ew0KCSJ0eXBlIiA6ICJ3ZWJhdXRobi5jcmVhdGUiLA0KCSJjaGFsbGVuZ2UiIDogIlBfSktRaWQxdHZzNEJsdGlaMUNzRWZYbDNHWjBJcG1MUFVRRmxZLW8weDlzZ3ZDS3lXNXpQUkpjTzc3M2VpOE93WEN5Rjl1Wk42X3B5elhOT0FKUjdBIiwNCgkib3JpZ2luIiA6ICJodHRwczovL2xvY2FsaG9zdDo0NDMyOSIsDQoJInRva2VuQmluZGluZyIgOiANCgl7DQoJCSJzdGF0dXMiIDogInN1cHBvcnRlZCINCgl9DQp9",
	  "attestationObject": "o2NmbXRmcGFja2VkaGF1dGhEYXRhWORJlg3liA6MaHQ0Fw9kdmBbj-SuuaKGMseZXPO6gx2XY0UAAChiQjgyRUQ3M0M4RkI0RTVBMgBghUf7WI3IZmoLOzYhHFe7U-df4QD17lQBMi9iS-z3dWFlr79MXOoTR8dJzb_Y7sAstHBrcC1nv8pOr6aFz50K65juYXWt8k26bKu-Hu4CulPo53bIStJ4kpOr2Dlr6Z4DpQECAyYgASFYIA9RHvpjfWoWN_Im7eYwG1Y8kA77s7QH9uf9TePknT3mIlggJ8tNsMrPPrewstqf65ItALMxBIi4VUoTIZEyAkXN6U1nYXR0U3RtdKNjYWxnJmNzaWdYRzBFAiBsbcx3U1xgYinrnczLOUDOlYGvYENDGzv77WdM1W3FTQIhAJ16HUK8XyG83cOVQFKkijdgHyDV97XylRMU_rWHAkP_Y3g1Y4NZAkUwggJBMIIB6KADAgECAhAVn3vCzYkY8Shrk0j6nzPiMAoGCCqGSM49BAMCMEkxCzAJBgNVBAYTAkNOMR0wGwYDVQQKDBRGZWl0aWFuIFRlY2hub2xvZ2llczEbMBkGA1UEAwwSRmVpdGlhbiBGSURPMiBDQS0xMCAXDTE4MDQxMTAwMDAwMFoYDzIwMzMwNDEwMjM1OTU5WjBvMQswCQYDVQQGEwJDTjEdMBsGA1UECgwURmVpdGlhbiBUZWNobm9sb2dpZXMxIjAgBgNVBAsMGUF1dGhlbnRpY2F0b3IgQXR0ZXN0YXRpb24xHTAbBgNVBAMMFEZUIEJpb1Bhc3MgRklETzIgVVNCMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEgAZ1XFn7yUmwFajSCpJYl76DCrLv6Cz4j-2gkJZj5UjHHxEnBTO0JEZ4nUz-4QFDipTpgz3iACwvKh3Xb03bXaOBiTCBhjAdBgNVHQ4EFgQUelSCQoBi2Irnr4SYJcSvkak0mPIwHwYDVR0jBBgwFoAUTTvYxGcVG7sT6POE2DBPnWkVwIMwDAYDVR0TAQH_BAIwADATBgsrBgEEAYLlHAIBAQQEAwIFIDAhBgsrBgEEAYLlHAEBBAQSBBBCODJFRDczQzhGQjRFNUEyMAoGCCqGSM49BAMCA0cAMEQCICRLRaO-iNy34CWixqMSz_uG7bwnSiLBBS4xSFHw6LCHAiA0Gr9OHCTyCxpz1T2swqn5FbQbsjprAW8f7_jg5_iQwFkB_zCCAfswggGgoAMCAQICEBWfe8LNiRjxKGuTSPqfM-EwCgYIKoZIzj0EAwIwSzELMAkGA1UEBhMCQ04xHTAbBgNVBAoMFEZlaXRpYW4gVGVjaG5vbG9naWVzMR0wGwYDVQQDDBRGZWl0aWFuIEZJRE8gUm9vdCBDQTAgFw0xODA0MTAwMDAwMDBaGA8yMDM4MDQwOTIzNTk1OVowSTELMAkGA1UEBhMCQ04xHTAbBgNVBAoMFEZlaXRpYW4gVGVjaG5vbG9naWVzMRswGQYDVQQDDBJGZWl0aWFuIEZJRE8yIENBLTEwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAASOfmAJ7MEWZcyg-sPpb-UIO5VtVyUR61sy9NZnOVfdZ9i2FzUd_0u5gOYLqbkzuZo0MPMX6iETB1a9agd03nWPo2YwZDAdBgNVHQ4EFgQUTTvYxGcVG7sT6POE2DBPnWkVwIMwHwYDVR0jBBgwFoAU0aGYTYF_w7lr9gdnvVAS_pBF8VQwEgYDVR0TAQH_BAgwBgEB_wIBADAOBgNVHQ8BAf8EBAMCAQYwCgYIKoZIzj0EAwIDSQAwRgIhAPt_o9JAR6ERUMJ4Vm0hzJAWmOyhf087SDRTecpg5MJlAiEA6wpDwYjB172IPpEkYFbCsLlbWKJ0bwufPKkcKS0rWexZAdwwggHYMIIBfqADAgECAhAVn3vCzYkY8Shrk0j6nzPWMAoGCCqGSM49BAMCMEsxCzAJBgNVBAYTAkNOMR0wGwYDVQQKDBRGZWl0aWFuIFRlY2hub2xvZ2llczEdMBsGA1UEAwwURmVpdGlhbiBGSURPIFJvb3QgQ0EwIBcNMTgwNDAxMDAwMDAwWhgPMjA0ODAzMzEyMzU5NTlaMEsxCzAJBgNVBAYTAkNOMR0wGwYDVQQKDBRGZWl0aWFuIFRlY2hub2xvZ2llczEdMBsGA1UEAwwURmVpdGlhbiBGSURPIFJvb3QgQ0EwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAASd8ApuO8xfUTLVvqT5ZBB01Uy30mAZbInc-8zgFIrlepN-j77SgCP_i2fDIgvQcUFH1K36S2OpJcN-OJcC6uzzo0IwQDAdBgNVHQ4EFgQU0aGYTYF_w7lr9gdnvVAS_pBF8VQwDwYDVR0TAQH_BAUwAwEB_zAOBgNVHQ8BAf8EBAMCAQYwCgYIKoZIzj0EAwIDSAAwRQIhALexPWUGMZ4X7EpOnNXUphTZyRqFN3iYsnLNg6Foe_iKAiAPYliR_IflDgGmjyuug7Qi3uhiMXaSDL95JndT0aVqrA"
	},
	"type": "public-key"
}`

const packedAttestationChallenge = "P_JKQid1tvs4BltiZ1CsEfXl3GZ0IpmLPUQFlY-o0x9sgvCKyW5zPRJcO773ei8OwXCyF9uZN6_pyzXNOAJR7A"

type testTranslationService struct{}

func (s *testTranslationService) RenderText(ctx context.Context, key string, args interface{}) (string, error) {
	return "Test App", nil
}

func TestPeekAttestationResponse(t *testing.T) {
	Convey("PeekAttestationResponse", t, func() {
		Convey("accepts valid attestation for matching origin", func() {
			svc, cleanup := newTestService(t, "localhost:44329")
			defer cleanup()

			attestationResponse := []byte(packedAttestationResponseES256)
			expectedCredentialID := mustParseCredentialID(t, attestationResponse)

			mustCreateSession(t, svc.Store, packedAttestationChallenge)

			creationOptions, credentialID, _, err := svc.PeekAttestationResponse(context.Background(), attestationResponse)
			So(err, ShouldBeNil)
			So(creationOptions, ShouldNotBeNil)
			So(credentialID, ShouldEqual, expectedCredentialID)
		})

		Convey("rejects attestation when origin does not match", func() {
			svc, cleanup := newTestService(t, "example.com")
			defer cleanup()

			mustCreateSession(t, svc.Store, packedAttestationChallenge)

			_, _, _, err := svc.PeekAttestationResponse(context.Background(), []byte(packedAttestationResponseES256))
			So(err, ShouldNotBeNil)
			So(strings.ToLower(err.Error()), ShouldContainSubstring, "origin")
		})

		Convey("uses attested credential ID instead of client supplied id", func() {
			svc, cleanup := newTestService(t, "localhost:44329")
			defer cleanup()

			attestationResponse := mustOverrideCredentialID(t, packedAttestationResponseES256, "AQID")
			expectedCredentialID := mustParseCredentialID(t, attestationResponse)

			mustCreateSession(t, svc.Store, packedAttestationChallenge)

			_, credentialID, _, err := svc.PeekAttestationResponse(context.Background(), attestationResponse)
			So(err, ShouldBeNil)
			So(credentialID, ShouldEqual, expectedCredentialID)
			So(credentialID, ShouldNotEqual, "AQID")
		})
	})
}

func newTestService(t *testing.T, host string) (*Service, func()) {
	t.Helper()

	mr := miniredis.RunT(t)
	pool := redis.NewPool()
	hub := redis.NewHub(context.Background(), pool)
	redisConfig := &config.RedisEnvironmentConfig{}
	redisCredentials := &config.RedisCredentials{
		RedisURL: "redis://" + mr.Addr(),
	}

	req := httptest.NewRequest("GET", "https://"+host, nil)
	req.TLS = &tls.ConnectionState{}

	svc := &Service{
		Store: &Store{
			Redis: appredis.NewHandle(pool, hub, redisConfig, redisCredentials),
			AppID: config.AppID("test"),
		},
		ConfigService: &ConfigService{
			Request:            req,
			TranslationService: &testTranslationService{},
		},
	}

	cleanup := func() {
		_ = pool.Close()
		mr.Close()
	}

	return svc, cleanup
}

func mustCreateSession(t *testing.T, store *Store, challenge string) {
	t.Helper()

	challengeBytes, err := base64.RawURLEncoding.DecodeString(challenge)
	if err != nil {
		t.Fatalf("failed to decode challenge: %v", err)
	}

	err = store.CreateSession(context.Background(), &Session{
		Challenge: challengeBytes,
		CreationOptions: &model.WebAuthnCreationOptions{
			PublicKey: model.PublicKeyCredentialCreationOptions{
				Challenge: challengeBytes,
				PublicKeyCredentialParameters: []model.PublicKeyCredentialParameter{
					{
						Type:      protocol.PublicKeyCredentialType,
						Algorithm: webauthncose.AlgES256,
					},
					{
						Type:      protocol.PublicKeyCredentialType,
						Algorithm: webauthncose.AlgRS256,
					},
				},
				AuthenticatorSelection: protocol.AuthenticatorSelection{
					UserVerification: protocol.VerificationPreferred,
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}
}

func mustParseCredentialID(t *testing.T, attestationResponse []byte) string {
	t.Helper()

	parsed, err := protocol.ParseCredentialCreationResponseBody(strings.NewReader(string(attestationResponse)))
	if err != nil {
		t.Fatalf("failed to parse attestation response: %v", err)
	}

	return base64.RawURLEncoding.EncodeToString(parsed.Response.AttestationObject.AuthData.AttData.CredentialID)
}

func mustOverrideCredentialID(t *testing.T, attestationResponse string, credentialID string) []byte {
	t.Helper()

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(attestationResponse), &payload); err != nil {
		t.Fatalf("failed to unmarshal attestation response: %v", err)
	}

	payload["id"] = credentialID

	out, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal attestation response: %v", err)
	}

	return out
}
