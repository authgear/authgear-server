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
)

// BaseViewModel contains data that are common to all pages.
type BaseViewModel struct {
	RequestURL            string
	CSRFField             htmltemplate.HTML
	StaticAssetURL        func(id string) (url string, err error)
	CountryCallingCodes   []string
	SliceContains         func([]interface{}, interface{}) bool
	MakeURL               func(path string, pairs ...string) string
	MakeURLState          func(path string, pairs ...string) string
	Error                 interface{}
	ForgotPasswordEnabled bool
	PublicSignupDisabled  bool
}

type StaticAssetResolver interface {
	StaticAssetURL(id string) (url string, err error)
}

type BaseViewModeler struct {
	AuthUI         *config.UIConfig
	StaticAssets   StaticAssetResolver
	ForgotPassword *config.ForgotPasswordConfig
	Authentication *config.AuthenticationConfig
}

func (m *BaseViewModeler) ViewModel(r *http.Request, anyError interface{}) BaseViewModel {
	model := BaseViewModel{
		RequestURL:          r.URL.String(),
		CSRFField:           csrf.TemplateField(r),
		StaticAssetURL:      m.StaticAssets.StaticAssetURL,
		CountryCallingCodes: m.AuthUI.CountryCallingCode.GetActiveCountryCodes(),
		SliceContains:       sliceContains,
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
