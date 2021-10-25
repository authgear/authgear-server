package viewmodels

import (
	"encoding/json"
	htmltemplate "html/template"
	"net/http"
	"net/url"

	"github.com/gorilla/csrf"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/clientid"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/geoip"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/wechat"
)

type TranslationService interface {
	HasKey(key string) (bool, error)
	RenderText(key string, args interface{}) (string, error)
}

// BaseViewModel contains data that are common to all pages.
type BaseViewModel struct {
	CSRFField                   htmltemplate.HTML
	Translations                TranslationService
	StaticAssetURL              func(id string) (url string)
	DarkThemeEnabled            bool
	WatermarkEnabled            bool
	AllowedPhoneCountryCodeJSON string
	PinnedPhoneCountryCodeJSON  string
	GeoIPCountryCode            string
	FormJSON                    string
	// ClientURI is the home page of the client.
	ClientURI             string
	ClientName            string
	SliceContains         func([]interface{}, interface{}) bool
	MakeURL               func(path string, pairs ...string) string
	MakeCurrentStepURL    func(pairs ...string) string
	RawError              *apierrors.APIError
	Error                 interface{}
	SessionStepURLs       []string
	ForgotPasswordEnabled bool
	PublicSignupDisabled  bool
	PageLoadedAt          int
	IsNativePlatform      bool
	FlashMessageType      string
	ResolvedLanguageTag   string
	// IsSupportedMobilePlatform is true when the user agent is iOS or Android.
	IsSupportedMobilePlatform bool
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

func (m *BaseViewModel) SetFormJSON(form url.Values) {
	// Do not restore CSRF token.
	delete(form, webapp.CSRFFieldName)
	simpleMap := make(map[string]string)
	for key := range form {
		simpleMap[key] = form.Get(key)
	}
	b, err := json.Marshal(simpleMap)
	if err != nil {
		panic(err)
	}
	m.FormJSON = string(b)
}

type StaticAssetResolver interface {
	StaticAssetURL(id string) (url string, err error)
}

type ErrorCookie interface {
	GetError(r *http.Request) (*webapp.ErrorState, bool)
	ResetError() *http.Cookie
}

type FlashMessage interface {
	Pop(r *http.Request, rw http.ResponseWriter) string
}

type BaseViewModeler struct {
	TrustProxy            config.TrustProxy
	OAuth                 *config.OAuthConfig
	AuthUI                *config.UIConfig
	AuthUIFeatureConfig   *config.UIFeatureConfig
	StaticAssets          StaticAssetResolver
	ForgotPassword        *config.ForgotPasswordConfig
	Authentication        *config.AuthenticationConfig
	ErrorCookie           ErrorCookie
	Translations          TranslationService
	Clock                 clock.Clock
	FlashMessage          FlashMessage
	DefaultLanguageTag    template.DefaultLanguageTag
	SupportedLanguageTags template.SupportedLanguageTags
}

func (m *BaseViewModeler) ViewModel(r *http.Request, rw http.ResponseWriter) BaseViewModel {
	now := m.Clock.NowUTC().Unix()
	clientID := clientid.GetClientID(r.Context())
	client, _ := m.OAuth.GetClient(clientID)
	clientURI := webapp.ResolveClientURI(client, m.AuthUI)
	clientName := ""
	if client != nil {
		clientName = client.Name
	}

	preferredLanguageTags := intl.GetPreferredLanguageTags(r.Context())
	_, resolvedLanguageTagTag := intl.Resolve(preferredLanguageTags, string(m.DefaultLanguageTag), []string(m.SupportedLanguageTags))
	resolvedLanguageTag := resolvedLanguageTagTag.String()

	allowedPhoneCountryCodeJSON, err := json.Marshal(m.AuthUI.PhoneInput.AllowList)
	if err != nil {
		panic(err)
	}
	pinnedPhoneCountryCodeJSON, err := json.Marshal(m.AuthUI.PhoneInput.PinnedList)
	if err != nil {
		panic(err)
	}

	geoipCountryCode := ""
	if !m.AuthUI.PhoneInput.PreselectByIPDisabled {
		requestIP := httputil.GetIP(r, bool(m.TrustProxy))
		geoipInfo, ok := geoip.DefaultDatabase.IPString(requestIP)
		if ok {
			geoipCountryCode = geoipInfo.CountryCode
		}
	}

	model := BaseViewModel{
		CSRFField:    csrf.TemplateField(r),
		Translations: m.Translations,
		// This function has to return 1-value only.
		// Otherwise it cannot be used in template variable declartion.
		// What I mean here is that
		// {{ $a, $b := call $.StaticAssetURL "foobar" }}
		// is NOT supported at all.
		StaticAssetURL: func(id string) (url string) {
			url, _ = m.StaticAssets.StaticAssetURL(id)
			return
		},
		DarkThemeEnabled: !m.AuthUI.DarkThemeDisabled,
		WatermarkEnabled: m.AuthUIFeatureConfig.WhiteLabeling.Disabled ||
			!m.AuthUI.WatermarkDisabled,
		AllowedPhoneCountryCodeJSON: string(allowedPhoneCountryCodeJSON),
		PinnedPhoneCountryCodeJSON:  string(pinnedPhoneCountryCodeJSON),
		GeoIPCountryCode:            geoipCountryCode,
		ClientURI:                   clientURI,
		ClientName:                  clientName,
		SliceContains:               sliceContains,
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
		PageLoadedAt:          int(now),
		FlashMessageType:      m.FlashMessage.Pop(r, rw),
		ResolvedLanguageTag:   resolvedLanguageTag,
	}

	if errorState, ok := m.ErrorCookie.GetError(r); ok {
		model.SetFormJSON(errorState.Form)
		model.SetError(errorState.Error)
		httputil.UpdateCookie(rw, m.ErrorCookie.ResetError())
	}

	if s := webapp.GetSession(r.Context()); s != nil {
		for _, step := range s.Steps {
			if path := step.Kind.Path(); path == "" {
				continue
			}
			model.SessionStepURLs = append(model.SessionStepURLs, step.URL().String())
		}
	}

	platform := r.Form.Get("x_platform")
	if platform == "" {
		// if platform is not provided from the query and form
		// check it from cookie
		platform = wechat.GetPlatform(r.Context())
	}
	model.IsNativePlatform = (platform == "ios" ||
		platform == "android")

	ua := apimodel.ParseUserAgent(r.UserAgent())
	model.IsSupportedMobilePlatform = ua.OS == "iOS" || ua.OS == "Android"

	return model
}
