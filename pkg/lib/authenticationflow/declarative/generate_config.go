package declarative

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

const (
	nameGeneratedFlow                   = "default"
	nameFormatStepAuthenticatePrimary   = "authenticate_primary_%s"
	nameFormatStepAuthenticateSecondary = "authenticate_secondary_%s"
	nameStepReauthenticate              = "reauthenticate"
)

// GenerateAuthenticationFlowConfig returns the flow of each type that Authgear generates from cfg.
func GenerateAuthenticationFlowConfig(cfg *config.AppConfig) *config.AuthenticationFlowConfig {
	return &config.AuthenticationFlowConfig{
		SignupFlows:          []*config.AuthenticationFlowSignupFlow{GenerateSignupFlowConfig(cfg)},
		PromoteFlows:         []*config.AuthenticationFlowSignupFlow{GeneratePromoteFlowConfig(cfg)},
		LoginFlows:           []*config.AuthenticationFlowLoginFlow{GenerateLoginFlowConfig(cfg)},
		SignupLoginFlows:     []*config.AuthenticationFlowSignupLoginFlow{GenerateSignupLoginFlowConfig(cfg)},
		ReauthFlows:          []*config.AuthenticationFlowReauthFlow{GenerateReauthFlowConfig(cfg)},
		AccountRecoveryFlows: []*config.AuthenticationFlowAccountRecoveryFlow{GenerateAccountRecoveryFlowConfig(cfg)},
	}
}

// nameStepIdentify returns a name that is unique across flow types.
// In account linking, if the same is used in signup flow, and login flow,
// then in type: verify, the type: identify in the login flow is found,
// which is incorrect.
func nameStepIdentify(flowType config.AuthenticationFlowType) string {
	return fmt.Sprintf("%v_identify", flowType)
}
