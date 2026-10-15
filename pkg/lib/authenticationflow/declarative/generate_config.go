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

// GenerateAccountRecoveryFlowConfigIfRecoverable returns nil when cfg gives no way to recover an account.
// The generated flow then has an identify step without options, which is not a valid custom flow.
func GenerateAccountRecoveryFlowConfigIfRecoverable(cfg *config.AppConfig) *config.AuthenticationFlowAccountRecoveryFlow {
	flow := GenerateAccountRecoveryFlowConfig(cfg)
	for _, step := range flow.Steps {
		if step.Type == config.AuthenticationFlowAccountRecoveryFlowTypeIdentify && len(step.OneOf) == 0 {
			return nil
		}
	}
	return flow
}

// nameStepIdentify returns a name that is unique across flow types.
// In account linking, if the same is used in signup flow, and login flow,
// then in type: verify, the type: identify in the login flow is found,
// which is incorrect.
func nameStepIdentify(flowType config.AuthenticationFlowType) string {
	return fmt.Sprintf("%v_identify", flowType)
}
