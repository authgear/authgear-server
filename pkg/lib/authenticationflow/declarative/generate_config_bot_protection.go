package declarative

import "github.com/authgear/authgear-server/pkg/lib/config"

func isBotProtectionEnabled(cfg *config.AppConfig) bool {
	return cfg.BotProtection != nil && cfg.BotProtection.Enabled && cfg.BotProtection.Provider != nil
}
func getBotProtectionRequirementsSignupOrLogin(cfg *config.AppConfig) (botProtection *config.AuthenticationFlowBotProtection) {
	if !isBotProtectionEnabled(cfg) {
		return nil
	}
	if cfg.BotProtection.Requirements == nil || cfg.BotProtection.Requirements.SignupOrLogin == nil {
		return nil
	}
	return &config.AuthenticationFlowBotProtection{
		Mode: cfg.BotProtection.Requirements.SignupOrLogin.Mode,
		Provider: &config.AuthenticationFlowBotProtectionProvider{
			Type: cfg.BotProtection.Provider.Type,
		},
	}
}

func getBotProtectionRequirementsAccountRecovery(cfg *config.AppConfig) (botProtection *config.AuthenticationFlowBotProtection) {
	if !isBotProtectionEnabled(cfg) {
		return nil
	}
	if cfg.BotProtection.Requirements == nil || cfg.BotProtection.Requirements.AccountRecovery == nil {
		return nil
	}
	return &config.AuthenticationFlowBotProtection{
		Mode: cfg.BotProtection.Requirements.AccountRecovery.Mode,
		Provider: &config.AuthenticationFlowBotProtectionProvider{
			Type: cfg.BotProtection.Provider.Type,
		},
	}
}
func getBotProtectionRequirementsPassword(cfg *config.AppConfig) (botProtection *config.AuthenticationFlowBotProtection) {
	if !isBotProtectionEnabled(cfg) {
		return nil
	}
	if cfg.BotProtection.Requirements == nil || cfg.BotProtection.Requirements.Password == nil {
		return nil
	}
	return &config.AuthenticationFlowBotProtection{
		Mode: cfg.BotProtection.Requirements.Password.Mode,
		Provider: &config.AuthenticationFlowBotProtectionProvider{
			Type: cfg.BotProtection.Provider.Type,
		},
	}
}
func getBotProtectionRequirementsOOBOTPEmail(cfg *config.AppConfig) (botProtection *config.AuthenticationFlowBotProtection) {
	if !isBotProtectionEnabled(cfg) {
		return nil
	}
	if cfg.BotProtection.Requirements == nil || cfg.BotProtection.Requirements.OOBOTPEmail == nil {
		return nil
	}
	return &config.AuthenticationFlowBotProtection{
		Mode: cfg.BotProtection.Requirements.OOBOTPEmail.Mode,
		Provider: &config.AuthenticationFlowBotProtectionProvider{
			Type: cfg.BotProtection.Provider.Type,
		},
	}
}
func getBotProtectionRequirementsOOBOTPSMS(cfg *config.AppConfig) (botProtection *config.AuthenticationFlowBotProtection) {
	if !isBotProtectionEnabled(cfg) {
		return nil
	}
	if cfg.BotProtection.Requirements == nil || cfg.BotProtection.Requirements.OOBOTPSMS == nil {
		return nil
	}
	return &config.AuthenticationFlowBotProtection{
		Mode: cfg.BotProtection.Requirements.OOBOTPSMS.Mode,
		Provider: &config.AuthenticationFlowBotProtectionProvider{
			Type: cfg.BotProtection.Provider.Type,
		},
	}
}
