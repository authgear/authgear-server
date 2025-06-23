package declarative

import "github.com/authgear/authgear-server/pkg/lib/config"

// For identification option, authentication option & create_authenticator option

type BotProtectionData struct {
	Enabled  *bool                      `json:"enabled,omitempty"`
	Provider *BotProtectionDataProvider `json:"provider,omitempty"`
}

func (d *BotProtectionData) IsRequired() bool {
	return d != nil && d.Enabled != nil && *d.Enabled && d.Provider != nil && d.Provider.Type != ""
}

type BotProtectionDataProvider struct {
	Type config.BotProtectionProviderType `json:"type,omitempty"`
}

func NewBotProtectionData(t config.BotProtectionProviderType) *BotProtectionData {
	var varTrue = true
	return &BotProtectionData{
		Enabled: &varTrue,
		Provider: &BotProtectionDataProvider{
			Type: t,
		},
	}
}

func GetBotProtectionData(authflowCfg *config.AuthenticationFlowBotProtection, appCfg *config.BotProtectionConfig) *BotProtectionData {
	if authflowCfg == nil {
		return nil
	}
	if appCfg == nil || !appCfg.Enabled || appCfg.Provider == nil || appCfg.Provider.Type == "" {
		return nil
	}

	switch authflowCfg.Mode {
	case config.BotProtectionRiskModeNever:
		break
	case config.BotProtectionRiskModeAlways:
		return NewBotProtectionData(appCfg.Provider.Type)
	}
	return nil
}
