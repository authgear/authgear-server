package declarative

import (
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

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

func GetBotProtectionData(flows authflow.Flows, authflowCfg *config.AuthenticationFlowBotProtection, appCfg *config.BotProtectionConfig) *BotProtectionData {
	if appCfg == nil || !appCfg.Enabled || appCfg.Provider == nil || appCfg.Provider.Type == "" {
		return nil
	}

	var effectiveMode config.BotProtectionRiskMode
	if authflowCfg != nil {
		effectiveMode = authflowCfg.Mode
	}
	_ = authflow.TraverseFlowIntentFirst(authflow.Traverser{
		NodeSimple: func(nodeSimple authflow.NodeSimple, w *authflow.Flow) error {
			if n, ok := nodeSimple.(MilestoneBotProjectionRequirementsProvider); ok {
				if c := n.MilestoneBotProjectionRequirementsProvider(); c != nil {
					effectiveMode = c.Mode
				}
			}
			return nil
		},
		Intent: func(intent authflow.Intent, w *authflow.Flow) error {
			if i, ok := intent.(MilestoneBotProjectionRequirementsProvider); ok {
				if c := i.MilestoneBotProjectionRequirementsProvider(); c != nil {
					effectiveMode = c.Mode
				}
			}
			return nil
		},
	}, flows.Root)

	switch effectiveMode {
	case config.BotProtectionRiskModeNever:
		return nil
	case config.BotProtectionRiskModeAlways:
		return NewBotProtectionData(appCfg.Provider.Type)
	default:
		return nil
	}
}
