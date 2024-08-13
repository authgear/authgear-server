package declarative

import "github.com/authgear/authgear-server/pkg/lib/config"

func isBotProtectionEnabled(cfg *config.AppConfig) bool {
	return cfg.BotProtection != nil && cfg.BotProtection.Enabled && cfg.BotProtection.Provider != nil
}
func getBotProtectionRequirementsSignupOrLogin(cfg *config.AppConfig) (botProtection *config.AuthenticationFlowBotProtection, exist bool) {
	if !isBotProtectionEnabled(cfg) {
		return nil, false
	}
	if cfg.BotProtection.Requirements == nil || cfg.BotProtection.Requirements.SignupOrLogin == nil {
		return nil, false
	}
	return &config.AuthenticationFlowBotProtection{
		Mode: cfg.BotProtection.Requirements.SignupOrLogin.Mode,
	}, true
}

func getBotProtectionRequirementsAccountRecovery(cfg *config.AppConfig) (botProtection *config.AuthenticationFlowBotProtection, exist bool) {
	if !isBotProtectionEnabled(cfg) {
		return nil, false
	}
	if cfg.BotProtection.Requirements == nil || cfg.BotProtection.Requirements.AccountRecovery == nil {
		return nil, false
	}
	return &config.AuthenticationFlowBotProtection{
		Mode: cfg.BotProtection.Requirements.AccountRecovery.Mode,
	}, true
}
func getBotProtectionRequirementsPassword(cfg *config.AppConfig) (botProtection *config.AuthenticationFlowBotProtection, exist bool) {
	if !isBotProtectionEnabled(cfg) {
		return nil, false
	}
	if cfg.BotProtection.Requirements == nil || cfg.BotProtection.Requirements.Password == nil {
		return nil, false
	}
	return &config.AuthenticationFlowBotProtection{
		Mode: cfg.BotProtection.Requirements.Password.Mode,
	}, true
}
func getBotProtectionRequirementsOOBOTPEmail(cfg *config.AppConfig) (botProtection *config.AuthenticationFlowBotProtection, exist bool) {
	if !isBotProtectionEnabled(cfg) {
		return nil, false
	}
	if cfg.BotProtection.Requirements == nil || cfg.BotProtection.Requirements.OOBOTPEmail == nil {
		return nil, false
	}
	return &config.AuthenticationFlowBotProtection{
		Mode: cfg.BotProtection.Requirements.OOBOTPEmail.Mode,
	}, true
}
func getBotProtectionRequirementsOOBOTPSMS(cfg *config.AppConfig) (botProtection *config.AuthenticationFlowBotProtection, exist bool) {
	if !isBotProtectionEnabled(cfg) {
		return nil, false
	}
	if cfg.BotProtection.Requirements == nil || cfg.BotProtection.Requirements.OOBOTPSMS == nil {
		return nil, false
	}
	return &config.AuthenticationFlowBotProtection{
		Mode: cfg.BotProtection.Requirements.OOBOTPSMS.Mode,
	}, true
}
