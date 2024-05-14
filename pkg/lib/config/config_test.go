package config_test

import (
	"errors"
	"io"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	goyaml "gopkg.in/yaml.v2"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/config"
	_ "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/adfs"
	_ "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/apple"
	_ "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/azureadb2c"
	_ "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/azureadv2"
	_ "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/facebook"
	_ "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/google"
	_ "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/linkedin"
	_ "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/wechat"
)

func TestAppConfig(t *testing.T) {
	Convey("AppConfig", t, func() {
		fixture := `id: test
http:
  public_origin: http://test
identity:
  oauth:
    providers:
    - type: google
      alias: google
      client_id: a
    - type: facebook
      alias: facebook
      client_id: a
    - type: linkedin
      alias: linkedin
      client_id: a
    - type: azureadv2
      alias: azureadv2
      client_id: a
      tenant: a
    - type: azureadb2c
      alias: azureadb2c
      client_id: a
      tenant: a
      policy: a
    - type: adfs
      alias: adfs
      client_id: a
      discovery_document_endpoint: http://test
    - type: apple
      alias: apple
      client_id: a
      key_id: a
      team_id: a
    - type: wechat
      alias: wechat
      client_id: a
      app_type: web
      account_id: gh_
`

		Convey("populate default values", func() {
			cfg, err := config.Parse([]byte(fixture))
			So(err, ShouldBeNil)

			data, err := os.ReadFile("testdata/default_config.yaml")
			if err != nil {
				panic(err)
			}

			_, err = config.Parse(data)
			So(err, ShouldBeNil)

			var defaultCfg config.AppConfig
			err = yaml.Unmarshal(data, &defaultCfg)
			if err != nil {
				panic(err)
			}

			So(cfg, ShouldResemble, &defaultCfg)
		})

		Convey("round-trip default configuration", func() {
			cfg, err := config.Parse([]byte(fixture))
			So(err, ShouldBeNil)

			data, err := yaml.Marshal(cfg)
			So(err, ShouldBeNil)

			cfg2, err := config.Parse(data)
			So(err, ShouldBeNil)
			So(cfg, ShouldResemble, cfg2)
		})

		Convey("parse validation", func() {
			f, err := os.Open("testdata/config_tests.yaml")
			if err != nil {
				panic(err)
			}
			defer f.Close()

			type TestCase struct {
				Name   string      `yaml:"name"`
				Error  *string     `yaml:"error"`
				Config interface{} `yaml:"config"`
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
					data, err := goyaml.Marshal(testCase.Config)
					if err != nil {
						panic(err)
					}

					_, err = config.Parse(data)
					if testCase.Error != nil {
						So(err, ShouldBeError, *testCase.Error)
					} else {
						So(err, ShouldBeNil)
					}
				})
			}
		})
	})
}
