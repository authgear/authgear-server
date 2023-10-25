package declarative

import (
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authflow.RegisterNode(&NodeUseAccountRecoveryCode{})
}

type NodeUseAccountRecoveryCode struct {
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
	Code        string        `json:"code,omitempty"`
}

var _ authflow.NodeSimple = &NodeUseAccountRecoveryCode{}
var _ authflow.Milestone = &NodeUseAccountRecoveryCode{}
var _ MilestoneAccountRecoveryCode = &NodeUseAccountRecoveryCode{}

func (*NodeUseAccountRecoveryCode) Kind() string {
	return "NodeUseAccountRecoveryCode"
}

func (*NodeUseAccountRecoveryCode) Milestone() {}
func (n *NodeUseAccountRecoveryCode) MilestoneAccountRecoveryCode() string {
	return n.Code
}
