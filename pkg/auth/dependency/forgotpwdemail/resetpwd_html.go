package forgotpwdemail

import (
	"net/url"

	authTemplate "github.com/skygeario/skygear-server/pkg/auth/template"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

type ResetPasswordHTMLProvider struct {
	TemplateEngine *template.Engine

	successRedirect *url.URL
	errorRedirect   *url.URL
}

func NewResetPasswordHTMLProvider(c config.ForgotPasswordConfiguration, templateEngine *template.Engine) *ResetPasswordHTMLProvider {
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

	return &ResetPasswordHTMLProvider{
		TemplateEngine:  templateEngine,
		successRedirect: successRedirect,
		errorRedirect:   errorRedirect,
	}
}

func (r *ResetPasswordHTMLProvider) SuccessHTML(context map[string]interface{}) (string, error) {
	return r.TemplateEngine.ParseTextTemplate(
		authTemplate.TemplateNameResetPasswordErrorHTML,
		context,
		template.ParseOption{Required: true},
	)
}

func (r *ResetPasswordHTMLProvider) ErrorHTML(context map[string]interface{}) (string, error) {
	return r.TemplateEngine.ParseTextTemplate(
		authTemplate.TemplateNameResetPasswordSuccessHTML,
		context,
		template.ParseOption{Required: true},
	)
}

func (r *ResetPasswordHTMLProvider) FormHTML(context map[string]interface{}) (string, error) {
	return r.TemplateEngine.ParseTextTemplate(
		authTemplate.TemplateNameResetPasswordHTML,
		context,
		template.ParseOption{Required: true},
	)
}

func (r *ResetPasswordHTMLProvider) SuccessRedirect(context map[string]interface{}) *url.URL {
	if r.successRedirect == nil {
		return nil
	}

	output := *r.successRedirect
	r.setURLQueryFromMap(&output, context)
	return &output
}

func (r *ResetPasswordHTMLProvider) ErrorRedirect(context map[string]interface{}) *url.URL {
	if r.errorRedirect == nil {
		return nil
	}

	output := *r.errorRedirect
	r.setURLQueryFromMap(&output, context)
	return &output
}

func (r *ResetPasswordHTMLProvider) setURLQueryFromMap(u *url.URL, values map[string]interface{}) {
	queryValues := url.Values{}
	for key, value := range values {
		if str, ok := value.(string); ok {
			queryValues.Set(key, str)
		}
	}

	u.RawQuery = queryValues.Encode()
}
