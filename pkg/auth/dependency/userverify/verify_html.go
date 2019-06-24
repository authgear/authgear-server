package userverify

import (
	"net/url"

	authTemplate "github.com/skygeario/skygear-server/pkg/auth/template"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

type VerifyHTMLProvider struct {
	configMap      map[string]verifyHTMLConfig
	errorRedirect  *url.URL
	templateEngine *template.Engine
}

type verifyHTMLConfig struct {
	successRedirect *url.URL
	errorRedirect   *url.URL
}

func newVerifyHTMLConfig(c config.UserVerificationKeyConfiguration) verifyHTMLConfig {
	var successRedirect *url.URL
	var errorRedirect *url.URL
	var err error

	if c.SuccessRedirect != "" {
		if successRedirect, err = url.Parse(c.SuccessRedirect); err != nil {
			panic("invalid Forgot password success redirect URL")
		}
	}

	if c.ErrorRedirect != "" {
		if errorRedirect, err = url.Parse(c.ErrorRedirect); err != nil {
			panic("invalid Forgot password error redirect URL")
		}
	}

	return verifyHTMLConfig{
		successRedirect: successRedirect,
		errorRedirect:   errorRedirect,
	}
}

func NewVerifyHTMLProvider(c config.UserVerificationConfiguration, templateEngine *template.Engine) *VerifyHTMLProvider {
	configMap := map[string]verifyHTMLConfig{}
	for _, keyConfig := range c.Keys {
		configMap[keyConfig.Key] = newVerifyHTMLConfig(keyConfig)
	}

	return &VerifyHTMLProvider{
		configMap:      configMap,
		templateEngine: templateEngine,
	}
}

func (v *VerifyHTMLProvider) SuccessHTML(key string, context map[string]interface{}) (string, error) {
	return v.templateEngine.ParseTextTemplate(
		authTemplate.VerifySuccessHTMLTemplateNameForKey(key),
		context,
		template.ParseOption{Required: true, FallbackTemplateName: authTemplate.TemplateNameVerifySuccessHTML},
	)
}

func (v *VerifyHTMLProvider) ErrorHTML(key string, context map[string]interface{}) (string, error) {
	if key != "" {
		return v.templateEngine.ParseTextTemplate(
			authTemplate.VerifyErrorHTMLTemplateNameForKey(key),
			context,
			template.ParseOption{Required: true, FallbackTemplateName: authTemplate.TemplateNameVerifyErrorHTML},
		)
	}

	return v.templateEngine.ParseTextTemplate(
		authTemplate.TemplateNameVerifyErrorHTML,
		context,
		template.ParseOption{Required: true},
	)
}

func (v *VerifyHTMLProvider) SuccessRedirect(key string, context map[string]interface{}) *url.URL {
	successRedirect := v.configMap[key].successRedirect
	if successRedirect == nil {
		return nil
	}

	output := *successRedirect
	v.setURLQueryFromMap(&output, context)
	return &output
}

func (v *VerifyHTMLProvider) ErrorRedirect(key string, context map[string]interface{}) (output *url.URL) {
	var errorRedirect *url.URL
	defer func() {
		if errorRedirect != nil {
			outputURL := *errorRedirect
			v.setURLQueryFromMap(&outputURL, context)
			output = &outputURL
		} else {
			output = nil
		}

		return
	}()

	if key != "" {
		errorRedirect = v.configMap[key].errorRedirect
		if errorRedirect != nil {
			return
		}
	}

	errorRedirect = v.errorRedirect
	return
}

func (v *VerifyHTMLProvider) setURLQueryFromMap(u *url.URL, values map[string]interface{}) {
	queryValues := url.Values{}
	for key, value := range values {
		if str, ok := value.(string); ok {
			queryValues.Set(key, str)
		}
	}

	u.RawQuery = queryValues.Encode()
}
