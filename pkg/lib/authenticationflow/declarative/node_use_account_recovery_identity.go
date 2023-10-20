package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterNode(&NodeUseAccountRecoveryIdentity{})
}

type NodeUseAccountRecoveryIdentity struct {
	JSONPointer    jsonpointer.T                                                          `json:"json_pointer,omitempty"`
	Identification config.AuthenticationFlowRequestAccountRecoveryIdentification          `json:"identification,omitempty"`
	OnFailure      config.AuthenticationFlowRequestAccountRecoveryIdentificationOnFailure `json:"on_failure,omitempty"`
}

var _ authflow.NodeSimple = &NodeUseAccountRecoveryIdentity{}
var _ authflow.Milestone = &NodeUseAccountRecoveryIdentity{}
var _ MilestoneAccountRecoveryIdentificationMethod = &NodeUseAccountRecoveryIdentity{}
var _ authflow.InputReactor = &NodeUseAccountRecoveryIdentity{}

func (*NodeUseAccountRecoveryIdentity) Kind() string {
	return "NodeUseAccountRecoveryIdentity"
}

func (*NodeUseAccountRecoveryIdentity) Milestone() {}
func (n *NodeUseAccountRecoveryIdentity) MilestoneAccountRecoveryIdentificationMethod() config.AuthenticationFlowRequestAccountRecoveryIdentification {
	return n.Identification
}

func (n *NodeUseAccountRecoveryIdentity) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	return &InputSchemaTakeLoginID{
		JSONPointer: n.JSONPointer,
	}, nil
}

func (n *NodeUseAccountRecoveryIdentity) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	var inputTakeLoginID inputTakeLoginID
	if authflow.AsInput(input, &inputTakeLoginID) {
		loginID := inputTakeLoginID.GetLoginID()
		spec := &identity.Spec{
			Type: model.IdentityTypeLoginID,
			LoginID: &identity.LoginIDSpec{
				Value: loginID,
			},
		}

		var exactMatch *identity.Info = nil
		switch n.OnFailure {
		case config.AuthenticationFlowRequestAccountRecoveryIdentificationOnFailureIgnore:
			var err error
			exactMatch, _, err = deps.Identities.SearchBySpec(spec)
			if err != nil {
				return nil, err
			}
		case config.AuthenticationFlowRequestAccountRecoveryIdentificationOnFailureError:
			var err error
			exactMatch, err = findExactOneIdentityInfo(deps, spec)
			if err != nil {
				return nil, err
			}
		}

		n := &NodeDoUseAccountRecoveryIdentity{
			MaybeIdentity: exactMatch,
		}

		return authflow.NewNodeSimple(n), nil
	}

	return nil, authflow.ErrIncompatibleInput
}
