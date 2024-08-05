package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/service"
)

func init() {
	authflow.RegisterNode(&NodeLoginFlowChangePassword{})
}

type NodeLoginFlowChangePassword struct {
	JSONPointer   jsonpointer.T         `json:"json_pointer,omitempty"`
	Authenticator *authenticator.Info   `json:"authenticator,omitempty"`
	Reason        *PasswordChangeReason `json:"reason,omitempty"`
}

func (n *NodeLoginFlowChangePassword) GetChangeReason() *PasswordChangeReason {
	return n.Reason
}

var _ authflow.NodeSimple = &NodeLoginFlowChangePassword{}
var _ authflow.InputReactor = &NodeLoginFlowChangePassword{}
var _ authflow.DataOutputer = &NodeLoginFlowChangePassword{}

func (*NodeLoginFlowChangePassword) Kind() string {
	return "NodeLoginFlowChangePassword"
}

func (n *NodeLoginFlowChangePassword) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
	if err != nil {
		return nil, err
	}

	return &InputSchemaTakeNewPassword{
		FlowRootObject: flowRootObject,
		JSONPointer:    n.JSONPointer,
	}, nil
}

func (n *NodeLoginFlowChangePassword) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	var inputTakeNewPassword inputTakeNewPassword
	if authflow.AsInput(input, &inputTakeNewPassword) {
		newPassword := inputTakeNewPassword.GetNewPassword()

		oldInfo := n.Authenticator
		changed, newInfo, err := deps.Authenticators.UpdatePassword(oldInfo, &service.UpdatePasswordOptions{
			PlainPassword:  newPassword,
			SetExpireAfter: true,
		})
		if err != nil {
			return nil, err
		}

		if !changed && n.Reason != nil && *n.Reason == PasswordChangeReasonExpiry {
			// Password is expired, but the user did not change the password.
			return nil, api.ErrPasswordReused
		}

		if !changed {
			// Nothing changed. End this flow.
			return authflow.NewNodeSimple(&NodeSentinel{}), nil
		}

		return authflow.NewNodeSimple(&NodeDoUpdateAuthenticator{
			Authenticator: newInfo,
		}), nil
	}

	return nil, authflow.ErrIncompatibleInput
}

func (n *NodeLoginFlowChangePassword) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	return NewForceChangePasswordData(ForceChangePasswordData{
		PasswordPolicy: NewPasswordPolicy(
			deps.FeatureConfig.Authenticator,
			deps.Config.Authenticator.Password.Policy,
		),
		ForceChangeReason: n.Reason,
	}), nil
}
