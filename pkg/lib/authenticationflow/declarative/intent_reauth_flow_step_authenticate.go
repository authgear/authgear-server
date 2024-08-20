package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api"
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

func NewIntentReauthFlowStepAuthenticate(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, i *IntentReauthFlowStepAuthenticate) (*IntentReauthFlowStepAuthenticate, error) {
	current, err := i.currentFlowObject(deps)
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

		flowRootObject, err := flowRootObject(deps, i.FlowReference)
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

func (i *IntentReauthFlowStepAuthenticate) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	current, err := i.currentFlowObject(deps)
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
			case config.AuthenticationFlowAuthenticationPrimaryPassword:
				fallthrough
			case config.AuthenticationFlowAuthenticationSecondaryPassword:
				return authflow.NewSubFlow(&IntentUseAuthenticatorPassword{
					JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
					UserID:         i.UserID,
					Authentication: authentication,
				}), nil
			case config.AuthenticationFlowAuthenticationPrimaryPasskey:
				return authflow.NewSubFlow(&IntentUseAuthenticatorPasskey{
					JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
					UserID:         i.UserID,
					Authentication: authentication,
				}), nil
			case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
				fallthrough
			case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
				fallthrough
			case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
				fallthrough
			case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
				return authflow.NewSubFlow(&IntentUseAuthenticatorOOBOTP{
					JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
					UserID:         i.UserID,
					Authentication: authentication,
					Options:        i.Options,
				}), nil
			case config.AuthenticationFlowAuthenticationSecondaryTOTP:
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

func (i *IntentReauthFlowStepAuthenticate) getIndex(step *config.AuthenticationFlowReauthFlowStep, am config.AuthenticationFlowAuthentication) (idx int, err error) {
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

func (*IntentReauthFlowStepAuthenticate) getAllAllowed(step *config.AuthenticationFlowReauthFlowStep) []config.AuthenticationFlowAuthentication {
	// Make empty slice.
	allAllowed := []config.AuthenticationFlowAuthentication{}

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

func (*IntentReauthFlowStepAuthenticate) authenticationMethod(flows authflow.Flows) config.AuthenticationFlowAuthentication {
	m, mFlows, ok := authflow.FindMilestoneInCurrentFlow[MilestoneFlowSelectAuthenticationMethod](flows)
	if !ok {
		panic(fmt.Errorf("authentication method not yet selected"))
	}

	mDidSelect, _, _ := m.MilestoneFlowSelectAuthenticationMethod(mFlows)

	return mDidSelect.MilestoneDidSelectAuthenticationMethod()
}

func (i *IntentReauthFlowStepAuthenticate) jsonPointer(step *config.AuthenticationFlowReauthFlowStep, am config.AuthenticationFlowAuthentication) jsonpointer.T {
	for idx, branch := range step.OneOf {
		branch := branch
		if branch.Authentication == am {
			return authflow.JSONPointerForOneOf(i.JSONPointer, idx)
		}
	}

	panic(fmt.Errorf("selected authentication method is not allowed"))
}

func (i *IntentReauthFlowStepAuthenticate) currentFlowObject(deps *authflow.Dependencies) (config.AuthenticationFlowObject, error) {
	rootObject, err := flowRootObject(deps, i.FlowReference)
	if err != nil {
		return nil, err
	}
	current, err := authflow.FlowObject(rootObject, i.JSONPointer)
	if err != nil {
		return nil, err
	}
	return current, nil
}
