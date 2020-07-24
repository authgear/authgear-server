package viewmodels

import (
	"encoding/json"
	htmltemplate "html/template"
	"net/http"
	"net/url"

	"github.com/gorilla/csrf"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
	"github.com/authgear/authgear-server/pkg/core/intl"
	"github.com/authgear/authgear-server/pkg/core/skyerr"
)

// BaseViewModel contains data that are common to all pages.
type BaseViewModel struct {
	CSRFField               htmltemplate.HTML
	CSS                     htmltemplate.CSS
	AppName                 string
	LogoURI                 string
	CountryCallingCodes     []string
	StaticAssetURLPrefix    string
	SliceContains           func([]interface{}, interface{}) bool
	MakeURLWithQuery        func(pairs ...string) string
	MakeURLWithPathWithoutX func(path string) string
	Error                   interface{}
	ForgotPasswordEnabled   bool
}

type BaseViewModeler struct {
	ServerConfig   *config.ServerConfig
	AuthUI         *config.UIConfig
	Localization   *config.LocalizationConfig
	ForgotPassword *config.ForgotPasswordConfig
	Metadata       config.AppMetadata
}

func (m *BaseViewModeler) ViewModel(r *http.Request, anyError interface{}) BaseViewModel {
	preferredLanguageTags := intl.GetPreferredLanguageTags(r.Context())

	model := BaseViewModel{
		CSRFField: csrf.TemplateField(r),
		// NOTE(authui): We assume the CSS provided by the developer is trusted.
		CSS:                  htmltemplate.CSS(m.AuthUI.CustomCSS),
		AppName:              intl.LocalizeJSONObject(preferredLanguageTags, intl.Fallback(m.Localization.FallbackLanguage), m.Metadata, "app_name"),
		LogoURI:              intl.LocalizeJSONObject(preferredLanguageTags, intl.Fallback(m.Localization.FallbackLanguage), m.Metadata, "logo_uri"),
		CountryCallingCodes:  m.AuthUI.CountryCallingCode.Values,
		StaticAssetURLPrefix: m.ServerConfig.StaticAsset.URLPrefix,
		SliceContains:        sliceContains,
		MakeURLWithQuery: func(pairs ...string) string {
			q := url.Values{}
			for i := 0; i < len(pairs); i += 2 {
				q.Set(pairs[i], pairs[i+1])
			}
			return webapp.MakeURLWithQuery(r.URL, q)
		},
		MakeURLWithPathWithoutX: func(path string) string {
			return webapp.MakeURLWithPathWithoutX(r.URL, path)
		},
		ForgotPasswordEnabled: *m.ForgotPassword.Enabled,
	}

	if apiError := asAPIError(anyError); apiError != nil {
		b, err := json.Marshal(struct {
			Error *skyerr.APIError `json:"error"`
		}{apiError})
		if err != nil {
			panic(err)
		}
		var eJSON map[string]interface{}
		err = json.Unmarshal(b, &eJSON)
		if err != nil {
			panic(err)
		}
		model.Error = eJSON["error"]
	}

	return model
}
