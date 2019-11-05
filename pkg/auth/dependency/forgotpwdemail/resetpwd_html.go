package forgotpwdemail

import (
	"net/url"
	"path"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

type ResetPasswordHTMLProvider struct {
	TemplateEngine *template.Engine

	successRedirect *url.URL
	errorRedirect   *url.URL

	urlPrefix *url.URL
	config    config.ForgotPasswordConfiguration

	err error
}

func NewResetPasswordHTMLProvider(urlPrefix *url.URL, c config.ForgotPasswordConfiguration, templateEngine *template.Engine) *ResetPasswordHTMLProvider {
	var successRedirect *url.URL
	var errorRedirect *url.URL
	var providerError error
	if c.SuccessRedirect != "" {
		u, err := url.Parse(c.SuccessRedirect)
		if err == nil {
			successRedirect = u
		} else {
			providerError = errors.Newf("invalid success redirect URL: %w", err)
		}
	}

	if c.ErrorRedirect != "" {
		u, err := url.Parse(c.ErrorRedirect)
		if err == nil {
			errorRedirect = u
		} else {
			providerError = errors.Newf("invalid error redirect URL: %w", err)
		}
	}

	return &ResetPasswordHTMLProvider{
		TemplateEngine:  templateEngine,
		successRedirect: successRedirect,
		errorRedirect:   errorRedirect,
		urlPrefix:       urlPrefix,
		config:          c,
		err:             providerError,
	}
}

func (r *ResetPasswordHTMLProvider) SuccessHTML(context map[string]interface{}) (string, error) {
	r.injectContext(context)
	return r.TemplateEngine.RenderHTMLTemplate(
		TemplateItemTypeForgotPasswordSuccessHTML,
		context,
		template.RenderOptions{Required: true},
	)
}

func (r *ResetPasswordHTMLProvider) ErrorHTML(context map[string]interface{}) (string, error) {
	r.injectContext(context)
	return r.TemplateEngine.RenderHTMLTemplate(
		TemplateItemTypeForgotPasswordErrorHTML,
		context,
		template.RenderOptions{Required: true},
	)
}

func (r *ResetPasswordHTMLProvider) FormHTML(context map[string]interface{}) (string, error) {
	r.injectContext(context)
	return r.TemplateEngine.RenderHTMLTemplate(
		TemplateItemTypeForgotPasswordResetHTML,
		context,
		template.RenderOptions{Required: true},
	)
}

func (r *ResetPasswordHTMLProvider) injectContext(context map[string]interface{}) {
	context["url_prefix"] = r.urlPrefix.String()
	u := *r.urlPrefix
	u.Path = path.Join(u.Path, "_auth/forgot_password/reset_password_form")
	context["action_url"] = u.String()
}

func (r *ResetPasswordHTMLProvider) SuccessRedirect(context map[string]interface{}) *url.URL {
	if r.err != nil {
		panic(r.err)
	}

	if r.successRedirect == nil {
		return nil
	}

	output := *r.successRedirect
	template.SetContextToURLQuery(&output, context)
	return &output
}

func (r *ResetPasswordHTMLProvider) ErrorRedirect(context map[string]interface{}) *url.URL {
	if r.err != nil {
		panic(r.err)
	}

	if r.errorRedirect == nil {
		return nil
	}

	output := *r.errorRedirect
	template.SetContextToURLQuery(&output, context)
	return &output
}
