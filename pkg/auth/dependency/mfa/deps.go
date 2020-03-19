package mfa

import (
	"net/http"

	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/mail"
	"github.com/skygeario/skygear-server/pkg/core/sms"
	"github.com/skygeario/skygear-server/pkg/core/template"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

type InsecureCookieConfig bool

func ProvideBearerTokenCookieConfiguration(
	r *http.Request,
	icc InsecureCookieConfig,
	c *config.TenantConfiguration,
) BearerTokenCookieConfiguration {
	return NewBearerTokenCookieConfiguration(r, bool(icc), *c.AppConfig.Session, *c.AppConfig.MFA)
}

func ProvideMFASender(
	tConfig *config.TenantConfiguration,
	smsClient sms.Client,
	mailSender mail.Sender,
	templateEngine *template.Engine,
) Sender {
	return NewSender(*tConfig, smsClient, mailSender, templateEngine)
}

func ProvideMFAProvider(store Store, config *config.TenantConfiguration, timeProvider time.Provider, sender Sender) Provider {
	return NewProvider(store, config.AppConfig.MFA, timeProvider, sender)
}

var DependencySet = wire.NewSet(
	ProvideBearerTokenCookieConfiguration,
	ProvideMFASender,
	ProvideMFAProvider,
)
