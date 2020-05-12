package webapp

import (
	"encoding/json"
	htmlTemplate "html/template"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/csrf"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/oob"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/intl"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

type IdentityProvider interface {
	ListCandidates(userID string) ([]identity.Candidate, error)
}

type RenderProviderImpl struct {
	StaticAssetURLPrefix        string
	IdentityConfiguration       *config.IdentityConfiguration
	AuthenticationConfiguration *config.AuthenticationConfiguration
	AuthUIConfiguration         *config.AuthUIConfiguration
	LocalizationConfiguration   *config.LocalizationConfiguration
	TemplateEngine              *template.Engine
	PasswordChecker             *audit.PasswordChecker
	Identity                    IdentityProvider
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

func (p *RenderProviderImpl) PrepareRequestData(r *http.Request, data map[string]interface{}) {
	data["MakeURLWithQuery"] = func(pairs ...string) string {
		q := url.Values{}
		for i := 0; i < len(pairs); i += 2 {
			q.Set(pairs[i], pairs[i+1])
		}
		return MakeURLWithQuery(r.URL, q)
	}

	data["MakeURLWithPathWithoutX"] = func(path string) string {
		return MakeURLWithPathWithoutX(r.URL, path)
	}

	accessKey := coreAuth.GetAccessKey(r.Context())
	clientMetadata := accessKey.Client
	preferredLanguageTags := intl.GetPreferredLanguageTags(r.Context())
	data["client_name"] = intl.LocalizeJSONObject(preferredLanguageTags, intl.Fallback(p.LocalizationConfiguration.FallbackLanguage), clientMetadata, "client_name")
	data["logo_uri"] = intl.LocalizeJSONObject(preferredLanguageTags, intl.Fallback(p.LocalizationConfiguration.FallbackLanguage), clientMetadata, "logo_uri")

	data[csrf.TemplateTag] = csrf.TemplateField(r)

}

func (p *RenderProviderImpl) PrepareStaticData(data map[string]interface{}) {
	data["x_oob_otp_code_length"] = oob.OOBCodeLength
	data["x_oob_otp_code_send_cooldown"] = oob.OOBCodeSendCooldownSeconds
	data["x_static_asset_url_prefix"] = p.StaticAssetURLPrefix
	data["x_calling_codes"] = p.AuthUIConfiguration.CountryCallingCode.Values

	// NOTE(authui): We assume the CSS provided by the developer is trusted.
	data["x_css"] = htmlTemplate.CSS(p.AuthUIConfiguration.CSS)
}

func (p *RenderProviderImpl) PrepareIdentityData(r *http.Request, data map[string]interface{}) (err error) {
	userID := ""
	if sess := auth.GetSession(r.Context()); sess != nil {
		userID = sess.AuthnAttrs().UserID
	}

	identityCandidates, err := p.Identity.ListCandidates(userID)
	if err != nil {
		return
	}

	for _, c := range identityCandidates {
		if c[identity.CandidateKeyType] == string(authn.IdentityTypeLoginID) {
			if c[identity.CandidateKeyLoginIDType] == "phone" {
				c["login_id_input_type"] = "phone"
				data["x_login_id_input_type_has_phone"] = true
			} else {
				c["login_id_input_type"] = "text"
				data["x_login_id_input_type_has_text"] = true
			}
		}
	}
	data["x_identity_candidates"] = identityCandidates

	return
}

func (p *RenderProviderImpl) PreparePasswordPolicyData(anyError interface{}, data map[string]interface{}) {
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
}

func (p *RenderProviderImpl) PrepareErrorData(anyError interface{}, data map[string]interface{}) {
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
}

func (p *RenderProviderImpl) WritePage(w http.ResponseWriter, r *http.Request, templateType config.TemplateItemType, anyError interface{}) {
	data := FormToJSON(r.Form)

	p.PrepareStaticData(data)
	err := p.PrepareIdentityData(r, data)
	if err != nil {
		panic(err)
	}
	p.PrepareRequestData(r, data)
	p.PreparePasswordPolicyData(anyError, data)
	p.PrepareErrorData(anyError, data)

	preferredLanguageTags := intl.GetPreferredLanguageTags(r.Context())
	out, err := p.TemplateEngine.WithValidatorOptions(
		template.AllowRangeNode(true),
		template.AllowTemplateNode(true),
		template.AllowDeclaration(true),
		template.MaxDepth(15),
	).WithPreferredLanguageTags(preferredLanguageTags).RenderTemplate(
		templateType,
		data,
		template.ResolveOptions{},
	)
	if err != nil {
		panic(err)
	}

	body := []byte(out)
	// It is very important to specify the encoding
	// because browsers assume ASCII if encoding is not specified.
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	if apiError := p.asAPIError(anyError); apiError != nil {
		w.WriteHeader(apiError.Code)
	}
	w.Write(body)
}
