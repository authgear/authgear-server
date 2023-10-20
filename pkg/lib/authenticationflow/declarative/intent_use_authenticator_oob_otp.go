package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentUseAuthenticatorOOBOTP{})
}

type IntentUseAuthenticatorOOBOTP struct {
	JSONPointer    jsonpointer.T                           `json:"json_pointer,omitempty"`
	UserID         string                                  `json:"user_id,omitempty"`
	Authentication config.AuthenticationFlowAuthentication `json:"authentication,omitempty"`
}

var _ authflow.Intent = &IntentUseAuthenticatorOOBOTP{}
var _ authflow.Milestone = &IntentUseAuthenticatorOOBOTP{}
var _ MilestoneAuthenticationMethod = &IntentUseAuthenticatorOOBOTP{}

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
			info, isNew, err := n.pickAuthenticator(deps, options, index)
			if err != nil {
				return nil, err
			}

			if isNew {
				return authflow.NewNodeSimple(&NodeDoJustInTimeCreateAuthenticator{
					Authenticator: info,
				}), nil
			}

			return authflow.NewNodeSimple(&NodeDidSelectAuthenticator{
				Authenticator: info,
			}), nil
		}
	case !claimVerified:
		info := m.MilestoneDidSelectAuthenticator()
		claimName, claimValue := info.OOBOTP.ToClaimPair()
		purpose := otp.PurposeOOBOTP
		otpForm := getOTPForm(purpose, claimName, deps.Config.Authenticator.OOB.Email)
		return authflow.NewSubFlow(&IntentVerifyClaim{
			JSONPointer: n.JSONPointer,
			UserID:      n.UserID,
			Purpose:     purpose,
			MessageType: n.otpMessageType(info),
			Form:        otpForm,
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

func (*IntentUseAuthenticatorOOBOTP) step(o config.AuthenticationFlowObject) *config.AuthenticationFlowLoginFlowStep {
	step, ok := o.(*config.AuthenticationFlowLoginFlowStep)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}

	return step
}

func (n *IntentUseAuthenticatorOOBOTP) pickAuthenticator(deps *authflow.Dependencies, options []AuthenticateOption, index int) (info *authenticator.Info, isNew bool, err error) {
	for idx, c := range options {
		if idx == index {
			switch {
			case c.AuthenticatorID != "":
				info, err = deps.Authenticators.Get(c.AuthenticatorID)
				if err != nil {
					return
				}

				return
			case c.IdentityID != "":
				var identityInfo *identity.Info
				identityInfo, err = deps.Identities.Get(c.IdentityID)
				if err != nil {
					return
				}

				info, err = n.createAuthenticator(deps, identityInfo)
				if err != nil {
					return
				}

				// Check if the just-in-time authenticator is a duplicate.
				var allAuthenticators []*authenticator.Info
				allAuthenticators, err = deps.Authenticators.List(n.UserID)
				if err != nil {
					return
				}

				for _, authenticator := range allAuthenticators {
					if authenticator.Equal(info) {
						// An existing authenticator is found.
						info = authenticator
						isNew = false
						return
					}
				}

				// If we reach here, then we need to just-in-time create the authenticator.
				isNew = true
				return
			default:
				panic(fmt.Errorf("expected option to have either IdentityID or AuthenticatorID"))
			}
		}
	}

	err = authflow.ErrIncompatibleInput
	return
}

func (n *IntentUseAuthenticatorOOBOTP) createAuthenticator(deps *authflow.Dependencies, info *identity.Info) (*authenticator.Info, error) {
	if info.Type != model.IdentityTypeLoginID || info.LoginID == nil {
		panic(fmt.Errorf("expected only Login ID identity can create OOB OTP authenticator just-in-time"))
	}

	target := info.LoginID.LoginID
	authenticatorInfo, err := createAuthenticator(deps, n.UserID, n.Authentication, target)
	if err != nil {
		return nil, err
	}

	return authenticatorInfo, nil
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
