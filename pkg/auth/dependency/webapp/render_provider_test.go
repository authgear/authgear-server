package webapp

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

const templateContent = `
<!DOCTYPE html>
<html>
<head>
<title>title</title>
<style>
{{ .x_css }}
</style>
</head>
<body>
<p>{{ .client_name }}</p>
<p>{{ .logo_uri }}</p>
<img src="{{ .x_static_asset_url_prefix }}/logo.png">

{{ if (ge (len .x_calling_codes) 0) }}
<p>has calling codes</p>
{{ else }}
<p>has no calling codes</p>
{{ end }}

{{ if .x_login_id_input_type_has_phone }}
<p>has phone</p>
{{ else }}
<p>has no phone<p>
{{ end }}

{{ if .x_login_id_input_type_has_text }}
<p>has text</p>
{{ else }}
<p>has no text<p>
{{ end }}

{{ if .x_error }}
<p>has error</p>
{{ else }}
<p>has no error</p>
{{ end }}

</body>
</html>
`

func TestRenderProvider(t *testing.T) {
	Convey("RenderProvider", t, func() {
		engine := template.NewEngine(template.NewEngineOptions{})
		engine.Register(template.Spec{
			Type:    "a",
			IsHTML:  true,
			Default: templateContent,
		})

		impl := RenderProviderImpl{
			StaticAssetURLPrefix: "/static",
			AuthenticationConfiguration: &config.AuthenticationConfiguration{
				Identities: []string{"oauth", "login_id"},
			},
			IdentityConfiguration: &config.IdentityConfiguration{
				LoginID: &config.LoginIDConfiguration{
					Keys: []config.LoginIDKeyConfiguration{
						config.LoginIDKeyConfiguration{
							Key: "email", Type: "email",
						},
						config.LoginIDKeyConfiguration{
							Key: "phone", Type: "phone",
						},
					},
				},
				OAuth: &config.OAuthConfiguration{},
			},
			AuthUIConfiguration: &config.AuthUIConfiguration{
				CSS: `a { color: red; }`,
				CountryCallingCode: &config.AuthUICountryCallingCodeConfiguration{
					Values:  []string{"852"},
					Default: "852",
				},
			},
			TemplateEngine:  engine,
			PasswordChecker: &audit.PasswordChecker{},
		}

		w := httptest.NewRecorder()
		r, _ := http.NewRequest("GET", "", nil)
		r = r.WithContext(auth.WithAccessKey(r.Context(), auth.AccessKey{
			Client: config.OAuthClientConfiguration{
				"client_name": "App A",
				"logo_uri":    "https://example.com/logo.png",
			},
		}))

		impl.WritePage(w, r, "a", errors.New("error"))

		So(w.Result().Header.Get("Content-Type"), ShouldEqual, "text/html; charset=utf-8")
		So(string(w.Body.Bytes()), ShouldEqual, `
<!DOCTYPE html>
<html>
<head>
<title>title</title>
<style>
a { color: red; }
</style>
</head>
<body>
<p>App A</p>
<p>https://example.com/logo.png</p>
<img src="/static/logo.png">


<p>has calling codes</p>



<p>has phone</p>



<p>has text</p>



<p>has error</p>


</body>
</html>
`)
	})
}

func TestPreferredLanguageTags(t *testing.T) {
	Convey("PreferredLanguageTags", t, func() {
		test := func(uiLocales string, acceptLanguage string, expected []string) {
			r, _ := http.NewRequest("GET", "", nil)
			_ = r.ParseForm()
			if uiLocales != "" {
				r.Form.Set("ui_locales", uiLocales)
			}
			if acceptLanguage != "" {
				r.Header.Set("Accept-Language", acceptLanguage)
			}
			actual := PreferredLanguageTags(r)
			So(actual, ShouldResemble, expected)
		}

		// No ui_locales or Accept-Language
		test("", "", nil)

		// Accept-Language
		test("", "zh-Hant-HK; q=0.5, en", []string{"en", "zh-Hant-HK"})

		// ui_locales
		test("ja-JP zh-Hant-TW", "zh-Hant-HK; q=0.5, en", []string{"ja-JP", "zh-Hant-TW"})
	})
}
