package declarative

import (
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterNode(&NodeDoUseAccountRecoveryIdentity{})
}

type NodeDoUseAccountRecoveryIdentity struct {
	Identification config.AuthenticationFlowRequestAccountRecoveryIdentification `json:"identification,omitempty"`
	Spec           *identity.Spec                                                `json:"spec,omitempty"`
	MaybeIdentity  *identity.Info                                                `json:"maybe_identity,omitempty"`
}

var _ authflow.NodeSimple = &NodeDoUseAccountRecoveryIdentity{}
var _ authflow.Milestone = &NodeDoUseAccountRecoveryIdentity{}
var _ MilestoneDoUseAccountRecoveryIdentity = &NodeDoUseAccountRecoveryIdentity{}

func (*NodeDoUseAccountRecoveryIdentity) Kind() string {
	return "NodeDoUseAccountRecoveryIdentity"
}

func (*NodeDoUseAccountRecoveryIdentity) Milestone() {}
func (n *NodeDoUseAccountRecoveryIdentity) MilestoneDoUseAccountRecoveryIdentity() AccountRecoveryIdentity {
	return AccountRecoveryIdentity{
		Identification: n.Identification,
		IdentitySpec:   n.Spec,
		MaybeIdentity:  n.MaybeIdentity,
	}
}
