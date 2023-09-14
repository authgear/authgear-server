package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type IntentSignupFlowStepAuthenticateTarget interface {
	GetOOBOTPClaims(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (map[model.ClaimName]string, error)
}

func init() {
	authflow.RegisterIntent(&IntentSignupFlowStepAuthenticate{})
}

type IntentSignupFlowStepAuthenticateData struct {
	PasswordPolicy *PasswordPolicy `json:"password_policy,omitempty"`
}

var _ authflow.Data = &IntentSignupFlowStepAuthenticateData{}

func (m IntentSignupFlowStepAuthenticateData) Data() {}

type IntentSignupFlowStepAuthenticate struct {
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
	StepID      string        `json:"step_id,omitempty"`
	UserID      string        `json:"user_id,omitempty"`
}

var _ FlowStep = &IntentSignupFlowStepAuthenticate{}

func (i *IntentSignupFlowStepAuthenticate) GetID() string {
	return i.StepID
}

func (i *IntentSignupFlowStepAuthenticate) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

var _ IntentSignupFlowStepVerifyTarget = &IntentSignupFlowStepAuthenticate{}

func (*IntentSignupFlowStepAuthenticate) GetVerifiableClaims(_ context.Context, _ *authflow.Dependencies, flows authflow.Flows) (map[model.ClaimName]string, error) {
	m, ok := authflow.FindMilestone[MilestoneDoCreateAuthenticator](flows.Nearest)
	if !ok {
		return nil, fmt.Errorf("MilestoneDoCreateAuthenticator cannot be found in IntentSignupFlowStepAuthenticate")
	}

	info := m.MilestoneDoCreateAuthenticator()

	return info.StandardClaims(), nil
}

func (*IntentSignupFlowStepAuthenticate) GetPurpose(_ context.Context, _ *authflow.Dependencies, _ authflow.Flows) otp.Purpose {
	return otp.PurposeOOBOTP
}

func (i *IntentSignupFlowStepAuthenticate) GetMessageType(_ context.Context, _ *authflow.Dependencies, flows authflow.Flows) otp.MessageType {
	authenticationMethod := i.authenticationMethod(flows)
	switch authenticationMethod {
	case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
		return otp.MessageTypeSetupPrimaryOOB
	case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
		return otp.MessageTypeSetupPrimaryOOB
	case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
		return otp.MessageTypeSetupSecondaryOOB
	case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
		return otp.MessageTypeSetupSecondaryOOB
	default:
		panic(fmt.Errorf("unexpected authentication method: %v", authenticationMethod))
	}
}

var _ authflow.Intent = &IntentSignupFlowStepAuthenticate{}
var _ authflow.DataOutputer = &IntentSignupFlowStepAuthenticate{}

func (*IntentSignupFlowStepAuthenticate) Kind() string {
	return "IntentSignupFlowStepAuthenticate"
}

func (i *IntentSignupFlowStepAuthenticate) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	// Let the input to select which authentication method to use.
	if len(flows.Nearest.Nodes) == 0 {
		current, err := authflow.FlowObject(authflow.GetFlowRootObject(ctx), i.JSONPointer)
		if err != nil {
			return nil, err
		}
		step := i.step(current)
		return &InputSchemaSignupFlowStepAuthenticate{
			JSONPointer: i.JSONPointer,
			OneOf:       step.OneOf,
		}, nil
	}

	_, authenticatorCreated := authflow.FindMilestone[MilestoneDoCreateAuthenticator](flows.Nearest)
	_, nestedStepsHandled := authflow.FindMilestone[MilestoneNestedSteps](flows.Nearest)

	switch {
	case authenticatorCreated && !nestedStepsHandled:
		// Handle nested steps.
		return nil, nil
	default:
		return nil, authflow.ErrEOF
	}
}

func (i *IntentSignupFlowStepAuthenticate) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	current, err := authflow.FlowObject(authflow.GetFlowRootObject(ctx), i.JSONPointer)
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
			case config.AuthenticationFlowAuthenticationPrimaryPassword:
				fallthrough
			case config.AuthenticationFlowAuthenticationSecondaryPassword:
				return authflow.NewNodeSimple(&NodeCreateAuthenticatorPassword{
					JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
					UserID:         i.UserID,
					Authentication: authentication,
				}), nil
			case config.AuthenticationFlowAuthenticationPrimaryPasskey:
				// Cannot create passkey in this step.
				return nil, authflow.ErrIncompatibleInput
			case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
				fallthrough
			case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
				fallthrough
			case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
				fallthrough
			case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
				return authflow.NewNodeSimple(&NodeCreateAuthenticatorOOBOTP{
					JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
					UserID:         i.UserID,
					Authentication: authentication,
				}), nil
			case config.AuthenticationFlowAuthenticationSecondaryTOTP:
				node, err := NewNodeCreateAuthenticatorTOTP(deps, &NodeCreateAuthenticatorTOTP{
					JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
					UserID:         i.UserID,
					Authentication: authentication,
				})
				if err != nil {
					return nil, err
				}
				return authflow.NewNodeSimple(node), nil
			}
		}
		return nil, authflow.ErrIncompatibleInput
	}

	_, authenticatorCreated := authflow.FindMilestone[MilestoneDoCreateAuthenticator](flows.Nearest)
	_, nestedStepsHandled := authflow.FindMilestone[MilestoneNestedSteps](flows.Nearest)

	switch {
	case authenticatorCreated && !nestedStepsHandled:
		authentication := i.authenticationMethod(flows)
		return authflow.NewSubFlow(&IntentSignupFlowSteps{
			JSONPointer: i.jsonPointer(step, authentication),
			UserID:      i.UserID,
		}), nil
	default:
		return nil, authflow.ErrIncompatibleInput
	}
}

func (i *IntentSignupFlowStepAuthenticate) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	return IntentSignupFlowStepAuthenticateData{
		PasswordPolicy: NewPasswordPolicy(deps.Config.Authenticator.Password.Policy),
	}, nil
}

func (*IntentSignupFlowStepAuthenticate) step(o config.AuthenticationFlowObject) *config.AuthenticationFlowSignupFlowStep {
	step, ok := o.(*config.AuthenticationFlowSignupFlowStep)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}

	return step
}

func (i *IntentSignupFlowStepAuthenticate) checkAuthenticationMethod(deps *authflow.Dependencies, step *config.AuthenticationFlowSignupFlowStep, am config.AuthenticationFlowAuthentication) (idx int, err error) {
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

func (*IntentSignupFlowStepAuthenticate) authenticationMethod(flows authflow.Flows) config.AuthenticationFlowAuthentication {
	m, ok := authflow.FindMilestone[MilestoneAuthenticationMethod](flows.Nearest)
	if !ok {
		panic(fmt.Errorf("authentication method not yet selected"))
	}

	am := m.MilestoneAuthenticationMethod()

	return am
}

func (i *IntentSignupFlowStepAuthenticate) jsonPointer(step *config.AuthenticationFlowSignupFlowStep, am config.AuthenticationFlowAuthentication) jsonpointer.T {
	for idx, branch := range step.OneOf {
		branch := branch
		if branch.Authentication == am {
			return authflow.JSONPointerForOneOf(i.JSONPointer, idx)
		}
	}

	panic(fmt.Errorf("selected identification method is not allowed"))
}
