package viewmodels

import (
	"encoding/json"
	htmltemplate "html/template"
	"net/http"
	"net/url"

	"github.com/gorilla/csrf"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/translation"
)

// BaseViewModel contains data that are common to all pages.
type BaseViewModel struct {
	RequestURL            string
	CSRFField             htmltemplate.HTML
	CSS                   htmltemplate.CSS
	AppName               string
	LogoURI               string
	CountryCallingCodes   []string
	StaticAssetURLPrefix  string
	SliceContains         func([]interface{}, interface{}) bool
	MakeURL               func(path string, pairs ...string) string
	MakeURLState          func(path string, pairs ...string) string
	Error                 interface{}
	ForgotPasswordEnabled bool
	PublicSignupDisabled  bool
}

type TranslationService interface {
	AppMetadata() (*translation.AppMetadata, error)
}

type BaseViewModeler struct {
	StaticAssetURLPrefix config.StaticAssetURLPrefix
	AuthUI               *config.UIConfig
	Translation          TranslationService
	ForgotPassword       *config.ForgotPasswordConfig
	Authentication       *config.AuthenticationConfig
}

func (m *BaseViewModeler) ViewModel(r *http.Request, anyError interface{}) BaseViewModel {
	appMeta, err := m.Translation.AppMetadata()
	if err != nil {
		panic(err)
	}

	model := BaseViewModel{
		RequestURL: r.URL.String(),
		CSRFField:  csrf.TemplateField(r),
		// We assume the CSS provided by the developer is trusted.
		CSS:                  htmltemplate.CSS(m.AuthUI.CustomCSS),
		AppName:              appMeta.AppName,
		LogoURI:              appMeta.LogoURI,
		CountryCallingCodes:  m.AuthUI.CountryCallingCode.Values,
		StaticAssetURLPrefix: string(m.StaticAssetURLPrefix),
		SliceContains:        sliceContains,
		MakeURL: func(path string, pairs ...string) string {
			u := r.URL
			inQuery := url.Values{}
			for i := 0; i < len(pairs); i += 2 {
				inQuery.Set(pairs[i], pairs[i+1])
			}
			return webapp.MakeURL(u, path, inQuery).String()
		},
		MakeURLState: func(path string, pairs ...string) string {
			u := r.URL
			inQuery := url.Values{}
			for i := 0; i < len(pairs); i += 2 {
				inQuery.Set(pairs[i], pairs[i+1])
			}
			sid := r.Form.Get("x_sid")
			if sid != "" {
				inQuery.Set("x_sid", sid)
			}
			return webapp.MakeURL(u, path, inQuery).String()
		},
		ForgotPasswordEnabled: *m.ForgotPassword.Enabled,
		PublicSignupDisabled:  m.Authentication.PublicSignupDisabled,
	}

	if apiError := asAPIError(anyError); apiError != nil {
		b, err := json.Marshal(struct {
			Error *apierrors.APIError `json:"error"`
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
