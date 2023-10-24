package declarative

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func GeneratePromoteFlowConfig(cfg *config.AppConfig) *config.AuthenticationFlowSignupFlow {
	return GenerateSignupFlowConfig(cfg)
}
