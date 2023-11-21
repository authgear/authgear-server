package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type IntentLoginFlowStepAuthenticateTarget interface {
	GetIdentity(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) *identity.Info
}

func init() {
	authflow.RegisterIntent(&IntentLoginFlowStepAuthenticate{})
}

type IntentLoginFlowStepAuthenticate struct {
	JSONPointer jsonpointer.T        `json:"json_pointer,omitempty"`
	StepName    string               `json:"step_name,omitempty"`
	UserID      string               `json:"user_id,omitempty"`
	Options     []AuthenticateOption `json:"options"`
}

var _ authflow.TargetStep = &IntentLoginFlowStepAuthenticate{}

func (i *IntentLoginFlowStepAuthenticate) GetName() string {
	return i.StepName
}

func (i *IntentLoginFlowStepAuthenticate) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

var _ IntentLoginFlowStepChangePasswordTarget = &IntentLoginFlowStepAuthenticate{}

func (*IntentLoginFlowStepAuthenticate) GetPasswordAuthenticator(_ context.Context, _ *authflow.Dependencies, flows authflow.Flows) (info *authenticator.Info, ok bool) {
	m, ok := authflow.FindMilestone[MilestoneDoUseAuthenticatorPassword](flows.Nearest)
	if !ok {
		return
	}

	ok = false
	n := m.MilestoneDoUseAuthenticatorPassword()

	if n.PasswordChangeRequired {
		info = n.Authenticator
		ok = true
	}

	return
}

var _ authflow.Intent = &IntentLoginFlowStepAuthenticate{}
var _ authflow.DataOutputer = &IntentLoginFlowStepAuthenticate{}

func NewIntentLoginFlowStepAuthenticate(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, i *IntentLoginFlowStepAuthenticate) (*IntentLoginFlowStepAuthenticate, error) {
	current, err := authflow.FlowObject(authflow.GetFlowRootObject(ctx), i.JSONPointer)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	options, err := getAuthenticationOptionsForLogin(ctx, deps, flows, i.UserID, step)
	if err != nil {
		return nil, err
	}

	i.Options = options
	return i, nil
}

func (*IntentLoginFlowStepAuthenticate) Kind() string {
	return "IntentLoginFlowStepAuthenticate"
}

func (i *IntentLoginFlowStepAuthenticate) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	current, err := authflow.FlowObject(authflow.GetFlowRootObject(ctx), i.JSONPointer)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	deviceTokenIndex := i.deviceTokenIndex(step)
	deviceTokenEnabled := deviceTokenIndex >= 0

	_, deviceTokenInspected := authflow.FindMilestone[MilestoneDeviceTokenInspected](flows.Nearest)

	_, authenticationMethodSelected := authflow.FindMilestone[MilestoneAuthenticationMethod](flows.Nearest)

	_, authenticated := authflow.FindMilestone[MilestoneDidAuthenticate](flows.Nearest)

	_, deviceTokenCreatedIfRequested := authflow.FindMilestone[MilestoneDoCreateDeviceTokenIfRequested](flows.Nearest)

	_, nestedStepsHandled := authflow.FindMilestone[MilestoneNestedSteps](flows.Nearest)

	switch {
	case deviceTokenEnabled && !deviceTokenInspected:
		// Inspect the device token
		return nil, nil
	case !authenticationMethodSelected:
		if len(i.Options) == 0 {
			if step.IsOptional() {
				// Skip this step and any nested step.
				return nil, authflow.ErrEOF
			}

			// Otherwise this step is NON-optional but have no options
			return nil, api.ErrNoAuthenticator
		}

		// Let the input to select which authentication method to use.
		return &InputSchemaLoginFlowStepAuthenticate{
			JSONPointer:        i.JSONPointer,
			Options:            i.Options,
			DeviceTokenEnabled: deviceTokenEnabled,
		}, nil
	case !authenticated:
		// This branch is only reached when there is a programming error.
		// We expect the selected authentication method to be authenticated before this intent becomes input reactor again.
		panic(fmt.Errorf("unauthenticated"))

	case deviceTokenEnabled && !deviceTokenCreatedIfRequested:
		// We look at the current input to see if device token is request.
		// So we do not need to take another input.
		return nil, nil
	case !nestedStepsHandled:
		// Handle nested steps.
		return nil, nil
	default:
		return nil, authflow.ErrEOF
	}
}

func (i *IntentLoginFlowStepAuthenticate) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	current, err := authflow.FlowObject(authflow.GetFlowRootObject(ctx), i.JSONPointer)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	deviceTokenIndex := i.deviceTokenIndex(step)
	deviceTokenEnabled := deviceTokenIndex >= 0

	_, deviceTokenInspected := authflow.FindMilestone[MilestoneDeviceTokenInspected](flows.Nearest)

	_, authenticationMethodSelected := authflow.FindMilestone[MilestoneAuthenticationMethod](flows.Nearest)

	_, authenticated := authflow.FindMilestone[MilestoneDidAuthenticate](flows.Nearest)

	_, deviceTokenCreatedIfRequested := authflow.FindMilestone[MilestoneDoCreateDeviceTokenIfRequested](flows.Nearest)

	_, nestedStepsHandled := authflow.FindMilestone[MilestoneNestedSteps](flows.Nearest)

	switch {
	case deviceTokenEnabled && !deviceTokenInspected:
		return authflow.NewSubFlow(&IntentInspectDeviceToken{
			UserID: i.UserID,
		}), nil

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
				return authflow.NewNodeSimple(&NodeUseAuthenticatorPassword{
					JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
					UserID:         i.UserID,
					Authentication: authentication,
				}), nil
			case config.AuthenticationFlowAuthenticationPrimaryPasskey:
				return authflow.NewNodeSimple(&NodeUseAuthenticatorPasskey{
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
				return authflow.NewNodeSimple(&NodeUseAuthenticatorTOTP{
					JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
					UserID:         i.UserID,
					Authentication: authentication,
				}), nil
			case config.AuthenticationFlowAuthenticationRecoveryCode:
				return authflow.NewNodeSimple(&NodeUseRecoveryCode{
					JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
					UserID:         i.UserID,
					Authentication: authentication,
				}), nil
			case config.AuthenticationFlowAuthenticationDeviceToken:
				// Device token is handled transparently.
				return nil, authflow.ErrIncompatibleInput
			}
		}

		return nil, authflow.ErrIncompatibleInput
	case !authenticated:
		panic(fmt.Errorf("unauthenticated"))
	case deviceTokenEnabled && !deviceTokenCreatedIfRequested:
		return authflow.NewSubFlow(&IntentCreateDeviceTokenIfRequested{
			JSONPointer: authflow.JSONPointerForOneOf(i.JSONPointer, deviceTokenIndex),
			UserID:      i.UserID,
		}), nil
	case !nestedStepsHandled:
		authentication := i.authenticationMethod(flows)
		return authflow.NewSubFlow(&IntentLoginFlowSteps{
			JSONPointer: i.jsonPointer(step, authentication),
		}), nil
	default:
		return nil, authflow.ErrIncompatibleInput
	}
}

func (i *IntentLoginFlowStepAuthenticate) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	current, err := authflow.FlowObject(authflow.GetFlowRootObject(ctx), i.JSONPointer)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	deviceTokenIndex := i.deviceTokenIndex(step)
	deviceTokenEnabled := deviceTokenIndex >= 0

	options := []AuthenticateOptionForOutput{}
	for _, o := range i.Options {
		options = append(options, o.ToOutput())
	}

	return stepAuthenticateData{
		Options:            options,
		DeviceTokenEnabled: deviceTokenEnabled,
	}, nil
}

func (i *IntentLoginFlowStepAuthenticate) getIndex(step *config.AuthenticationFlowLoginFlowStep, am config.AuthenticationFlowAuthentication) (idx int, err error) {
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

func (*IntentLoginFlowStepAuthenticate) getAllAllowed(step *config.AuthenticationFlowLoginFlowStep) []config.AuthenticationFlowAuthentication {
	// Make empty slice.
	allAllowed := []config.AuthenticationFlowAuthentication{}

	for _, branch := range step.OneOf {
		branch := branch
		allAllowed = append(allAllowed, branch.Authentication)
	}

	return allAllowed
}

func (i *IntentLoginFlowStepAuthenticate) deviceTokenIndex(step *config.AuthenticationFlowLoginFlowStep) int {
	allAllowed := i.getAllAllowed(step)
	for idx, am := range allAllowed {
		if am == config.AuthenticationFlowAuthenticationDeviceToken {
			return idx
		}
	}
	return -1
}

func (*IntentLoginFlowStepAuthenticate) step(o config.AuthenticationFlowObject) *config.AuthenticationFlowLoginFlowStep {
	step, ok := o.(*config.AuthenticationFlowLoginFlowStep)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}

	return step
}

func (*IntentLoginFlowStepAuthenticate) authenticationMethod(flows authflow.Flows) config.AuthenticationFlowAuthentication {
	m, ok := authflow.FindMilestone[MilestoneAuthenticationMethod](flows.Nearest)
	if !ok {
		panic(fmt.Errorf("authentication method not yet selected"))
	}

	am := m.MilestoneAuthenticationMethod()

	return am
}

func (i *IntentLoginFlowStepAuthenticate) jsonPointer(step *config.AuthenticationFlowLoginFlowStep, am config.AuthenticationFlowAuthentication) jsonpointer.T {
	for idx, branch := range step.OneOf {
		branch := branch
		if branch.Authentication == am {
			return authflow.JSONPointerForOneOf(i.JSONPointer, idx)
		}
	}

	panic(fmt.Errorf("selected authentication method is not allowed"))
}
