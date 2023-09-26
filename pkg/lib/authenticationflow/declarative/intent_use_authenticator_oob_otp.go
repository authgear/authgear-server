package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentUseAuthenticatorOOBOTP{})
}

type IntentUseAuthenticatorOOBOTPData struct {
	Options []UseAuthenticationOption `json:"options"`
}

var _ authflow.Data = IntentUseAuthenticatorOOBOTPData{}

func (m IntentUseAuthenticatorOOBOTPData) Data() {}

type IntentUseAuthenticatorOOBOTP struct {
	JSONPointer    jsonpointer.T                           `json:"json_pointer,omitempty"`
	UserID         string                                  `json:"user_id,omitempty"`
	Authentication config.AuthenticationFlowAuthentication `json:"authentication,omitempty"`
}

var _ authflow.Intent = &IntentUseAuthenticatorOOBOTP{}
var _ authflow.Milestone = &IntentUseAuthenticatorOOBOTP{}
var _ MilestoneAuthenticationMethod = &IntentUseAuthenticatorOOBOTP{}
var _ authflow.DataOutputer = &IntentUseAuthenticatorOOBOTP{}

func (*IntentUseAuthenticatorOOBOTP) Kind() string {
	return "IntentUseAuthenticatorOOBOTP"
}

func (*IntentUseAuthenticatorOOBOTP) Milestone() {}
func (n *IntentUseAuthenticatorOOBOTP) MilestoneAuthenticationMethod() config.AuthenticationFlowAuthentication {
	return n.Authentication
}

func (n *IntentUseAuthenticatorOOBOTP) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	current, err := authflow.FlowObject(authflow.GetFlowRootObject(ctx), n.jsonPointerToStep())
	if err != nil {
		return nil, err
	}
	step := n.step(current)

	options, err := getAuthenticationOptionsForStep(ctx, deps, flows, n.UserID, step)
	if err != nil {
		return nil, err
	}
	_, authenticatorSelected := authflow.FindMilestone[MilestoneDidSelectAuthenticator](flows.Nearest)
	_, claimVerified := authflow.FindMilestone[MilestoneDoMarkClaimVerified](flows.Nearest)
	_, authenticated := authflow.FindMilestone[MilestoneDidAuthenticate](flows.Nearest)

	switch {
	case !authenticatorSelected:
		return &InputSchemaUseAuthenticatorOOBOTP{
			JSONPointer: n.JSONPointer,
			Options:     options,
		}, nil
	case !claimVerified:
		// Verify the claim
		return nil, nil
	case !authenticated:
		// Achieve the milestone.
		return nil, nil
	default:
		return nil, authflow.ErrEOF
	}
}

func (n *IntentUseAuthenticatorOOBOTP) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	m, authenticatorSelected := authflow.FindMilestone[MilestoneDidSelectAuthenticator](flows.Nearest)
	_, claimVerified := authflow.FindMilestone[MilestoneDoMarkClaimVerified](flows.Nearest)
	_, authenticated := authflow.FindMilestone[MilestoneDidAuthenticate](flows.Nearest)

	switch {
	case !authenticatorSelected:
		var inputTakeAuthenticationOptionIndex inputTakeAuthenticationOptionIndex
		if authflow.AsInput(input, &inputTakeAuthenticationOptionIndex) {
			current, err := authflow.FlowObject(authflow.GetFlowRootObject(ctx), n.jsonPointerToStep())
			if err != nil {
				return nil, err
			}
			step := n.step(current)

			options, err := getAuthenticationOptionsForStep(ctx, deps, flows, n.UserID, step)
			if err != nil {
				return nil, err
			}

			index := inputTakeAuthenticationOptionIndex.GetIndex()
			info, err := n.pickAuthenticator(deps, options, index)
			if err != nil {
				return nil, err
			}

			return authflow.NewNodeSimple(&NodeDidSelectAuthenticator{
				Authenticator: info,
			}), nil
		}
	case !claimVerified:
		info := m.MilestoneDidSelectAuthenticator()
		claimName, claimValue := info.OOBOTP.ToClaimPair()
		return authflow.NewSubFlow(&IntentVerifyClaim{
			JSONPointer: n.JSONPointer,
			UserID:      n.UserID,
			Purpose:     otp.PurposeOOBOTP,
			MessageType: n.otpMessageType(info),
			ClaimName:   claimName,
			ClaimValue:  claimValue,
		}), nil
	case !authenticated:
		info := m.MilestoneDidSelectAuthenticator()
		return authflow.NewNodeSimple(&NodeDoUseAuthenticatorSimple{
			Authenticator: info,
		}), nil
	}

	return nil, authflow.ErrIncompatibleInput
}

func (n *IntentUseAuthenticatorOOBOTP) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	current, err := authflow.FlowObject(authflow.GetFlowRootObject(ctx), n.jsonPointerToStep())
	if err != nil {
		return nil, err
	}
	step := n.step(current)

	options, err := getAuthenticationOptionsForStep(ctx, deps, flows, n.UserID, step)
	if err != nil {
		return nil, err
	}

	return IntentUseAuthenticatorOOBOTPData{
		Options: options,
	}, nil
}

func (*IntentUseAuthenticatorOOBOTP) step(o config.AuthenticationFlowObject) *config.AuthenticationFlowLoginFlowStep {
	step, ok := o.(*config.AuthenticationFlowLoginFlowStep)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}

	return step
}

func (n *IntentUseAuthenticatorOOBOTP) pickAuthenticator(deps *authflow.Dependencies, options []UseAuthenticationOption, index int) (*authenticator.Info, error) {
	for idx, c := range options {
		if idx == index {
			id := c.AuthenticatorID
			info, err := deps.Authenticators.Get(id)
			if err != nil {
				return nil, err
			}
			return info, nil
		}
	}

	return nil, authflow.ErrIncompatibleInput
}

func (*IntentUseAuthenticatorOOBOTP) otpMessageType(info *authenticator.Info) otp.MessageType {
	switch info.Kind {
	case model.AuthenticatorKindPrimary:
		return otp.MessageTypeAuthenticatePrimaryOOB
	case model.AuthenticatorKindSecondary:
		return otp.MessageTypeAuthenticateSecondaryOOB
	default:
		panic(fmt.Errorf("unexpected OOB OTP authenticator kind: %v", info.Kind))
	}
}

func (i *IntentUseAuthenticatorOOBOTP) jsonPointerToStep() jsonpointer.T {
	return authflow.JSONPointerToParent(i.JSONPointer)
}
