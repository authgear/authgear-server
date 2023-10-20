package declarative

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type AccountRecoveryIdentificationOption struct {
	Identification config.AuthenticationFlowRequestAccountRecoveryIdentification `json:"identification"`
}
