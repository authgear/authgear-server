package botprotection

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type ProviderLogger struct{ *log.Logger }

func NewProviderLogger(lf *log.Factory) ProviderLogger {
	return ProviderLogger{lf.New("botprotection")}
}

type Provider struct{}

func (p *Provider) Verify(t config.BotProtectionProviderType, response string) error {
	// TODO: Implement bot protection provider
	return nil
}
