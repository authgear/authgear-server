package botprotection

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/botprotection"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type ProviderLogger struct{ *log.Logger }

func NewProviderLogger(lf *log.Factory) ProviderLogger {
	return ProviderLogger{lf.New("botprotection")}
}

type Provider struct {
	RemoteIP          httputil.RemoteIP
	Config            *config.BotProtectionConfig
	Logger            ProviderLogger
	CloudflareClient  *botprotection.CloudflareClient
	RecaptchaV2Client *botprotection.RecaptchaV2Client
}

func (p *Provider) Verify(token string) error {
	if !p.Config.Enabled || p.Config.Provider == nil {
		return fmt.Errorf("bot_protection provider not configured")
	}
	if token == "" {
		return fmt.Errorf("empty token for bot_protection")
	}

	switch p.Config.Provider.Type {
	case config.BotProtectionProviderTypeCloudflare:
		return p.verifyTokenByCloudflare(token)
	case config.BotProtectionProviderTypeRecaptchaV2:
		return p.verifyTokenByRecaptchaV2(token)
	}

	return fmt.Errorf("unknown bot_protection provider")
}

func (p *Provider) verifyTokenByCloudflare(token string) error {
	if p.CloudflareClient == nil {
		return fmt.Errorf("missing cloudflare credential")
	}
	result, err := p.CloudflareClient.Verify(token, string(p.RemoteIP))
	if err != nil {
		return err
	}
	if !*result.Success {
		p.Logger.WithField("cloudflare verification error-codes:", result.ErrorCodes)
		isServiceUnavailable := p.CloudflareClient.IsServiceUnavailable(result)
		if isServiceUnavailable {
			return ErrVerificationServiceUnavailable
		}
		return ErrVerificationFailed
	}
	return nil
}

func (p *Provider) verifyTokenByRecaptchaV2(token string) error {
	if p.RecaptchaV2Client == nil {
		return fmt.Errorf("missing recaptchaV2 credentials")
	}

	result, err := p.RecaptchaV2Client.Verify(token, string(p.RemoteIP))
	if err != nil {
		return err
	}
	if !*result.Success {
		p.Logger.WithField("cloudflare verification error-codes:", result.ErrorCodes)
		isServiceUnavailable := p.RecaptchaV2Client.IsServiceUnavailable(result)
		if isServiceUnavailable {
			return ErrVerificationServiceUnavailable
		}
		return ErrVerificationFailed
	}

	return nil
}
