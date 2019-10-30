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
	err             error
	successRedirect *url.URL
	errorRedirect   *url.URL
}

func newVerifyHTMLConfig(c config.UserVerificationKeyConfiguration) verifyHTMLConfig {
	var successRedirect *url.URL
	var errorRedirect *url.URL
	var e, err error

	if c.SuccessRedirect != "" {
		if successRedirect, e = url.Parse(c.SuccessRedirect); e != nil {
			err = e
		}
	}

	if c.ErrorRedirect != "" {
		if errorRedirect, e = url.Parse(c.ErrorRedirect); e != nil {
			err = e
		}
	}

	return verifyHTMLConfig{
		successRedirect: successRedirect,
		errorRedirect:   errorRedirect,
		err:             err,
	}
}

func NewVerifyHTMLProvider(c config.UserVerificationConfiguration, templateEngine *template.Engine) *VerifyHTMLProvider {
	configMap := map[string]verifyHTMLConfig{}
	for _, config := range c.LoginIDKeys {
		configMap[config.Key] = newVerifyHTMLConfig(config)
	}

	return &VerifyHTMLProvider{
		configMap:      configMap,
		templateEngine: templateEngine,
	}
}

func (v *VerifyHTMLProvider) SuccessHTML(key string, context map[string]interface{}) (string, error) {
	return v.templateEngine.RenderHTMLTemplate(
		// TODO(template): support type + key
		authTemplate.TemplateItemTypeUserVerificationSuccessHTML,
		context,
		template.RenderOptions{Required: true},
	)
}

func (v *VerifyHTMLProvider) ErrorHTML(key string, context map[string]interface{}) (string, error) {
	if key != "" {
		return v.templateEngine.RenderHTMLTemplate(
			// TODO(template): support type + key
			authTemplate.TemplateItemTypeUserVerificationErrorHTML,
			context,
			template.RenderOptions{Required: true},
		)
	}

	return v.templateEngine.RenderHTMLTemplate(
		authTemplate.TemplateItemTypeUserVerificationErrorHTML,
		context,
		template.RenderOptions{Required: true},
	)
}

func (v *VerifyHTMLProvider) SuccessRedirect(key string, context map[string]interface{}) *url.URL {
	config := v.configMap[key]
	if config.err != nil {
		panic(config.err)
	}

	successRedirect := config.successRedirect
	if successRedirect == nil {
		return nil
	}

	output := *successRedirect
	v.setURLQueryFromMap(&output, context)
	return &output
}

func (v *VerifyHTMLProvider) ErrorRedirect(key string, context map[string]interface{}) (output *url.URL) {
	config := v.configMap[key]
	if config.err != nil {
		panic(config.err)
	}

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
		errorRedirect = config.errorRedirect
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
