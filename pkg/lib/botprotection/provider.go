package botprotection

import (
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/config"
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
	CloudflareClient  *CloudflareClient
	RecaptchaV2Client *RecaptchaV2Client
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
	successResp, err := p.CloudflareClient.Verify(token, string(p.RemoteIP))
	if err != nil {
		p.Logger.WithField("cloudflare verification error:", err)
		return err
	}
	if successResp == nil {
		err = fmt.Errorf("cloudflare no error but no success response")
		return errors.Join(err, ErrVerificationFailed)
	}
	return nil
}

func (p *Provider) verifyTokenByRecaptchaV2(token string) error {
	if p.RecaptchaV2Client == nil {
		return fmt.Errorf("missing recaptchaV2 credentials")
	}

	successResp, err := p.RecaptchaV2Client.Verify(token, string(p.RemoteIP))
	if err != nil {
		p.Logger.WithField("recaptchav2 verification error:", err)
		return err
	}
	if successResp == nil {
		err = fmt.Errorf("recaptchav2 no error but no success response")
		return errors.Join(err, ErrVerificationFailed)
	}

	return nil
}
