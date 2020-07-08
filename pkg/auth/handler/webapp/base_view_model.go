package webapp

import (
	"encoding/json"
	"fmt"
	htmltemplate "html/template"
	"net/http"
	"net/url"
	"reflect"

	"github.com/gorilla/csrf"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
	"github.com/authgear/authgear-server/pkg/core/intl"
	"github.com/authgear/authgear-server/pkg/core/skyerr"
)

func sliceContains(slice []interface{}, value interface{}) bool {
	for _, v := range slice {
		if reflect.DeepEqual(v, value) {
			return true
		}
	}
	return false
}

// Embed embeds the given struct s into data.
func Embed(data map[string]interface{}, s interface{}) {
	v := reflect.ValueOf(s)
	typ := v.Type()
	if typ.Kind() != reflect.Struct {
		panic(fmt.Errorf("webapp: expected struct but was %T", s))
	}
	numField := typ.NumField()
	for i := 0; i < numField; i++ {
		structField := typ.Field(i)
		data[structField.Name] = v.Field(i).Interface()
	}
}

func EmbedForm(data map[string]interface{}, form url.Values) {
	for name := range form {
		data[name] = form.Get(name)
	}
}

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
}

type BaseViewModeler struct {
	ServerConfig *config.ServerConfig
	AuthUI       *config.UIConfig
	Localization *config.LocalizationConfig
	Metadata     config.AppMetadata
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
