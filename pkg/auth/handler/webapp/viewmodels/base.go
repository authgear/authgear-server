package viewmodels

import (
	"encoding/json"
	htmltemplate "html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/csrf"
	"golang.org/x/text/language"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/uiparam"
	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/geoip"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/slice"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/wechat"
)

type BaseLogger struct{ *log.Logger }

func NewBaseLogger(lf *log.Factory) BaseLogger {
	return BaseLogger{lf.New("webapp")}
}

type TranslationService interface {
	HasKey(key string) (bool, error)
	RenderText(key string, args interface{}) (string, error)
}

// BaseViewModel contains data that are common to all pages.
type BaseViewModel struct {
	RequestURI                  string
	ColorScheme                 string
	CSPNonce                    string
	CSRFField                   htmltemplate.HTML
	Translations                TranslationService
	HasAppSpecificAsset         func(id string) bool
	StaticAssetURL              func(id string) (url string)
	GeneratedStaticAssetURL     func(id string) (url string)
	DarkThemeEnabled            bool
	LightThemeEnabled           bool
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
	ErrorJSON             map[string]map[string]interface{}
	SessionStepURLs       []string
	ForgotPasswordEnabled bool
	PublicSignupDisabled  bool
	PageLoadedAt          int
	IsNativePlatform      bool
	FlashMessageType      string
	ResolvedLanguageTag   string
	ResolvedCLDRLocale    string
	HTMLDir               string
	// IsSupportedMobilePlatform is true when the user agent is iOS or Android.
	IsSupportedMobilePlatform   bool
	GoogleTagManagerContainerID string
	TutorialMessageType         string
	// HasThirdPartyApp indicates whether the project has third-party client
	HasThirdPartyClient               bool
	AuthUISentryDSN                   string
	AuthUIWindowMessageAllowedOrigins string

	FirstNonPasskeyPrimaryAuthenticatorType string
	// websocket is used in interaction only, we disable it in authflow
	WebsocketDisabled bool

	InlinePreview bool

	ShouldFocusInput bool
	LogUnknownError  func(err map[string]interface{}) string

	BotProtectionEnabled          bool
	BotProtectionProviderType     string
	BotProtectionProviderSiteKey  string
	ResolvedBotProtectionLanguage string
}

func (m *BaseViewModel) SetError(err error) {
	if apiError := asAPIError(err); apiError != nil {
		b, err := json.Marshal(struct {
			Error *apierrors.APIError `json:"error"`
		}{apiError})
		if err != nil {
			panic(err)
		}

		var eJSON map[string]map[string]interface{}
		var eJSONError map[string]interface{}
		err = json.Unmarshal(b, &eJSON)
		if err != nil {
			panic(err)
		}
		eJSONError = eJSON["error"]
		m.Error = eJSONError
		m.ErrorJSON = GetErrorJSON(apiError)
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

func (m *BaseViewModel) SetTutorial(name httputil.TutorialCookieName) {
	m.TutorialMessageType = string(name)
}

type StaticAssetResolver interface {
	HasAppSpecificAsset(id string) bool
	StaticAssetURL(id string) (url string, err error)
	GeneratedStaticAssetURL(id string) (url string, err error)
}

type ErrorService interface {
	PopError(w http.ResponseWriter, r *http.Request) (*webapp.ErrorState, bool)
}

type FlashMessage interface {
	Pop(r *http.Request, rw http.ResponseWriter) string
}

type WebappOAuthClientResolver interface {
	ResolveClient(clientID string) *config.OAuthClientConfig
}

type BaseViewModeler struct {
	TrustProxy                        config.TrustProxy
	OAuth                             *config.OAuthConfig
	AuthUI                            *config.UIConfig
	AuthUIFeatureConfig               *config.UIFeatureConfig
	StaticAssets                      StaticAssetResolver
	ForgotPassword                    *config.ForgotPasswordConfig
	Authentication                    *config.AuthenticationConfig
	GoogleTagManager                  *config.GoogleTagManagerConfig
	BotProtection                     *config.BotProtectionConfig
	ErrorService                      ErrorService
	Translations                      TranslationService
	Clock                             clock.Clock
	FlashMessage                      FlashMessage
	DefaultLanguageTag                template.DefaultLanguageTag
	SupportedLanguageTags             template.SupportedLanguageTags
	AuthUISentryDSN                   config.AuthUISentryDSN
	AuthUIWindowMessageAllowedOrigins config.AuthUIWindowMessageAllowedOrigins
	OAuthClientResolver               WebappOAuthClientResolver
	Logger                            BaseLogger
}

func (m *BaseViewModeler) ViewModelForAuthFlow(r *http.Request, rw http.ResponseWriter) BaseViewModel {
	vm := m.ViewModel(r, rw)
	vm.WebsocketDisabled = true
	return vm
}

func (m *BaseViewModeler) ViewModelForInlinePreviewAuthFlow(r *http.Request, rw http.ResponseWriter) BaseViewModel {
	vm := m.ViewModelForAuthFlow(r, rw)
	vm.InlinePreview = true
	return vm
}

// nolint: gocognit
func (m *BaseViewModeler) ViewModel(r *http.Request, rw http.ResponseWriter) BaseViewModel {
	now := m.Clock.NowUTC().Unix()
	uiParam := uiparam.GetUIParam(r.Context())
	clientID := uiParam.ClientID
	client := m.OAuthClientResolver.ResolveClient(clientID)
	clientURI := webapp.ResolveClientURI(client, m.AuthUI)
	clientName := ""
	if client != nil {
		clientName = client.Name
	}
	hasThirdPartyApp := false
	for _, c := range m.OAuth.Clients {
		if c.IsThirdParty() {
			hasThirdPartyApp = true
		}
	}

	cspNonce := web.GetCSPNonce(r.Context())

	preferredLanguageTags := intl.GetPreferredLanguageTags(r.Context())
	_, resolvedLanguageTagTag := intl.Resolve(preferredLanguageTags, string(m.DefaultLanguageTag), []string(m.SupportedLanguageTags))
	resolvedLanguageTag := resolvedLanguageTagTag.String()

	locale := intl.ResolveUnicodeCldr(resolvedLanguageTagTag, language.MustParse(string(m.DefaultLanguageTag)))

	htmlDir := intl.HTMLDir(resolvedLanguageTag)

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

	bpProviderType := m.BotProtection.GetProviderType()
	var bpLang string
	switch bpProviderType {
	case config.BotProtectionProviderTypeCloudflare:
		bpLang = intl.ResolveCloudflareTurnstile(resolvedLanguageTag)
	case config.BotProtectionProviderTypeRecaptchaV2:
		bpLang = intl.ResolveRecaptchaV2(resolvedLanguageTag)
	}

	model := BaseViewModel{
		ColorScheme:  webapp.GetColorScheme(r.Context()),
		RequestURI:   r.URL.RequestURI(),
		CSPNonce:     cspNonce,
		CSRFField:    csrf.TemplateField(r),
		Translations: m.Translations,
		HasAppSpecificAsset: func(id string) bool {
			return m.StaticAssets.HasAppSpecificAsset(id)
		},
		// This function has to return 1-value only.
		// Otherwise it cannot be used in template variable declartion.
		// What I mean here is that
		// {{ $a, $b := call $.StaticAssetURL "foobar" }}
		// is NOT supported at all.
		StaticAssetURL: func(id string) (url string) {
			url, _ = m.StaticAssets.StaticAssetURL(id)
			return
		},
		GeneratedStaticAssetURL: func(id string) (url string) {
			url, _ = m.StaticAssets.GeneratedStaticAssetURL(id)
			return
		},
		// If it is previewing, then we always allow both themes.
		// Imagine the case when the project is initially dark theme only,
		// In the design page, the preview is switched to light, then light must be enabled.
		DarkThemeEnabled:  !m.AuthUI.DarkThemeDisabled || webapp.IsInlinePreviewPageRequest(r),
		LightThemeEnabled: !m.AuthUI.LightThemeDisabled || webapp.IsInlinePreviewPageRequest(r),
		WatermarkEnabled: m.AuthUIFeatureConfig.WhiteLabeling.Disabled ||
			!m.AuthUI.WatermarkDisabled,
		AllowedPhoneCountryCodeJSON: string(allowedPhoneCountryCodeJSON),
		PinnedPhoneCountryCodeJSON:  string(pinnedPhoneCountryCodeJSON),
		GeoIPCountryCode:            geoipCountryCode,
		ClientURI:                   clientURI,
		ClientName:                  clientName,
		SliceContains:               SliceContains,
		MakeURL: func(path string, pairs ...string) string {
			u := r.URL
			outQuery := webapp.PreserveQuery(u.Query())
			for i := 0; i < len(pairs); i += 2 {
				key := pairs[i]
				val := pairs[i+1]
				if val != "" {
					outQuery.Set(key, val)
				} else {
					outQuery.Del(key)
				}
			}
			return webapp.MakeURL(u, path, outQuery).String()
		},
		ForgotPasswordEnabled:         *m.ForgotPassword.Enabled,
		PublicSignupDisabled:          m.Authentication.PublicSignupDisabled,
		PageLoadedAt:                  int(now),
		FlashMessageType:              m.FlashMessage.Pop(r, rw),
		ResolvedLanguageTag:           resolvedLanguageTag,
		ResolvedCLDRLocale:            locale,
		HTMLDir:                       htmlDir,
		GoogleTagManagerContainerID:   m.GoogleTagManager.ContainerID,
		BotProtectionEnabled:          m.BotProtection.IsEnabled(),
		BotProtectionProviderType:     string(bpProviderType),
		BotProtectionProviderSiteKey:  m.BotProtection.GetSiteKey(),
		ResolvedBotProtectionLanguage: bpLang,
		HasThirdPartyClient:           hasThirdPartyApp,
		AuthUISentryDSN:               string(m.AuthUISentryDSN),
		AuthUIWindowMessageAllowedOrigins: func() string {
			requestProto := httputil.GetProto(r, bool(m.TrustProxy))
			processedAllowedOrgins := slice.Map(m.AuthUIWindowMessageAllowedOrigins, func(origin string) string {
				return composeAuthUIWindowMessageAllowedOrigin(origin, requestProto)
			})
			return strings.Join(processedAllowedOrgins, ",")
		}(),
		LogUnknownError: func(err map[string]interface{}) string {
			if err != nil {
				m.Logger.WithFields(err).Errorf("unknown error: %v", err)
			}

			return ""
		},
	}

	if errorState, ok := m.ErrorService.PopError(rw, r); ok {
		model.SetFormJSON(errorState.Form)
		model.SetError(errorState.Error)
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

	for _, typ := range *m.Authentication.PrimaryAuthenticators {
		if typ != apimodel.AuthenticatorTypePasskey {
			t := typ
			model.FirstNonPasskeyPrimaryAuthenticatorType = string(t)
			break
		}
	}

	model.ShouldFocusInput = model.Error == nil && model.FlashMessageType == ""

	return model
}

// Assume allowed origin is either host or a real origin
func composeAuthUIWindowMessageAllowedOrigin(allowedOrigin string, proto string) string {
	if strings.HasPrefix(allowedOrigin, "http://") || strings.HasPrefix(allowedOrigin, "https://") {
		return allowedOrigin
	}
	u := url.URL{
		Scheme: proto,
		Host:   allowedOrigin,
	}
	return u.String()
}
