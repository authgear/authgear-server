package webapp

import (
	"encoding/json"
	htmlTemplate "html/template"
	"net/http"
	"strconv"

	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/phone"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

type RenderProviderImpl struct {
	StaticAssetURLPrefix string
	AuthConfiguration    *config.AuthConfiguration
	AuthUIConfiguration  *config.AuthUIConfiguration
	OAuthProviders       []config.OAuthProviderConfiguration
	TemplateEngine       *template.Engine
}

func (p *RenderProviderImpl) WritePage(w http.ResponseWriter, r *http.Request, templateType config.TemplateItemType, inputErr error) {
	data := FormToJSON(r.Form)
	accessKey := coreAuth.GetAccessKey(r.Context())
	data["client_name"] = accessKey.Client["client_name"]
	data["logo_uri"] = accessKey.Client["logo_uri"]

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

	for _, keyConfig := range p.AuthConfiguration.LoginIDKeys {
		if string(keyConfig.Type) == "phone" {
			data["x_login_id_input_type_has_phone"] = true
		} else {
			data["x_login_id_input_type_has_text"] = true
		}
	}

	var loginIDKeys []map[string]interface{}
	for _, loginIDKey := range p.AuthConfiguration.LoginIDKeys {
		loginIDKeys = append(loginIDKeys, map[string]interface{}{
			"key":  loginIDKey.Key,
			"type": loginIDKey.Type,
		})
	}
	data["x_login_id_keys"] = loginIDKeys

	// Populate inputErr into data
	if inputErr != nil {
		b, err := json.Marshal(struct {
			Error *skyerr.APIError `json:"error"`
		}{skyerr.AsAPIError(inputErr)})
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

	out, err := p.TemplateEngine.RenderTemplate(templateType, data, template.ResolveOptions{}, func(v *template.Validator) {
		v.AllowRangeNode = true
		v.AllowTemplateNode = true
		v.MaxDepth = 10
	})
	if err != nil {
		panic(err)
	}
	body := []byte(out)
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	if apiError := skyerr.AsAPIError(inputErr); apiError != nil {
		w.WriteHeader(apiError.Code)
	}
	w.Write(body)
}
