package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentReauthFlowStepAuthenticate{})
}

// IntentReauthFlowStepAuthenticate
//
//   IntentUseAuthenticatorPassword (MilestoneFlowAuthenticate)
//     NodeDoUseAuthenticatorPassword (MilestoneDidAuthenticate)
//
//   IntentUseAuthenticatorPasskey (MilestoneFlowAuthenticate)
//     NodeDoUseAuthenticatorPasskey (MilestoneDidAuthenticate)
//
//   IntentUseAuthenticatorOOBOTP (MilestoneFlowAuthenticate)
//     NodeDoUseAuthenticatorSimple (MilestoneDidAuthenticate)
//
//   IntentUseAuthenticatorTOTP (MilestoneFlowAuthenticate)
//     NodeDoUseAuthenticatorSimple (MilestoneDidAuthenticate)

type IntentReauthFlowStepAuthenticate struct {
	FlowReference authflow.FlowReference `json:"flow_reference,omitempty"`
	JSONPointer   jsonpointer.T          `json:"json_pointer,omitempty"`
	StepName      string                 `json:"step_name,omitempty"`
	UserID        string                 `json:"user_id,omitempty"`
	Options       []AuthenticateOption   `json:"options"`
}

var _ authflow.Intent = &IntentReauthFlowStepAuthenticate{}
var _ authflow.DataOutputer = &IntentReauthFlowStepAuthenticate{}
var _ authflow.Milestone = &IntentReauthFlowStepAuthenticate{}

func NewIntentReauthFlowStepAuthenticate(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, i *IntentReauthFlowStepAuthenticate, originNode authflow.NodeOrIntent) (*IntentReauthFlowStepAuthenticate, error) {
	current, err := i.currentFlowObject(deps, flows, originNode)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	options, err := getAuthenticationOptionsForReauth(ctx, deps, flows, i.UserID, step)
	if err != nil {
		return nil, err
	}

	i.Options = options

	return i, nil
}

func (i *IntentReauthFlowStepAuthenticate) Milestone() {
}

func (*IntentReauthFlowStepAuthenticate) Kind() string {
	return "IntentReauthFlowStepAuthenticate"
}

func (i *IntentReauthFlowStepAuthenticate) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	authenticationMethodSelected := false
	mFlowSelect, mFlowSelectFlows, ok := authflow.FindMilestoneInCurrentFlow[MilestoneFlowSelectAuthenticationMethod](flows)
	if ok {
		_, _, authenticationMethodSelected = mFlowSelect.MilestoneFlowSelectAuthenticationMethod(mFlowSelectFlows)
	}

	authenticated := false
	mFlowAuthenticate, mFlowAuthenticateFlows, ok := authflow.FindMilestoneInCurrentFlow[MilestoneFlowAuthenticate](flows)
	if ok {
		_, _, authenticated = mFlowAuthenticate.MilestoneFlowAuthenticate(mFlowAuthenticateFlows)
	}

	_, _, nestedStepsHandled := authflow.FindMilestoneInCurrentFlow[MilestoneNestedSteps](flows)

	switch {
	case !authenticationMethodSelected:
		if len(i.Options) == 0 {
			// This step is NON-optional but have no options
			return nil, api.ErrNoAuthenticator
		}

		flowRootObject, err := findNearestFlowObjectInFlow(deps, flows, i)
		if err != nil {
			return nil, err
		}

		shouldBypassBotProtection := ShouldExistingResultBypassBotProtectionRequirement(ctx)

		// Let the input to select which authentication method to use.
		return &InputSchemaReauthFlowStepAuthenticate{
			FlowRootObject:            flowRootObject,
			JSONPointer:               i.JSONPointer,
			Options:                   i.Options,
			ShouldBypassBotProtection: shouldBypassBotProtection,
			BotProtectionCfg:          deps.Config.BotProtection,
		}, nil
	case !authenticated:
		// This branch is only reached when there is a programming error.
		// We expect the selected authentication method to be authenticated before this intent becomes input reactor again.
		panic(fmt.Errorf("unauthenticated"))

	case !nestedStepsHandled:
		// Handle nested steps.
		return nil, nil
	default:
		return nil, authflow.ErrEOF
	}
}

func (i *IntentReauthFlowStepAuthenticate) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	current, err := i.currentFlowObject(deps, flows, i)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	authenticationMethodSelected := false
	mFlowSelect, mFlowSelectFlows, ok := authflow.FindMilestoneInCurrentFlow[MilestoneFlowSelectAuthenticationMethod](flows)
	if ok {
		_, _, authenticationMethodSelected = mFlowSelect.MilestoneFlowSelectAuthenticationMethod(mFlowSelectFlows)
	}

	authenticated := false
	mFlowAuthenticate, mFlowAuthenticateFlows, ok := authflow.FindMilestoneInCurrentFlow[MilestoneFlowAuthenticate](flows)
	if ok {
		_, _, authenticated = mFlowAuthenticate.MilestoneFlowAuthenticate(mFlowAuthenticateFlows)
	}

	_, _, nestedStepsHandled := authflow.FindMilestoneInCurrentFlow[MilestoneNestedSteps](flows)

	switch {
	case !authenticationMethodSelected:
		var inputTakeAuthenticationMethod inputTakeAuthenticationMethod
		if authflow.AsInput(input, &inputTakeAuthenticationMethod) {
			authentication := inputTakeAuthenticationMethod.GetAuthenticationMethod()

			idx, err := i.getIndex(step, authentication)
			if err != nil {
				return nil, err
			}

			switch authentication {
			case model.AuthenticationFlowAuthenticationPrimaryPassword:
				fallthrough
			case model.AuthenticationFlowAuthenticationSecondaryPassword:
				return authflow.NewSubFlow(&IntentUseAuthenticatorPassword{
					JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
					UserID:         i.UserID,
					Authentication: authentication,
				}), nil
			case model.AuthenticationFlowAuthenticationPrimaryPasskey:
				return authflow.NewSubFlow(&IntentUseAuthenticatorPasskey{
					JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
					UserID:         i.UserID,
					Authentication: authentication,
				}), nil
			case model.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
				fallthrough
			case model.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
				fallthrough
			case model.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
				fallthrough
			case model.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
				return authflow.NewSubFlow(&IntentUseAuthenticatorOOBOTP{
					JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
					UserID:         i.UserID,
					Authentication: authentication,
					Options:        i.Options,
				}), nil
			case model.AuthenticationFlowAuthenticationSecondaryTOTP:
				return authflow.NewSubFlow(&IntentUseAuthenticatorTOTP{
					JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
					UserID:         i.UserID,
					Authentication: authentication,
				}), nil
			}
		}

		return nil, authflow.ErrIncompatibleInput
	case !authenticated:
		panic(fmt.Errorf("unauthenticated"))
	case !nestedStepsHandled:
		authentication := i.authenticationMethod(flows)
		return authflow.NewSubFlow(&IntentReauthFlowSteps{
			FlowReference: i.FlowReference,
			JSONPointer:   i.jsonPointer(step, authentication),
		}), nil
	default:
		return nil, authflow.ErrIncompatibleInput
	}
}

func (i *IntentReauthFlowStepAuthenticate) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	var options []AuthenticateOptionForOutput
	for _, o := range i.Options {
		options = append(options, o.ToOutput(ctx))
	}

	return NewStepAuthenticateData(StepAuthenticateData{
		Options: options,
	}), nil
}

func (i *IntentReauthFlowStepAuthenticate) getIndex(step *config.AuthenticationFlowReauthFlowStep, am model.AuthenticationFlowAuthentication) (idx int, err error) {
	idx = -1

	allAllowed := i.getAllAllowed(step)

	for index := range allAllowed {
		thisMethod := allAllowed[index]
		for _, option := range i.Options {
			if thisMethod == option.Authentication && thisMethod == am {
				idx = index
			}
		}
	}

	if idx >= 0 {
		return
	}

	err = authflow.ErrIncompatibleInput
	return
}

func (*IntentReauthFlowStepAuthenticate) getAllAllowed(step *config.AuthenticationFlowReauthFlowStep) []model.AuthenticationFlowAuthentication {
	// Make empty slice.
	allAllowed := []model.AuthenticationFlowAuthentication{}

	for _, branch := range step.OneOf {
		branch := branch
		allAllowed = append(allAllowed, branch.Authentication)
	}

	return allAllowed
}

func (*IntentReauthFlowStepAuthenticate) step(o config.AuthenticationFlowObject) *config.AuthenticationFlowReauthFlowStep {
	step, ok := o.(*config.AuthenticationFlowReauthFlowStep)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}

	return step
}

func (*IntentReauthFlowStepAuthenticate) authenticationMethod(flows authflow.Flows) model.AuthenticationFlowAuthentication {
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

func (i *IntentReauthFlowStepAuthenticate) jsonPointer(step *config.AuthenticationFlowReauthFlowStep, am model.AuthenticationFlowAuthentication) jsonpointer.T {
	for idx, branch := range step.OneOf {
		branch := branch
		if branch.Authentication == am {
			return authflow.JSONPointerForOneOf(i.JSONPointer, idx)
		}
	}

	panic(fmt.Errorf("selected authentication method is not allowed"))
}

func (i *IntentReauthFlowStepAuthenticate) currentFlowObject(deps *authflow.Dependencies, flows authflow.Flows, origin authflow.NodeOrIntent) (config.AuthenticationFlowObject, error) {
	rootObject, err := findNearestFlowObjectInFlow(deps, flows, origin)
	if err != nil {
		return nil, err
	}
	current, err := authflow.FlowObject(rootObject, i.JSONPointer)
	if err != nil {
		return nil, err
	}
	return current, nil
}
