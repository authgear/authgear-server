package web

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type UIImplementationService struct {
	UIConfig                       *config.UIConfig
	GlobalUIImplementation         config.GlobalUIImplementation
	GlobalUISettingsImplementation config.GlobalUISettingsImplementation
}

func (s *UIImplementationService) GetUIImplementation() config.UIImplementation {
	return config.UIImplementationAuthflowV2
}

func (s *UIImplementationService) GetSettingsUIImplementation() config.SettingsUIImplementation {
	return config.SettingsUIImplementationV2
}
