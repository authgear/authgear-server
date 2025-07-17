package captcha

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/captcha"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type Provider struct {
	RemoteIP         httputil.RemoteIP
	Config           *config.CaptchaConfig
	CloudflareClient *captcha.CloudflareClient
}

func (p *Provider) VerifyToken(ctx context.Context, token string) error {
	if p.Config.Deprecated_Provider == nil {
		return fmt.Errorf("captcha provider not configured")
	}
	switch *p.Config.Deprecated_Provider {
	case config.Deprecated_CaptchaProviderCloudflare:
		return p.verifyTokenByCloudflare(ctx, token)
	}
	return fmt.Errorf("unknown captcha provider")
}

func (p *Provider) verifyTokenByCloudflare(ctx context.Context, token string) error {
	if p.CloudflareClient == nil {
		return fmt.Errorf("missing cloudflare credential")
	}
	result, err := p.CloudflareClient.Verify(ctx, token, string(p.RemoteIP))
	if err != nil {
		return err
	}
	if !result.Success {
		return ErrVerificationFailed
	}
	return nil
}
