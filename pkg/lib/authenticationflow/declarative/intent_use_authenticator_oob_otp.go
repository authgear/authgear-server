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
	Candidates []UseAuthenticationCandidate `json:"candidates"`
}

var _ authflow.Data = IntentUseAuthenticatorOOBOTPData{}

func (m IntentUseAuthenticatorOOBOTPData) Data() {}

type IntentUseAuthenticatorOOBOTP struct {
	FlowReference     authflow.FlowReference                  `json:"flow_reference,omitempty"`
	JSONPointer       jsonpointer.T                           `json:"json_pointer,omitempty"`
	JSONPointerToStep jsonpointer.T                           `json:"json_pointer_to_step,omitempty"`
	UserID            string                                  `json:"user_id,omitempty"`
	Authentication    config.AuthenticationFlowAuthentication `json:"authentication,omitempty"`
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
	current, err := flowObject(authflow.GetFlowRootObject(ctx), n.JSONPointerToStep)
	if err != nil {
		return nil, err
	}
	step := n.step(current)

	candidates, err := getAuthenticationCandidatesForStep(ctx, deps, flows, n.UserID, step)
	if err != nil {
		return nil, err
	}
	_, authenticatorSelected := authflow.FindMilestone[MilestoneDidSelectAuthenticator](flows.Nearest)
	_, claimVerified := authflow.FindMilestone[MilestoneDoMarkClaimVerified](flows.Nearest)
	_, authenticatorVerified := authflow.FindMilestone[MilestoneDidVerifyAuthenticator](flows.Nearest)

	switch {
	case !authenticatorSelected:
		return &InputSchemaUseAuthenticatorOOBOTP{
			Candidates: candidates,
		}, nil
	case !claimVerified:
		// Verify the claim
		return nil, nil
	case !authenticatorVerified:
		// Achieve the milestone.
		return nil, nil
	default:
		return nil, authflow.ErrEOF
	}
}

func (n *IntentUseAuthenticatorOOBOTP) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	m, authenticatorSelected := authflow.FindMilestone[MilestoneDidSelectAuthenticator](flows.Nearest)
	_, claimVerified := authflow.FindMilestone[MilestoneDoMarkClaimVerified](flows.Nearest)
	_, authenticatorVerified := authflow.FindMilestone[MilestoneDidVerifyAuthenticator](flows.Nearest)

	switch {
	case !authenticatorSelected:
		var inputTakeAuthenticationCandidateIndex inputTakeAuthenticationCandidateIndex
		if authflow.AsInput(input, &inputTakeAuthenticationCandidateIndex) {
			current, err := flowObject(authflow.GetFlowRootObject(ctx), n.JSONPointerToStep)
			if err != nil {
				return nil, err
			}
			step := n.step(current)

			candidates, err := getAuthenticationCandidatesForStep(ctx, deps, flows, n.UserID, step)
			if err != nil {
				return nil, err
			}

			index := inputTakeAuthenticationCandidateIndex.GetIndex()
			info, err := n.pickAuthenticator(deps, candidates, index)
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
			UserID:      n.UserID,
			Purpose:     otp.PurposeOOBOTP,
			MessageType: n.otpMessageType(info),
			ClaimName:   claimName,
			ClaimValue:  claimValue,
		}), nil
	case !authenticatorVerified:
		info := m.MilestoneDidSelectAuthenticator()
		return authflow.NewNodeSimple(&NodeDidVerifyAuthenticator{
			Authenticator: info,
		}), nil
	}

	return nil, authflow.ErrIncompatibleInput
}

func (n *IntentUseAuthenticatorOOBOTP) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	current, err := flowObject(authflow.GetFlowRootObject(ctx), n.JSONPointerToStep)
	if err != nil {
		return nil, err
	}
	step := n.step(current)

	candidates, err := getAuthenticationCandidatesForStep(ctx, deps, flows, n.UserID, step)
	if err != nil {
		return nil, err
	}

	return IntentUseAuthenticatorOOBOTPData{
		Candidates: candidates,
	}, nil
}

func (*IntentUseAuthenticatorOOBOTP) step(o config.AuthenticationFlowObject) *config.AuthenticationFlowLoginFlowStep {
	step, ok := o.(*config.AuthenticationFlowLoginFlowStep)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}

	return step
}

func (n *IntentUseAuthenticatorOOBOTP) pickAuthenticator(deps *authflow.Dependencies, candidates []UseAuthenticationCandidate, index int) (*authenticator.Info, error) {
	for idx, c := range candidates {
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
