package botprotection

import (
	"context"
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type ProviderLogger struct{ *log.Logger }

func NewProviderLogger(lf *log.Factory) ProviderLogger {
	return ProviderLogger{lf.New("botprotection")}
}

type EventService interface {
	DispatchEventImmediately(ctx context.Context, payload event.NonBlockingPayload) error
}

type Provider struct {
	RemoteIP          httputil.RemoteIP
	Config            *config.BotProtectionConfig
	Logger            ProviderLogger
	CloudflareClient  *CloudflareClient
	RecaptchaV2Client *RecaptchaV2Client
	Events            EventService
}

func (p *Provider) Verify(ctx context.Context, token string) (err error) {
	if !p.Config.Enabled || p.Config.Provider == nil {
		return fmt.Errorf("bot_protection provider not configured")
	}
	if token == "" {
		return fmt.Errorf("empty token for bot_protection")
	}

	switch p.Config.Provider.Type {
	case config.BotProtectionProviderTypeCloudflare:
		err = p.verifyTokenByCloudflare(ctx, token)
	case config.BotProtectionProviderTypeRecaptchaV2:
		err = p.verifyTokenByRecaptchaV2(ctx, token)
	default:
		panic(fmt.Errorf("unknown bot_protection provider"))
	}

	if errors.Is(err, ErrVerificationFailed) {
		dispatchErr := p.Events.DispatchEventImmediately(ctx, &nonblocking.BotProtectionVerificationFailedEventPayload{
			Token:        token,
			ProviderType: string(p.Config.Provider.Type),
		})
		err = errors.Join(err, dispatchErr)
	}

	return
}

func (p *Provider) verifyTokenByCloudflare(ctx context.Context, token string) error {
	if p.CloudflareClient == nil {
		return fmt.Errorf("missing cloudflare credential")
	}
	_, err := p.CloudflareClient.Verify(ctx, token, string(p.RemoteIP))
	if err != nil {
		return err
	}
	return nil
}

func (p *Provider) verifyTokenByRecaptchaV2(ctx context.Context, token string) error {
	if p.RecaptchaV2Client == nil {
		return fmt.Errorf("missing recaptchaV2 credentials")
	}

	_, err := p.RecaptchaV2Client.Verify(ctx, token, string(p.RemoteIP))
	if err != nil {
		return err
	}

	return nil
}
