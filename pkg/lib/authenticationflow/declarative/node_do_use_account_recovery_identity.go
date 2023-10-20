package declarative

import (
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

func init() {
	authflow.RegisterNode(&NodeDoUseAccountRecoveryIdentity{})
}

type NodeDoUseAccountRecoveryIdentity struct {
	MaybeIdentity *identity.Info `json:"maybe_identity,omitempty"`
}

var _ authflow.NodeSimple = &NodeDoUseAccountRecoveryIdentity{}
var _ authflow.Milestone = &NodeDoUseAccountRecoveryIdentity{}
var _ MilestoneDoUseAccountRecoveryIdentity = &NodeDoUseAccountRecoveryIdentity{}

func (*NodeDoUseAccountRecoveryIdentity) Kind() string {
	return "NodeDoUseAccountRecoveryIdentity"
}

func (*NodeDoUseAccountRecoveryIdentity) Milestone() {}
func (n *NodeDoUseAccountRecoveryIdentity) MilestoneDoUseAccountRecoveryIdentity() *identity.Info {
	return n.MaybeIdentity
}
