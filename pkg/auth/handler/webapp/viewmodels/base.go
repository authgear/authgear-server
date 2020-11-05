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
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

// BaseViewModel contains data that are common to all pages.
type BaseViewModel struct {
	// RequestURL is the absolute request URL.
	RequestURL string
	// RequestURI is the request URI as appeared in the first line of the HTTP textual format.
	// That is, it is the path plus the optional query.
	RequestURI            string
	CSRFField             htmltemplate.HTML
	StaticAssetURL        func(id string) (url string, err error)
	CountryCallingCodes   []string
	SliceContains         func([]interface{}, interface{}) bool
	MakeURL               func(path string, pairs ...string) string
	MakeCurrentStepURL    func(pairs ...string) string
	RawError              *apierrors.APIError
	Error                 interface{}
	ForgotPasswordEnabled bool
	PublicSignupDisabled  bool
}

func (m *BaseViewModel) SetError(err error) {
	if apiError := asAPIError(err); apiError != nil {
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
		m.Error = eJSON["error"]
		m.RawError = apiError
	}
}

type StaticAssetResolver interface {
	StaticAssetURL(id string) (url string, err error)
}

type ErrorCookie interface {
	GetError(r *http.Request) (*apierrors.APIError, bool)
	ResetError() *http.Cookie
}

type BaseViewModeler struct {
	AuthUI         *config.UIConfig
	StaticAssets   StaticAssetResolver
	ForgotPassword *config.ForgotPasswordConfig
	Authentication *config.AuthenticationConfig
	ErrorCookie    ErrorCookie
}

func (m *BaseViewModeler) ViewModel(r *http.Request, rw http.ResponseWriter) BaseViewModel {
	requestURI := &url.URL{
		Path:     r.URL.Path,
		RawQuery: r.URL.RawQuery,
	}
	model := BaseViewModel{
		RequestURL:          r.URL.String(),
		RequestURI:          requestURI.String(),
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
		MakeCurrentStepURL: func(pairs ...string) string {
			u := r.URL
			inQuery := url.Values{}
			for i := 0; i < len(pairs); i += 2 {
				inQuery.Set(pairs[i], pairs[i+1])
			}
			step := r.Form.Get("x_step")
			if step != "" {
				inQuery.Set("x_step", step)
			}
			return webapp.MakeURL(u, u.Path, inQuery).String()
		},
		ForgotPasswordEnabled: *m.ForgotPassword.Enabled,
		PublicSignupDisabled:  m.Authentication.PublicSignupDisabled,
	}

	if apiError, ok := m.ErrorCookie.GetError(r); ok {
		model.SetError(apiError)
		httputil.UpdateCookie(rw, m.ErrorCookie.ResetError())
	}

	return model
}
