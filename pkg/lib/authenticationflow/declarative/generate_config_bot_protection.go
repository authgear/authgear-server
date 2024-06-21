package declarative

import "github.com/authgear/authgear-server/pkg/lib/config"

func getBotProtectionProviderConfig(cfg *config.AppConfig) (botProtection *config.AuthenticationFlowBotProtection, ok bool) {
	if cfg.BotProtection == nil || cfg.BotProtection.Enabled == nil || !*cfg.BotProtection.Enabled || cfg.BotProtection.Provider == nil {
		return nil, false
	}
	return &config.AuthenticationFlowBotProtection{
		Mode: config.AuthenticationFlowBotProtectionModeAlways, // default always in generated flow
		Provider: &config.AuthenticationFlowBotProtectionProvider{
			Type: cfg.BotProtection.Provider.Type,
		},
	}, true
}
