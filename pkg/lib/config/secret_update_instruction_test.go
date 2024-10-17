package config_test

import (
	"encoding/json"
	"errors"
	"io"
	mathrand "math/rand"
	"os"
	"testing"
	"time"

	goyaml "gopkg.in/yaml.v2"
	"sigs.k8s.io/yaml"

	"github.com/lestrrat-go/jwx/v2/jwk"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/jwkutil"
)

func TestSecretConfigUpdateInstruction(t *testing.T) {

	mockClock := clock.NewMockClockAt("2006-01-02T15:04:05Z")
	var GenerateFixedOctetKeyFunc = func(createdAt time.Time) jwk.Key {
		key := []byte("secret1")

		jwkKey, err := jwk.FromRaw(key)
		if err != nil {
			panic(err)
		}

		_ = jwkKey.Set(jwk.KeyIDKey, "kid")
		_ = jwkKey.Set(jwkutil.KeyCreatedAt, float64(createdAt.Unix()))
		return jwkKey
	}
	var GenerateOctetKeyFunc = func(createdAt time.Time, rng *mathrand.Rand) jwk.Key {
		return GenerateFixedOctetKeyFunc(createdAt)
	}
	var GenerateRSAKeyFunc = func(createdAt time.Time, rng *mathrand.Rand) jwk.Key {
		// Use octet key instead for smaller testcase file size
		return GenerateFixedOctetKeyFunc(createdAt)
	}
	var GenerateSAMLIdpSigningCertificate = func() (*config.SAMLIdpSigningCertificate, error) {
		signingSecret := &config.SAMLIdpSigningCertificate{
			Certificate: &config.X509Certificate{
				Pem: config.X509CertificatePem(`-----BEGIN CERTIFICATE-----
MIIC1TCCAb2gAwIBAgIRAJpxx1DW2ObGLT5lUpXARWkwDQYJKoZIhvcNAQELBQAw
GzEZMBcGA1UEAxMQbXktYXBwLmxvY2FsaG9zdDAgFw0yNDA4MDkwODA3MzlaGA8y
MDc0MDcyODA4MDczOVowGzEZMBcGA1UEAxMQbXktYXBwLmxvY2FsaG9zdDCCASIw
DQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAN83SCP6m3ayNriEX6VLiwCqoIHu
E1d2vFwULyUWOjinI3olWWkA1txAZu2e0Rm+Zslq2sWx/HZ5e83NCzyLQ8aaG1JQ
OtpbxV2IOybOonveZr1qszvs+1ofGw9sW6AZa7vhH9HhuDqZnM6ArsC7E/D03D4x
J/2hb6uVj9zHb+Cx4vh1nAnBXXwOSIuo1Jm4a0vZHFs8HT2gmX31K/5hhJuchqiH
ptqerf0OHq/Zyx+v40oj3/cFwGAJ291z6kv318bfjBhZTdQ2ovbnFnU9NfQ02IgW
tSj1Grr8dAp5aIDZvgvvYg/m+FnyMqrSU5s0NIyn13tqipZgN4YUk8CUkCECAwEA
AaMSMBAwDgYDVR0PAQH/BAQDAgeAMA0GCSqGSIb3DQEBCwUAA4IBAQAVuZEbgLi0
gzKy5x+L1j+uQMFdY4taFWGdTF7gZx/hw2YpKakPSCl/Sb+624u3+XhQSzByjt7m
0yGhAml5aLQ+y7jOAwagL0pWhK/AW6kZKU2lz36J+T8LTzq3YOFBHrLTJ58ZcWKe
kgwAWDr8Uj9BgxnQWF4Rwu8yAP8POV4E6aIajalFK3tNdyGaXIS5rSHGd/QKuJNW
eCHF7sKGUSTw3p3MADXGkDykUCuXevyNACH6opOLrDCHr/uEEFmSTVf5zlIeSk+Y
EMgvAyAtQw4fi3WItQNOSLm+01kxkCC1SF+LXTSUPMsLOnX++WJ4u4VJTMfqrh6d
UgPkRnolBQXT
-----END CERTIFICATE-----`),
			},
			Key: &config.JWK{
				Key: GenerateFixedOctetKeyFunc(mockClock.NowUTC()),
			},
		}
		return signingSecret, nil
	}
	Convey("SecretConfigUpdateInstruction", t, func() {
		f, err := os.Open("testdata/secret_update_instruction.yaml")
		if err != nil {
			panic(err)
		}
		defer f.Close()

		type TestCase struct {
			Name                    string  `yaml:"name"`
			Error                   *string `yaml:"error"`
			CurrentSecretConfigYAML string  `yaml:"currentSecretConfigYAML"`
			NewSecretConfigYAML     string  `yaml:"newSecretConfigYAML"`
			UpdateInstructionJSON   string  `yaml:"updateInstructionJSON"`
		}

		decoder := goyaml.NewDecoder(f)
		for {
			var testCase TestCase
			err := decoder.Decode(&testCase)
			if errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				panic(err)
			}

			Convey(testCase.Name, func() {
				var err error

				currentSecretConfig, err := config.ParseSecret([]byte(testCase.CurrentSecretConfigYAML))
				So(err, ShouldBeNil)

				var updateInstruction *config.SecretConfigUpdateInstruction
				err = json.Unmarshal([]byte(testCase.UpdateInstructionJSON), &updateInstruction)
				So(err, ShouldBeNil)

				updateInstructionContext := &config.SecretConfigUpdateInstructionContext{
					Clock:                             mockClock,
					GenerateClientSecretOctetKeyFunc:  GenerateOctetKeyFunc,
					GenerateAdminAPIAuthKeyFunc:       GenerateRSAKeyFunc,
					GenerateSAMLIdpSigningCertificate: GenerateSAMLIdpSigningCertificate,
				}
				actualNewSecretConfig, err := updateInstruction.ApplyTo(updateInstructionContext, currentSecretConfig)
				if testCase.Error != nil {
					So(err, ShouldBeError, *testCase.Error)
				} else {
					So(err, ShouldBeNil)

					var expectedNewSecretConfig *config.SecretConfig
					err = yaml.Unmarshal([]byte(testCase.NewSecretConfigYAML), &expectedNewSecretConfig)
					So(err, ShouldBeNil)

					// Compare the secret config
					So(len(actualNewSecretConfig.Secrets), ShouldEqual, len(expectedNewSecretConfig.Secrets))
					for _, actualItem := range actualNewSecretConfig.Secrets {
						_, expectedItem, ok := expectedNewSecretConfig.Lookup(actualItem.Key)
						So(ok, ShouldBeTrue)
						So(string(actualItem.RawData), ShouldEqualJSON, string(expectedItem.RawData))
					}
				}
			})
		}
	})
}
