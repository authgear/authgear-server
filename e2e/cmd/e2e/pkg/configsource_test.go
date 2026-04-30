package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/authgear/authgear-server/pkg/api/model"
	. "github.com/smartystreets/goconvey/convey"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/config"
	_ "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/azureadv2"
	_ "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/google"
)

func TestConfigOverrideMergeForWebappSignup(t *testing.T) {
	Convey("config override merge for webapp signup overwrites nested false values", t, func() {
		baseConfigPath := filepath.Join("..", "..", "..", "var", "authgear.yaml")
		baseYAML, err := os.ReadFile(baseConfigPath)
		So(err, ShouldBeNil)

		overrideYAML := `
authentication:
  identities:
  - login_id
  primary_authenticators:
  - password
identity:
  login_id:
    keys:
    - key: email
      type: email
verification:
  claims:
    email:
      enabled: true
      required: false
    phone_number:
      enabled: true
      required: true
`

		var overrideCfg config.AppConfig
		overrideJSON, err := yaml.YAMLToJSON([]byte(overrideYAML))
		So(err, ShouldBeNil)

		decoder := json.NewDecoder(bytes.NewReader(overrideJSON))
		err = decoder.Decode(&overrideCfg)
		So(err, ShouldBeNil)
		So(*overrideCfg.Verification.Claims.Email.Required, ShouldBeFalse)

		mergedYAML, err := mergeYAMLObjects(baseYAML, []byte(overrideYAML))
		So(err, ShouldBeNil)

		cfg, err := config.Parse(context.Background(), mergedYAML)
		So(err, ShouldBeNil)

		So(*cfg.Verification.Claims.Email.Required, ShouldBeFalse)
		So(cfg.Authentication.Identities, ShouldResemble, []model.IdentityType{
			model.IdentityTypeLoginID,
		})
		So(*cfg.Authentication.PrimaryAuthenticators, ShouldResemble, []model.AuthenticatorType{
			model.AuthenticatorTypePassword,
		})
		So(cfg.Identity.LoginID.Keys, ShouldHaveLength, 1)
		So(cfg.Identity.LoginID.Keys[0].Type, ShouldEqual, model.LoginIDKeyTypeEmail)

		flow := declarative.GenerateSignupFlowConfig(cfg)

		flowJSON, err := json.Marshal(flow)
		So(err, ShouldBeNil)

		expectedYAML := `
name: default
steps:
- name: signup_identify
  type: identify
  one_of:
  - identification: email
    steps:
    - name: authenticate_primary_email
      type: create_authenticator
      one_of:
      - authentication: primary_password
`
		expectedJSON, err := yaml.YAMLToJSON([]byte(expectedYAML))
		So(err, ShouldBeNil)

		So(string(flowJSON), ShouldEqualJSON, string(expectedJSON))
	})
}

func TestMergeYAMLObjectsNestedBoolOverride(t *testing.T) {
		Convey("mergeYAMLObjects overwrites nested false without removing unrelated pointers", t, func() {
			baseYAML := []byte(`
id: test
http:
  public_origin: http://example.com
verification:
  claims:
    email:
      enabled: true
      required: true
`)
		overrideYAML := []byte(`
verification:
  claims:
    email:
      required: false
`)

		mergedYAML, err := mergeYAMLObjects(baseYAML, overrideYAML)
		So(err, ShouldBeNil)

		cfg, err := config.Parse(context.Background(), mergedYAML)
		So(err, ShouldBeNil)
		So(cfg.HTTP, ShouldNotBeNil)
		So(cfg.HTTP.PublicOrigin, ShouldEqual, "http://example.com")
		So(*cfg.Verification.Claims.Email.Required, ShouldBeFalse)
	})
}
