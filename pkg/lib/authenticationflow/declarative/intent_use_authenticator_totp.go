package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/facade"
)

func init() {
	authflow.RegisterIntent(&IntentUseAuthenticatorTOTP{})
}

type IntentUseAuthenticatorTOTP struct {
	JSONPointer    jsonpointer.T                           `json:"json_pointer,omitempty"`
	UserID         string                                  `json:"user_id,omitempty"`
	Authentication config.AuthenticationFlowAuthentication `json:"authentication,omitempty"`
}

var _ authflow.Intent = &IntentUseAuthenticatorTOTP{}
var _ authflow.Milestone = &IntentUseAuthenticatorTOTP{}
var _ MilestoneAuthenticationMethod = &IntentUseAuthenticatorTOTP{}
var _ MilestoneFlowAuthenticate = &IntentUseAuthenticatorTOTP{}
var _ authflow.InputReactor = &IntentUseAuthenticatorTOTP{}

func (*IntentUseAuthenticatorTOTP) Kind() string {
	return "IntentUseAuthenticatorTOTP"
}

func (*IntentUseAuthenticatorTOTP) Milestone() {}
func (n *IntentUseAuthenticatorTOTP) MilestoneAuthenticationMethod() config.AuthenticationFlowAuthentication {
	return n.Authentication
}

func (*IntentUseAuthenticatorTOTP) MilestoneFlowAuthenticate(flows authflow.Flows) (MilestoneDidAuthenticate, authflow.Flows, bool) {
	return authflow.FindMilestoneInCurrentFlow[MilestoneDidAuthenticate](flows)
}

func (n *IntentUseAuthenticatorTOTP) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	_, _, authenticated := authflow.FindMilestoneInCurrentFlow[MilestoneDidAuthenticate](flows)
	if authenticated {
		return nil, authflow.ErrEOF
	}
	flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
	if err != nil {
		return nil, err
	}

	return &InputSchemaTakeTOTP{
		JSONPointer:    n.JSONPointer,
		FlowRootObject: flowRootObject,
	}, nil
}

func (n *IntentUseAuthenticatorTOTP) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	var inputTakeTOTP inputTakeTOTP
	if authflow.AsInput(input, &inputTakeTOTP) {
		as, err := deps.Authenticators.List(
			n.UserID,
			authenticator.KeepKind(n.Authentication.AuthenticatorKind()),
			authenticator.KeepType(model.AuthenticatorTypeTOTP),
		)
		if err != nil {
			return nil, err
		}

		code := inputTakeTOTP.GetCode()
		spec := &authenticator.Spec{
			TOTP: &authenticator.TOTPSpec{
				Code: code,
			},
		}

		info, _, err := deps.Authenticators.VerifyOneWithSpec(
			n.UserID,
			model.AuthenticatorTypeTOTP,
			as,
			spec,
			&facade.VerifyOptions{
				AuthenticationDetails: facade.NewAuthenticationDetails(
					n.UserID,
					authn.AuthenticationStageFromAuthenticationMethod(n.Authentication),
					authn.AuthenticationTypeTOTP,
				),
			},
		)
		if err != nil {
			return nil, err
		}

		return authflow.NewNodeSimple(&NodeDoUseAuthenticatorSimple{
			Authenticator: info,
		}), nil
	}

	return nil, authflow.ErrIncompatibleInput
}
