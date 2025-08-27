package declarative

import (
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authflow.RegisterNode(&NodeUseAccountRecoveryCode{})
}

type NodeUseAccountRecoveryCode struct {
	JSONPointer  jsonpointer.T `json:"json_pointer,omitempty"`
	Code         string        `json:"code,omitempty"`
	MaskedTarget string        `json:"masked_target,omitempty"`
}

var _ authflow.NodeSimple = &NodeUseAccountRecoveryCode{}
var _ authflow.Milestone = &NodeUseAccountRecoveryCode{}
var _ MilestoneAccountRecoveryCode = &NodeUseAccountRecoveryCode{}

func (*NodeUseAccountRecoveryCode) Kind() string {
	return "NodeUseAccountRecoveryCode"
}

func (*NodeUseAccountRecoveryCode) Milestone() {}
func (n *NodeUseAccountRecoveryCode) MilestoneAccountRecoveryCode() struct {
	MaskedTarget string
	Code         string
} {
	return struct {
		MaskedTarget string
		Code         string
	}{
		MaskedTarget: n.MaskedTarget,
		Code:         n.Code,
	}
}
