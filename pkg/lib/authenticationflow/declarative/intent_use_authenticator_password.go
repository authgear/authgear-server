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
	authflow.RegisterIntent(&IntentUseAuthenticatorPassword{})
}

type IntentUseAuthenticatorPassword struct {
	JSONPointer    jsonpointer.T                           `json:"json_pointer,omitempty"`
	UserID         string                                  `json:"user_id,omitempty"`
	Authentication config.AuthenticationFlowAuthentication `json:"authentication,omitempty"`
}

var _ authflow.Intent = &IntentUseAuthenticatorPassword{}
var _ authflow.Milestone = &IntentUseAuthenticatorPassword{}
var _ MilestoneFlowSelectAuthenticationMethod = &IntentUseAuthenticatorPassword{}
var _ MilestoneDidSelectAuthenticationMethod = &IntentUseAuthenticatorPassword{}
var _ MilestoneFlowAuthenticate = &IntentUseAuthenticatorPassword{}
var _ authflow.InputReactor = &IntentUseAuthenticatorPassword{}

func (*IntentUseAuthenticatorPassword) Kind() string {
	return "IntentUseAuthenticatorPassword"
}

func (*IntentUseAuthenticatorPassword) Milestone() {}
func (n *IntentUseAuthenticatorPassword) MilestoneFlowSelectAuthenticationMethod(flows authflow.Flows) (MilestoneDidSelectAuthenticationMethod, authflow.Flows, bool) {
	return n, flows, true
}
func (n *IntentUseAuthenticatorPassword) MilestoneDidSelectAuthenticationMethod() config.AuthenticationFlowAuthentication {
	return n.Authentication
}

func (n *IntentUseAuthenticatorPassword) MilestoneFlowAuthenticate(flows authflow.Flows) (MilestoneDidAuthenticate, authflow.Flows, bool) {
	return authflow.FindMilestoneInCurrentFlow[MilestoneDidAuthenticate](flows)
}

func (n *IntentUseAuthenticatorPassword) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	_, _, authenticated := authflow.FindMilestoneInCurrentFlow[MilestoneDidAuthenticate](flows)
	if authenticated {
		return nil, authflow.ErrEOF
	}
	flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
	if err != nil {
		return nil, err
	}
	isBotProtectionRequired, err := IsBotProtectionRequired(ctx, deps, flows, n.JSONPointer)
	if err != nil {
		return nil, err
	}

	return &InputSchemaTakePassword{
		FlowRootObject:          flowRootObject,
		JSONPointer:             n.JSONPointer,
		IsBotProtectionRequired: isBotProtectionRequired,
		BotProtectionCfg:        deps.Config.BotProtection,
	}, nil
}

func (i *IntentUseAuthenticatorPassword) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	var inputTakePassword inputTakePassword
	if authflow.AsInput(input, &inputTakePassword) {
		var bpSpecialErr error
		bpSpecialErr, err := HandleBotProtection(ctx, deps, flows, i.JSONPointer, input)
		if err != nil {
			return nil, err
		}
		as, err := deps.Authenticators.List(
			i.UserID,
			authenticator.KeepKind(i.Authentication.AuthenticatorKind()),
			authenticator.KeepType(model.AuthenticatorTypePassword),
		)
		if err != nil {
			return nil, err
		}

		password := inputTakePassword.GetPassword()
		spec := &authenticator.Spec{
			Password: &authenticator.PasswordSpec{
				PlainPassword: password,
			},
		}

		info, verifyResult, err := deps.Authenticators.VerifyOneWithSpec(
			i.UserID,
			model.AuthenticatorTypePassword,
			as,
			spec,
			&facade.VerifyOptions{
				AuthenticationDetails: facade.NewAuthenticationDetails(
					i.UserID,
					authn.AuthenticationStageFromAuthenticationMethod(i.Authentication),
					authn.AuthenticationTypePassword,
				),
			},
		)
		if err != nil {
			return nil, err
		}

		var reason PasswordChangeReason
		if verifyResult.Password.ExpiryForceChange {
			reason = PasswordChangeReasonExpiry
		} else {
			reason = PasswordChangeReasonPolicy
		}

		return authflow.NewNodeSimple(&NodeDoUseAuthenticatorPassword{
			Authenticator:          info,
			PasswordChangeRequired: verifyResult.Password.RequireUpdate(),
			PasswordChangeReason:   reason,
			JSONPointer:            i.JSONPointer,
		}), bpSpecialErr
	}

	return nil, authflow.ErrIncompatibleInput
}
