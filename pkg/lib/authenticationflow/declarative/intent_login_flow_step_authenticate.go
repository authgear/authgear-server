package declarative

import (
	"context"
	"fmt"
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
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

type IntentLoginFlowStepAuthenticateData struct {
	Candidates []UseAuthenticationCandidate `json:"candidates"`
}

var _ authflow.Data = IntentLoginFlowStepAuthenticateData{}

func (m IntentLoginFlowStepAuthenticateData) Data() {}

type IntentLoginFlowStepAuthenticate struct {
	LoginFlow   string        `json:"login_flow,omitempty"`
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
	StepID      string        `json:"step_id,omitempty"`
	UserID      string        `json:"user_id,omitempty"`
}

var _ FlowStep = &IntentLoginFlowStepAuthenticate{}

func (i *IntentLoginFlowStepAuthenticate) GetID() string {
	return i.StepID
}

func (i *IntentLoginFlowStepAuthenticate) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

var _ IntentLoginFlowStepChangePasswordTarget = &IntentLoginFlowStepAuthenticate{}

func (*IntentLoginFlowStepAuthenticate) GetPasswordAuthenticator(_ context.Context, _ *authflow.Dependencies, flows authflow.Flows) (info *authenticator.Info, ok bool) {
	m, ok := authflow.FindMilestone[MilestoneDidVerifyAuthenticator](flows.Nearest)
	if !ok {
		return
	}

	ok = false
	n := m.MilestoneDidVerifyAuthenticator()

	if n.Authenticator.Type == model.AuthenticatorTypePassword {
		if n.PasswordChangeRequired {
			info = n.Authenticator
			ok = true
		}
	}

	return
}

var _ authflow.Intent = &IntentLoginFlowStepAuthenticate{}
var _ authflow.Boundary = &IntentLoginFlowStepAuthenticate{}
var _ authflow.DataOutputer = &IntentLoginFlowStepAuthenticate{}

func NewIntentLoginFlowStepAuthenticate(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, i *IntentLoginFlowStepAuthenticate) (*IntentLoginFlowStepAuthenticate, error) {
	// OutputData will include usable authenticators.
	// If it returns error, there is no usable authenticators.
	// This intent cannot proceed if there is no usable authenticators.
	// Therefore, we prevent from adding this intent to the flow if such case happens.
	_, err := i.OutputData(ctx, deps, flows)
	if err != nil {
		return nil, err
	}

	return i, nil
}

func (*IntentLoginFlowStepAuthenticate) Kind() string {
	return "IntentLoginFlowStepAuthenticate"
}

func (i *IntentLoginFlowStepAuthenticate) Boundary() string {
	return i.JSONPointer.String()
}

func (i *IntentLoginFlowStepAuthenticate) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	current, err := loginFlowCurrent(deps, i.LoginFlow, i.JSONPointer)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	deviceTokenEnabled := i.deviceTokenEnabled(step)

	candidates, err := getAuthenticationCandidatesForStep(ctx, deps, flows, i.UserID, step)
	if err != nil {
		return nil, err
	}

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
		// Let the input to select which authentication method to use.
		return &InputSchemaLoginFlowStepAuthenticate{
			Candidates:         candidates,
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
	current, err := loginFlowCurrent(deps, i.LoginFlow, i.JSONPointer)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	deviceTokenEnabled := i.deviceTokenEnabled(step)

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

			candidates, err := getAuthenticationCandidatesForStep(ctx, deps, flows, i.UserID, step)
			if err != nil {
				return nil, err
			}

			idx := i.getIndex(step, candidates, authentication)

			switch authentication {
			case config.AuthenticationFlowAuthenticationPrimaryPassword:
				fallthrough
			case config.AuthenticationFlowAuthenticationSecondaryPassword:
				return authflow.NewNodeSimple(&NodeUseAuthenticatorPassword{
					UserID:         i.UserID,
					Authentication: authentication,
				}), nil
			case config.AuthenticationFlowAuthenticationPrimaryPasskey:
				// FIXME(authflow): authenticate with passkey
			case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
				fallthrough
			case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
				fallthrough
			case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
				fallthrough
			case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
				return authflow.NewSubFlow(&IntentUseAuthenticatorOOBOTP{
					LoginFlow:         i.LoginFlow,
					JSONPointer:       JSONPointerForOneOf(i.JSONPointer, idx),
					JSONPointerToStep: i.JSONPointer,
					UserID:            i.UserID,
					Authentication:    authentication,
				}), nil
			case config.AuthenticationFlowAuthenticationSecondaryTOTP:
				return authflow.NewNodeSimple(&NodeUseAuthenticatorTOTP{
					UserID:         i.UserID,
					Authentication: authentication,
				}), nil
			case config.AuthenticationFlowAuthenticationRecoveryCode:
				return authflow.NewNodeSimple(&NodeUseRecoveryCode{
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
			UserID: i.UserID,
		}), nil
	case !nestedStepsHandled:
		authentication := i.authenticationMethod(flows)
		return authflow.NewSubFlow(&IntentLoginFlowSteps{
			LoginFlow:   i.LoginFlow,
			JSONPointer: i.jsonPointer(step, authentication),
		}), nil
	default:
		return nil, authflow.ErrIncompatibleInput
	}
}

func (i *IntentLoginFlowStepAuthenticate) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	current, err := loginFlowCurrent(deps, i.LoginFlow, i.JSONPointer)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	candidates, err := getAuthenticationCandidatesForStep(ctx, deps, flows, i.UserID, step)
	if err != nil {
		return nil, err
	}

	return IntentLoginFlowStepAuthenticateData{
		Candidates: candidates,
	}, nil
}

func (i *IntentLoginFlowStepAuthenticate) getIndex(step *config.AuthenticationFlowLoginFlowStep, candidates []UseAuthenticationCandidate, am config.AuthenticationFlowAuthentication) (idx int) {
	idx = -1

	allAllowed := i.getAllAllowed(step)

	for i := range allAllowed {
		thisMethod := allAllowed[i]
		for _, candidate := range candidates {
			if thisMethod == candidate.Authentication && thisMethod == am {
				idx = i
			}
		}
	}

	if idx >= 0 {
		return
	}

	panic(fmt.Errorf("the input schema should have ensured index can always be found"))
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

func (i *IntentLoginFlowStepAuthenticate) deviceTokenEnabled(step *config.AuthenticationFlowLoginFlowStep) bool {
	allAllowed := i.getAllAllowed(step)
	for _, am := range allAllowed {
		if am == config.AuthenticationFlowAuthenticationDeviceToken {
			return true
		}
	}
	return false
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
			return JSONPointerForOneOf(i.JSONPointer, idx)
		}
	}

	panic(fmt.Errorf("selected authentication method is not allowed"))
}
