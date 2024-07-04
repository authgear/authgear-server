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
	Options        []AuthenticateOption                    `json:"options,omitempty"`
}

var _ authflow.Intent = &IntentUseAuthenticatorOOBOTP{}
var _ authflow.Milestone = &IntentUseAuthenticatorOOBOTP{}
var _ MilestoneAuthenticationMethod = &IntentUseAuthenticatorOOBOTP{}
var _ MilestoneFlowAuthenticate = &IntentUseAuthenticatorOOBOTP{}

func (*IntentUseAuthenticatorOOBOTP) Kind() string {
	return "IntentUseAuthenticatorOOBOTP"
}

func (*IntentUseAuthenticatorOOBOTP) Milestone() {}
func (n *IntentUseAuthenticatorOOBOTP) MilestoneAuthenticationMethod() config.AuthenticationFlowAuthentication {
	return n.Authentication
}

func (*IntentUseAuthenticatorOOBOTP) MilestoneFlowAuthenticate(flows authflow.Flows) (MilestoneDidAuthenticate, authflow.Flows, bool) {
	return authflow.FindMilestoneInCurrentFlow[MilestoneDidAuthenticate](flows)
}

func (n *IntentUseAuthenticatorOOBOTP) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	_, _, authenticatorSelected := authflow.FindMilestoneInCurrentFlow[MilestoneDidSelectAuthenticator](flows)
	_, _, claimVerified := authflow.FindMilestoneInCurrentFlow[MilestoneDoMarkClaimVerified](flows)
	_, _, authenticated := authflow.FindMilestoneInCurrentFlow[MilestoneDidAuthenticate](flows)

	flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
	if err != nil {
		return nil, err
	}

	switch {
	case !authenticatorSelected:
		return &InputSchemaUseAuthenticatorOOBOTP{
			FlowRootObject: flowRootObject,
			JSONPointer:    n.JSONPointer,
			Options:        n.Options,
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
	m, _, authenticatorSelected := authflow.FindMilestoneInCurrentFlow[MilestoneDidSelectAuthenticator](flows)
	_, _, claimVerified := authflow.FindMilestoneInCurrentFlow[MilestoneDoMarkClaimVerified](flows)
	_, _, authenticated := authflow.FindMilestoneInCurrentFlow[MilestoneDidAuthenticate](flows)

	switch {
	case !authenticatorSelected:
		var inputTakeAuthenticationOptionIndex inputTakeAuthenticationOptionIndex
		if authflow.AsInput(input, &inputTakeAuthenticationOptionIndex) {
			var bpSpecialErr error
			bpRequired, err := IsNodeBotProtectionRequired(ctx, deps, flows, n.JSONPointer)
			if err != nil {
				return nil, err
			}
			if bpRequired {
				var inputTakeBotProtection inputTakeBotProtection
				if !authflow.AsInput(input, &inputTakeBotProtection) {
					return nil, authflow.ErrIncompatibleInput
				}

				token := inputTakeBotProtection.GetBotProtectionProviderResponse()
				bpSpecialErr, err = HandleBotProtection(ctx, deps, token)
				if err != nil {
					return nil, err
				}
				if !IsBotProtectionSpecialErrorSuccess(bpSpecialErr) {
					return nil, bpSpecialErr
				}
			}
			index := inputTakeAuthenticationOptionIndex.GetIndex()
			info, isNew, err := n.pickAuthenticator(deps, n.Options, index)
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
		claimName, _ := info.OOBOTP.ToClaimPair()
		purpose := otp.PurposeOOBOTP
		otpForm := getOTPForm(purpose, claimName, deps.Config.Authenticator.OOB.Email)
		return authflow.NewSubFlow(&IntentAuthenticationOOB{
			JSONPointer:    n.JSONPointer,
			UserID:         n.UserID,
			Purpose:        purpose,
			Authentication: n.Authentication,
			Info:           info,
			Form:           otpForm,
		}), nil
	case !authenticated:
		info := m.MilestoneDidSelectAuthenticator()
		return authflow.NewNodeSimple(&NodeDoUseAuthenticatorSimple{
			Authenticator: info,
		}), nil
	}

	return nil, authflow.ErrIncompatibleInput
}

// nolint:gocognit
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
