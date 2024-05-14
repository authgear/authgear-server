package config_test

import (
	"encoding/json"
	"errors"
	"fmt"
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
	var GenerateOctetKeyFunc = func(createdAt time.Time, rng *mathrand.Rand) jwk.Key {
		key := []byte("secret1")

		jwkKey, err := jwk.FromRaw(key)
		if err != nil {
			panic(err)
		}

		fmt.Println("createdAt", createdAt)
		_ = jwkKey.Set(jwk.KeyIDKey, "kid")
		_ = jwkKey.Set(jwkutil.KeyCreatedAt, float64(createdAt.Unix()))
		return jwkKey
	}
	var GenerateRSAKeyFunc = func(createdAt time.Time, rng *mathrand.Rand) jwk.Key {
		// Use octet key instead for smaller testcase file size
		return GenerateOctetKeyFunc(createdAt, rng)
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
					Clock:                            clock.NewMockClockAt("2006-01-02T15:04:05Z"),
					GenerateClientSecretOctetKeyFunc: GenerateOctetKeyFunc,
					GenerateAdminAPIAuthKeyFunc:      GenerateRSAKeyFunc,
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
