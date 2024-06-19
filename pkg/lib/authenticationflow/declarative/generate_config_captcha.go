package declarative

import "github.com/authgear/authgear-server/pkg/lib/config"

func hasCaptcha(cfg *config.AppConfig) bool {
	return cfg.Captcha != nil && cfg.Captcha.Enabled && len(cfg.Captcha.Providers) > 0
}

func getBoolPtr(b bool) *bool {
	return &b
}

func generateFlowCaptcha(cfg *config.AppConfig) *config.AuthenticationFlowObjectCaptchaConfig {
	if cfg.Captcha == nil {
		return nil
	}
	if cfg.Captcha.Enabled && len(cfg.Captcha.Providers) > 0 {
		return &config.AuthenticationFlowObjectCaptchaConfig{
			Enabled: getBoolPtr(true),
		}
	}
	return nil
}
