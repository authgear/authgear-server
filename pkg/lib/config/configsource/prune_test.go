package configsource

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/filepathutil"
)

const pruneTestAuthgearYAML = `id: test
http:
  public_origin: http://test.authgear.example.com
oauth:
  clients:
  - client_id: client-id
    redirect_uris:
    - http://test.authgear.example.com/
`

const pruneTestAuthgearSecretYAML = `secrets:
- key: db
  data:
    database_url: postgres://postgres@127.0.0.1:5432/postgres
- key: oauth
  data:
    keys:
    - created_at: 1136214245
      k: c2VjcmV0MQ
      kid: oauth-kid
      kty: oct
- key: admin-api.auth
  data:
    keys:
    - created_at: 1136214245
      k: c2VjcmV0Mg
      kid: admin-api-kid
      kty: oct
- key: csrf
  data:
    keys:
    - created_at: 1136214245
      k: c2VjcmV0Mw
      kid: csrf-kid
      kty: oct
- key: mail.smtp
  data:
    host: 127.0.0.1
    port: 25
    username: user
    password: secret
    sender: Authgear <noreply@authgear.com>
`

func TestTrimDataForPrune(t *testing.T) {
	ctx := context.Background()

	Convey("TrimDataForPrune", t, func() {
		data := map[string][]byte{
			filepathutil.EscapePath(AuthgearYAML):                              []byte(pruneTestAuthgearYAML),
			filepathutil.EscapePath(AuthgearSecretYAML):                        []byte(pruneTestAuthgearSecretYAML),
			filepathutil.EscapePath("templates/en/messages/welcome_email.txt"): []byte("hello"),
		}

		trimmed, err := TrimDataForPrune(ctx, data)
		So(err, ShouldBeNil)

		Convey("it drops every resource other than authgear.yaml and authgear.secrets.yaml", func() {
			So(len(trimmed), ShouldEqual, 2)
			_, ok := trimmed[filepathutil.EscapePath(AuthgearYAML)]
			So(ok, ShouldBeTrue)
			_, ok = trimmed[filepathutil.EscapePath(AuthgearSecretYAML)]
			So(ok, ShouldBeTrue)
		})

		Convey("it reduces authgear.yaml to id and http.public_origin", func() {
			appConfig, err := config.Parse(ctx, trimmed[filepathutil.EscapePath(AuthgearYAML)])
			So(err, ShouldBeNil)
			So(string(appConfig.ID), ShouldEqual, "test")
			So(appConfig.HTTP.PublicOrigin, ShouldEqual, "http://test.authgear.example.com")
			So(appConfig.OAuth.Clients, ShouldBeEmpty)

			jsonBytes, err := yaml.YAMLToJSON(trimmed[filepathutil.EscapePath(AuthgearYAML)])
			So(err, ShouldBeNil)
			So(string(jsonBytes), ShouldEqualJSON, `{
				"id": "test",
				"http": { "public_origin": "http://test.authgear.example.com" }
			}`)
		})

		Convey("it keeps only oauth, admin-api.auth, and csrf secrets", func() {
			secretConfig, err := config.ParseSecret(ctx, trimmed[filepathutil.EscapePath(AuthgearSecretYAML)])
			So(err, ShouldBeNil)
			So(len(secretConfig.Secrets), ShouldEqual, 3)

			_, _, hasOAuth := secretConfig.Lookup(config.OAuthKeyMaterialsKey)
			So(hasOAuth, ShouldBeTrue)
			_, _, hasAdminAPI := secretConfig.Lookup(config.AdminAPIAuthKeyKey)
			So(hasAdminAPI, ShouldBeTrue)
			_, _, hasCSRF := secretConfig.Lookup(config.CSRFKeyMaterialsKey)
			So(hasCSRF, ShouldBeTrue)
			_, _, hasDB := secretConfig.Lookup(config.DatabaseCredentialsKey)
			So(hasDB, ShouldBeFalse)
			_, _, hasSMTP := secretConfig.Lookup(config.SMTPServerCredentialsKey)
			So(hasSMTP, ShouldBeFalse)
		})
	})

	Convey("TrimDataForPrune is a no-op when the app has neither file", t, func() {
		trimmed, err := TrimDataForPrune(ctx, map[string][]byte{
			filepathutil.EscapePath("templates/en/messages/welcome_email.txt"): []byte("hello"),
		})
		So(err, ShouldBeNil)
		So(trimmed, ShouldBeEmpty)
	})
}
