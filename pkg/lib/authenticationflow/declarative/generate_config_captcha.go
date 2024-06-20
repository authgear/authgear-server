package declarative

import "github.com/authgear/authgear-server/pkg/lib/config"

func getCaptchaProviderConfig(cfg *config.AppConfig) (captcha *config.AuthenticationFlowCaptcha, ok bool) {
	if cfg.Captcha == nil || !cfg.Captcha.Enabled || len(cfg.Captcha.Providers) == 0 {
		return nil, false
	}
	return &config.AuthenticationFlowCaptcha{
		Mode: config.AuthenticationFlowCaptchaModeAlways, // default always in generated flow
		Provider: &config.AuthenticationFlowCaptchaProvider{
			Alias: cfg.Captcha.Providers[0].Alias, // default use first provider
		},
	}, true
}
