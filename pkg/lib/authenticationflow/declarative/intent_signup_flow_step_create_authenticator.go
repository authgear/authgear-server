package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

type IntentSignupFlowStepCreateAuthenticatorTarget interface {
	GetOOBOTPClaims(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (map[model.ClaimName]string, error)
	IsSkipped() bool
}

func init() {
	authflow.RegisterIntent(&IntentSignupFlowStepCreateAuthenticator{})
}

// IntentSignupFlowStepCreateAuthenticator
//
//   IntentCreateAuthenticatorPassword (MilestoneFlowCreateAuthenticator, MilestoneFlowSelectAuthenticationMethod, MilestoneFlowDidSelectAuthenticationMethod)
//     NodeDoCreateAuthenticator (MilestoneDoCreateAuthenticator)
//
//   IntentCreateAuthenticatorOOBOTP (MilestoneFlowCreateAuthenticator, MilestoneFlowSelectAuthenticationMethod, MilestoneFlowDidSelectAuthenticationMethod)
//     IntentVerifyClaim (MilestoneVerifyClaim)
//       NodeVerifyClaim
//     NodeDoCreateAuthenticator (MilestoneDoCreateAuthenticator)
//     NodeDidSelectAuthenticator (MilestoneDidSelectAuthenticator)
//
//   IntentCreateAuthenticatorTOTP (MilestoneFlowCreateAuthenticator, MilestoneFlowSelectAuthenticationMethod, MilestoneFlowDidSelectAuthenticationMethod)
//     NodeDoCreateAuthenticator (MilestoneDoCreateAuthenticator)

type IntentSignupFlowStepCreateAuthenticator struct {
	FlowReference          authflow.FlowReference `json:"flow_reference,omitempty"`
	JSONPointer            jsonpointer.T          `json:"json_pointer,omitempty"`
	StepName               string                 `json:"step_name,omitempty"`
	UserID                 string                 `json:"user_id,omitempty"`
	IsUpdatingExistingUser bool                   `json:"is_updating_existing_user,omitempty"`

	Options         []CreateAuthenticatorOptionInternal `json:"options,omitempty"`
	CannotBeSkipped bool                                `json:"cannot_be_skipped,omitempty"`
}

func NewIntentSignupFlowStepCreateAuthenticator(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, i *IntentSignupFlowStepCreateAuthenticator, originNode authflow.NodeOrIntent) (*IntentSignupFlowStepCreateAuthenticator, error) {
	current, err := i.currentFlowObject(deps, flows, originNode)
	if err != nil {
		return nil, err
	}
	step := i.step(current)
	options, err := NewCreateAuthenticationOptions(ctx, deps, flows, step, i.UserID)
	if err != nil {
		return nil, err
	}
	i.Options = options
	return i, nil
}

var _ authflow.TargetStep = &IntentSignupFlowStepCreateAuthenticator{}

func (i *IntentSignupFlowStepCreateAuthenticator) GetName() string {
	return i.StepName
}

func (i *IntentSignupFlowStepCreateAuthenticator) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

var _ authflow.Intent = &IntentSignupFlowStepCreateAuthenticator{}
var _ authflow.DataOutputer = &IntentSignupFlowStepCreateAuthenticator{}
var _ authflow.Milestone = &IntentSignupFlowStepCreateAuthenticator{}
var _ MilestoneSwitchToExistingUser = &IntentSignupFlowStepCreateAuthenticator{}

func (*IntentSignupFlowStepCreateAuthenticator) Milestone() {}
func (i *IntentSignupFlowStepCreateAuthenticator) MilestoneSwitchToExistingUser(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, newUserID string) error {
	i.UserID = newUserID
	i.IsUpdatingExistingUser = true

	m1, m1Flows, ok := authflow.FindMilestoneInCurrentFlow[MilestoneFlowCreateAuthenticator](flows)
	if ok {
		milestone, _, ok := m1.MilestoneFlowCreateAuthenticator(m1Flows)
		if ok {
			authn, ok := milestone.MilestoneDoCreateAuthenticator()
			if ok {
				existing, err := i.findAuthenticatorOfSameType(ctx, deps, authn.Type)
				if err != nil {
					return err
				}
				if existing != nil {
					milestone.MilestoneDoCreateAuthenticatorSkipCreate()
				} else {
					milestone.MilestoneDoCreateAuthenticatorUpdate(authn.UpdateUserID(newUserID))
				}
			}
		}
	}

	return nil
}

func (*IntentSignupFlowStepCreateAuthenticator) Kind() string {
	return "IntentSignupFlowStepCreateAuthenticator"
}

func (i *IntentSignupFlowStepCreateAuthenticator) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	if len(flows.Nearest.Nodes) == 0 && len(i.Options) == 0 {
		if i.CannotBeSkipped {
			return nil, api.ErrNoAuthenticator
		}
		// Nothing can be selected, skip this step.
		return nil, authflow.ErrEOF
	}

	if len(flows.Nearest.Nodes) == 0 && i.IsUpdatingExistingUser {
		option, _, _, err := i.findSkippableOption(ctx, deps, flows)
		if err != nil {
			return nil, err
		}
		if option != nil {
			// Proceed without user input to use the existing authenticator automatically
			return nil, nil
		}
	}

	// Let the input to select which authentication method to use.
	if len(flows.Nearest.Nodes) == 0 {
		flowRootObject, err := findNearestFlowObjectInFlow(deps, flows, i)
		if err != nil {
			return nil, err
		}

		options, err := i.getPublicOptions()
		if err != nil {
			return nil, err
		}

		shouldBypassBotProtection := ShouldExistingResultBypassBotProtectionRequirement(ctx)
		return &InputSchemaSignupFlowStepCreateAuthenticator{
			FlowRootObject:            flowRootObject,
			JSONPointer:               i.JSONPointer,
			Options:                   options,
			ShouldBypassBotProtection: shouldBypassBotProtection,
			BotProtectionCfg:          deps.Config.BotProtection,
		}, nil
	}

	_, _, authenticatorCreated := authflow.FindMilestoneInCurrentFlow[MilestoneFlowCreateAuthenticator](flows)
	_, _, nestedStepsHandled := authflow.FindMilestoneInCurrentFlow[MilestoneNestedSteps](flows)

	switch {
	case authenticatorCreated && !nestedStepsHandled:
		// Handle nested steps.
		return nil, nil
	default:
		return nil, authflow.ErrEOF
	}
}

func (i *IntentSignupFlowStepCreateAuthenticator) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {

	if len(flows.Nearest.Nodes) == 0 && i.IsUpdatingExistingUser {
		option, idx, authn, err := i.findSkippableOption(ctx, deps, flows)
		if err != nil {
			return nil, err
		}
		if option != nil {
			return i.reactToExistingAuthenticator(ctx, deps, flows, *option, authn, idx)
		}
	}

	current, err := i.currentFlowObject(deps, flows, i)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	if len(flows.Nearest.Nodes) == 0 {
		var inputTakeAuthenticationMethod inputTakeAuthenticationMethod
		if authflow.AsInput(input, &inputTakeAuthenticationMethod) {

			authentication := inputTakeAuthenticationMethod.GetAuthenticationMethod()
			idx, err := i.checkAuthenticationMethod(deps, step, authentication)
			if err != nil {
				return nil, err
			}

			switch authentication {
			case model.AuthenticationFlowAuthenticationPrimaryPassword:
				fallthrough
			case model.AuthenticationFlowAuthenticationSecondaryPassword:
				return authflow.NewSubFlow(&IntentCreateAuthenticatorPassword{
					JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
					UserID:         i.UserID,
					Authentication: authentication,
				}), nil
			case model.AuthenticationFlowAuthenticationPrimaryPasskey:
				// Cannot create passkey in this step.
				return nil, authflow.ErrIncompatibleInput
			case model.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
				fallthrough
			case model.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
				fallthrough
			case model.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
				fallthrough
			case model.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
				return authflow.NewSubFlow(&IntentCreateAuthenticatorOOBOTP{
					JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
					UserID:         i.UserID,
					Authentication: authentication,
				}), nil
			case model.AuthenticationFlowAuthenticationSecondaryTOTP:
				intent, err := NewIntentCreateAuthenticatorTOTP(ctx, deps, &IntentCreateAuthenticatorTOTP{
					JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
					UserID:         i.UserID,
					Authentication: authentication,
				})
				if err != nil {
					return nil, err
				}
				return authflow.NewSubFlow(intent), nil
			}
		}
		return nil, authflow.ErrIncompatibleInput
	}

	_, _, authenticatorCreated := authflow.FindMilestoneInCurrentFlow[MilestoneFlowCreateAuthenticator](flows)
	_, _, nestedStepsHandled := authflow.FindMilestoneInCurrentFlow[MilestoneNestedSteps](flows)

	switch {
	case authenticatorCreated && !nestedStepsHandled:
		authentication := i.authenticationMethod(flows)
		return authflow.NewSubFlow(&IntentSignupFlowSteps{
			FlowReference:          i.FlowReference,
			JSONPointer:            i.jsonPointer(step, authentication),
			UserID:                 i.UserID,
			IsUpdatingExistingUser: i.IsUpdatingExistingUser,
		}), nil
	default:
		return nil, authflow.ErrIncompatibleInput
	}
}

func (i *IntentSignupFlowStepCreateAuthenticator) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	options, err := i.getPublicOptions()
	if err != nil {
		return nil, err
	}

	optionsForOutput := []CreateAuthenticatorOptionForOutput{}
	for _, o := range options {
		optionsForOutput = append(optionsForOutput, o.ToOutput(ctx))
	}

	return NewCreateAuthenticatorData(CreateAuthenticatorData{
		Options: optionsForOutput,
	}), nil
}

func (*IntentSignupFlowStepCreateAuthenticator) step(o config.AuthenticationFlowObject) *config.AuthenticationFlowSignupFlowStep {
	step, ok := o.(*config.AuthenticationFlowSignupFlowStep)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}

	return step
}

func (i *IntentSignupFlowStepCreateAuthenticator) checkAuthenticationMethod(deps *authflow.Dependencies, step *config.AuthenticationFlowSignupFlowStep, am model.AuthenticationFlowAuthentication) (idx int, err error) {
	idx = -1

	for index, branch := range step.OneOf {
		branch := branch
		if am == branch.Authentication {
			idx = index
		}
	}

	if idx >= 0 {
		return
	}

	err = authflow.ErrIncompatibleInput
	return
}

func (*IntentSignupFlowStepCreateAuthenticator) authenticationMethod(flows authflow.Flows) model.AuthenticationFlowAuthentication {
	m, mFlows, ok := authflow.FindMilestoneInCurrentFlow[MilestoneFlowSelectAuthenticationMethod](flows)
	if !ok {
		panic(fmt.Errorf("authentication method not yet selected"))
	}

	mDidSelect, _, ok := m.MilestoneFlowSelectAuthenticationMethod(mFlows)
	if !ok {
		panic(fmt.Errorf("authentication method not yet selected"))
	}

	return mDidSelect.MilestoneDidSelectAuthenticationMethod()
}

func (i *IntentSignupFlowStepCreateAuthenticator) jsonPointer(step *config.AuthenticationFlowSignupFlowStep, am model.AuthenticationFlowAuthentication) jsonpointer.T {
	for idx, branch := range step.OneOf {
		branch := branch
		if branch.Authentication == am {
			return authflow.JSONPointerForOneOf(i.JSONPointer, idx)
		}
	}

	panic(fmt.Errorf("selected identification method is not allowed"))
}

func (i *IntentSignupFlowStepCreateAuthenticator) currentFlowObject(deps *authflow.Dependencies, flows authflow.Flows, originNode authflow.NodeOrIntent) (config.AuthenticationFlowObject, error) {
	rootObject, err := findNearestFlowObjectInFlow(deps, flows, originNode)
	if err != nil {
		return nil, err
	}
	current, err := authflow.FlowObject(rootObject, i.JSONPointer)
	if err != nil {
		return nil, err
	}
	return current, nil
}

func (i *IntentSignupFlowStepCreateAuthenticator) findAuthenticatorOfSameType(ctx context.Context, deps *authflow.Dependencies, typ model.AuthenticatorType) (*authenticator.Info, error) {

	userAuthns, err := deps.Authenticators.List(ctx, i.UserID)
	if err != nil {
		return nil, err
	}

	var existing *authenticator.Info

	for _, uAuthn := range userAuthns {
		uAuthn := uAuthn
		if uAuthn.Type == typ {
			existing = uAuthn

		}
	}

	return existing, nil
}

func (i *IntentSignupFlowStepCreateAuthenticator) getPublicOptions() ([]CreateAuthenticatorOption, error) {
	return slice.Map(i.Options, func(o CreateAuthenticatorOptionInternal) CreateAuthenticatorOption {
		return o.CreateAuthenticatorOption
	}), nil
}

func (i *IntentSignupFlowStepCreateAuthenticator) reactToExistingAuthenticator(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, option CreateAuthenticatorOptionInternal, authn *authenticator.Info, idx int) (*authflow.Node, error) {
	if len(flows.Nearest.Nodes) == 0 {
		return authflow.NewNodeSimple(&NodeSkipCreationByExistingAuthenticator{
			Authenticator:  authn,
			JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
			Authentication: option.Authentication,
		}), nil
	}

	_, _, authenticatorCreated := authflow.FindMilestoneInCurrentFlow[MilestoneFlowCreateAuthenticator](flows)
	_, _, nestedStepsHandled := authflow.FindMilestoneInCurrentFlow[MilestoneNestedSteps](flows)

	current, err := i.currentFlowObject(deps, flows, i)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	switch {
	case authenticatorCreated && !nestedStepsHandled:
		authentication := i.authenticationMethod(flows)
		return authflow.NewSubFlow(&IntentSignupFlowSteps{
			FlowReference:          i.FlowReference,
			JSONPointer:            i.jsonPointer(step, authentication),
			UserID:                 i.UserID,
			IsUpdatingExistingUser: i.IsUpdatingExistingUser,
		}), nil
	default:
		return nil, authflow.ErrIncompatibleInput
	}
}

func (i *IntentSignupFlowStepCreateAuthenticator) findSkippableOption(
	ctx context.Context,
	deps *authflow.Dependencies,
	flows authflow.Flows) (option *CreateAuthenticatorOptionInternal, idx int, info *authenticator.Info, err error) {
	userAuthns, err := deps.Authenticators.List(ctx, i.UserID)
	if err != nil {
		return nil, -1, nil, err
	}
	// For each option, see if any existing identities can be reused
	for idx, option := range i.Options {
		option := option
		existingAuthn := i.findAuthenticatorByOption(userAuthns, option)
		if existingAuthn != nil {
			return &option, idx, existingAuthn, nil
		}
	}
	return nil, -1, nil, nil
}

func (i *IntentSignupFlowStepCreateAuthenticator) findAuthenticatorByOption(in []*authenticator.Info, option CreateAuthenticatorOptionInternal) *authenticator.Info {

	switch option.Authentication {
	case model.AuthenticationFlowAuthenticationPrimaryPassword:
		return findPassword(in, authenticator.KindPrimary)
	case model.AuthenticationFlowAuthenticationSecondaryPassword:
		return findPassword(in, authenticator.KindSecondary)
	case model.AuthenticationFlowAuthenticationPrimaryPasskey:
		return findPrimaryPasskey(in, authenticator.KindPrimary)
	case model.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
		return findEmailOOB(in, authenticator.KindPrimary, option.UnmaskedTarget)
	case model.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
		return findEmailOOB(in, authenticator.KindSecondary, option.UnmaskedTarget)
	case model.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
		return findSMSOOB(in, authenticator.KindPrimary, option.UnmaskedTarget)
	case model.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
		return findSMSOOB(in, authenticator.KindSecondary, option.UnmaskedTarget)
	case model.AuthenticationFlowAuthenticationSecondaryTOTP:
		return findTOTP(in, authenticator.KindSecondary)
	}
	return nil
}
