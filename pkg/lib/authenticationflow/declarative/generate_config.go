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

// nameStepIdentify returns a name that is unique across flow types.
// In account linking, if the same is used in signup flow, and login flow,
// then in type: verify, the type: identify in the login flow is found,
// which is incorrect.
func nameStepIdentify(flowType config.AuthenticationFlowType) string {
	return fmt.Sprintf("%v_identify", flowType)
}
