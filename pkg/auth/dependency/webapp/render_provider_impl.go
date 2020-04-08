package webapp

import (
	"encoding/json"
	htmlTemplate "html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/csrf"
	"golang.org/x/text/language"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/phone"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

type RenderProviderImpl struct {
	StaticAssetURLPrefix string
	LoginIDConfiguration *config.LoginIDConfiguration
	AuthUIConfiguration  *config.AuthUIConfiguration
	OAuthProviders       []config.OAuthProviderConfiguration
	TemplateEngine       *template.Engine
	PasswordChecker      *audit.PasswordChecker
}

func (p *RenderProviderImpl) asAPIError(anyError interface{}) *skyerr.APIError {
	if apiError, ok := anyError.(*skyerr.APIError); ok {
		return apiError
	}
	if err, ok := anyError.(error); ok {
		return skyerr.AsAPIError(err)
	}
	return nil
}

func (p *RenderProviderImpl) WritePage(w http.ResponseWriter, r *http.Request, templateType config.TemplateItemType, anyError interface{}) {
	data := FormToJSON(r.Form)
	accessKey := coreAuth.GetAccessKey(r.Context())

	data["MakeURLWithQuery"] = func(pairs ...string) string {
		q := url.Values{}
		for i := 0; i < len(pairs); i += 2 {
			q.Set(pairs[i], pairs[i+1])
		}
		return MakeURLWithQuery(r.URL, q)
	}

	data["MakeURLWithPath"] = func(path string) string {
		return MakeURLWithPath(r.URL, path)
	}

	data["client_name"] = accessKey.Client["client_name"]
	data["logo_uri"] = accessKey.Client["logo_uri"]

	data[csrf.TemplateTag] = csrf.TemplateField(r)

	data["x_static_asset_url_prefix"] = p.StaticAssetURLPrefix

	var providers []map[string]interface{}
	for _, provider := range p.OAuthProviders {
		providers = append(providers, map[string]interface{}{
			"id":   provider.ID,
			"type": provider.Type,
		})
	}
	data["x_idp_providers"] = providers

	// NOTE(authui): We assume the CSS provided by the developer is trusted.
	data["x_css"] = htmlTemplate.CSS(p.AuthUIConfiguration.CSS)

	data["x_calling_codes"] = phone.CountryCallingCodes

	for _, keyConfig := range p.LoginIDConfiguration.Keys {
		if string(keyConfig.Type) == "phone" {
			data["x_login_id_input_type_has_phone"] = true
		} else {
			data["x_login_id_input_type_has_text"] = true
		}
	}

	var loginIDKeys []map[string]interface{}
	for _, loginIDKey := range p.LoginIDConfiguration.Keys {
		inputType := "text"
		if loginIDKey.Type == "phone" {
			inputType = "phone"
		}
		loginIDKeys = append(loginIDKeys, map[string]interface{}{
			"key":        loginIDKey.Key,
			"type":       loginIDKey.Type,
			"input_type": inputType,
		})
	}
	data["x_login_id_keys"] = loginIDKeys

	passwordPolicy := p.PasswordChecker.PasswordPolicy()
	if apiError := p.asAPIError(anyError); apiError != nil {
		if apiError.Reason == "PasswordPolicyViolated" {
			for i, policy := range passwordPolicy {
				if policy.Info == nil {
					policy.Info = map[string]interface{}{}
				}
				policy.Info["x_error_is_password_policy_violated"] = true
				for _, causei := range apiError.Info["causes"].([]interface{}) {
					if cause, ok := causei.(map[string]interface{}); ok {
						if kind, ok := cause["kind"].(string); ok {
							if kind == string(policy.Name) {
								policy.Info["x_is_violated"] = true
							}
						}
					}
				}
				passwordPolicy[i] = policy
			}
		}
	}
	passwordPolicyBytes, err := json.Marshal(passwordPolicy)
	if err != nil {
		panic(err)
	}
	var passwordPolicyJSON interface{}
	err = json.Unmarshal(passwordPolicyBytes, &passwordPolicyJSON)
	if err != nil {
		panic(err)
	}
	data["x_password_policies"] = passwordPolicyJSON

	// Populate inputErr into data
	if apiError := p.asAPIError(anyError); apiError != nil {
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
		data["x_error"] = eJSON["error"]
	}

	out, err := p.TemplateEngine.WithValidatorOptions(
		template.AllowRangeNode(true),
		template.AllowTemplateNode(true),
		template.MaxDepth(15),
	).WithPreferredLanguageTags(PreferredLanguageTags(r)).RenderTemplate(
		templateType,
		data,
		template.ResolveOptions{},
	)
	if err != nil {
		panic(err)
	}
	body := []byte(out)
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	if apiError := p.asAPIError(anyError); apiError != nil {
		w.WriteHeader(apiError.Code)
	}
	w.Write(body)
}

func PreferredLanguageTags(r *http.Request) (out []string) {
	acceptLanguage := r.Header.Get("Accept-Language")
	if uiLocales := r.Form.Get("ui_locales"); uiLocales != "" {
		acceptLanguage = strings.ReplaceAll(uiLocales, " ", ", ")
	}
	tags, _, err := language.ParseAcceptLanguage(acceptLanguage)
	if err != nil {
		return
	}
	for _, tag := range tags {
		out = append(out, tag.String())
	}
	return
}
