package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentPromoteFlowStepIdentify{})
}

type intentPromoteFlowStepIdentifyData struct {
	Options []IdentificationOption `json:"options"`
}

var _ authflow.Data = intentPromoteFlowStepIdentifyData{}

func (intentPromoteFlowStepIdentifyData) Data() {}

type IntentPromoteFlowStepIdentify struct {
	JSONPointer jsonpointer.T          `json:"json_pointer,omitempty"`
	StepName    string                 `json:"step_name,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	Options     []IdentificationOption `json:"options"`
}

var _ authflow.TargetStep = &IntentPromoteFlowStepIdentify{}

func (i *IntentPromoteFlowStepIdentify) GetName() string {
	return i.StepName
}

func (i *IntentPromoteFlowStepIdentify) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

var _ IntentSignupFlowStepVerifyTarget = &IntentPromoteFlowStepIdentify{}

func (*IntentPromoteFlowStepIdentify) GetVerifiableClaims(_ context.Context, _ *authflow.Dependencies, flows authflow.Flows) (map[model.ClaimName]string, error) {
	m, ok := authflow.FindMilestone[MilestoneDoCreateIdentity](flows.Nearest)
	if !ok {
		return nil, fmt.Errorf("MilestoneDoCreateIdentity cannot be found in IntentPromoteFlowStepIdentify")
	}
	info := m.MilestoneDoCreateIdentity()

	return info.IdentityAwareStandardClaims(), nil
}

func (*IntentPromoteFlowStepIdentify) GetPurpose(_ context.Context, _ *authflow.Dependencies, _ authflow.Flows) otp.Purpose {
	return otp.PurposeVerification
}

func (*IntentPromoteFlowStepIdentify) GetMessageType(_ context.Context, _ *authflow.Dependencies, _ authflow.Flows) otp.MessageType {
	return otp.MessageTypeVerification
}

var _ IntentSignupFlowStepCreateAuthenticatorTarget = &IntentPromoteFlowStepIdentify{}

func (n *IntentPromoteFlowStepIdentify) GetOOBOTPClaims(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (map[model.ClaimName]string, error) {
	return n.GetVerifiableClaims(ctx, deps, flows)
}

var _ authflow.Intent = &IntentPromoteFlowStepIdentify{}
var _ authflow.DataOutputer = &IntentPromoteFlowStepIdentify{}

func NewIntentPromoteFlowStepIdentify(ctx context.Context, deps *authflow.Dependencies, i *IntentPromoteFlowStepIdentify) (*IntentPromoteFlowStepIdentify, error) {
	current, err := authflow.FlowObject(authflow.GetFlowRootObject(ctx), i.JSONPointer)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	options := []IdentificationOption{}
	for _, b := range step.OneOf {
		switch b.Identification {
		case config.AuthenticationFlowIdentificationEmail:
			fallthrough
		case config.AuthenticationFlowIdentificationPhone:
			fallthrough
		case config.AuthenticationFlowIdentificationUsername:
			c := NewIdentificationOptionLoginID(b.Identification)
			options = append(options, c)
		case config.AuthenticationFlowIdentificationOAuth:
			oauthOptions := NewIdentificationOptionsOAuth(
				deps.Config.Identity.OAuth,
				deps.FeatureConfig.Identity.OAuth.Providers,
			)
			options = append(options, oauthOptions...)
		case config.AuthenticationFlowIdentificationPasskey:
			// Do not support create passkey in signup because
			// passkey is not considered as a persistent identifier.
			break
		}
	}

	i.Options = options
	return i, nil
}

func (*IntentPromoteFlowStepIdentify) Kind() string {
	return "IntentPromoteFlowStepIdentify"
}

func (i *IntentPromoteFlowStepIdentify) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	// Let the input to select which identification method to use.
	if len(flows.Nearest.Nodes) == 0 {
		return &InputSchemaStepIdentify{
			JSONPointer: i.JSONPointer,
			Options:     i.Options,
		}, nil
	}

	_, identityCreated := authflow.FindMilestone[MilestoneDoCreateIdentity](flows.Nearest)
	_, standardAttributesPopulated := authflow.FindMilestone[MilestoneDoPopulateStandardAttributes](flows.Nearest)
	_, nestedStepHandled := authflow.FindMilestone[MilestoneNestedSteps](flows.Nearest)

	switch {
	case identityCreated && !standardAttributesPopulated && !nestedStepHandled:
		// Populate standard attributes
		return nil, nil
	case identityCreated && standardAttributesPopulated && !nestedStepHandled:
		// Handle nested steps.
		return nil, nil
	default:
		return nil, authflow.ErrEOF
	}
}

func (i *IntentPromoteFlowStepIdentify) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	current, err := authflow.FlowObject(authflow.GetFlowRootObject(ctx), i.JSONPointer)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	if len(flows.Nearest.Nodes) == 0 {
		var inputTakeIdentificationMethod inputTakeIdentificationMethod

		if authflow.AsInput(input, &inputTakeIdentificationMethod) {

			identification := inputTakeIdentificationMethod.GetIdentificationMethod()
			idx, err := i.checkIdentificationMethod(deps, step, identification)
			if err != nil {
				return nil, err
			}

			syntheticInput := &InputStepIdentify{
				Identification: identification,
			}

			switch identification {
			case config.AuthenticationFlowIdentificationEmail:
				fallthrough
			case config.AuthenticationFlowIdentificationPhone:
				fallthrough
			case config.AuthenticationFlowIdentificationUsername:
				return authflow.NewNodeSimple(&NodePromoteIdentityLoginID{
					JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
					UserID:         i.UserID,
					Identification: identification,
					SyntheticInput: syntheticInput,
				}), nil
			case config.AuthenticationFlowIdentificationOAuth:
				return authflow.NewSubFlow(&IntentPromoteIdentityOAuth{
					JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
					UserID:         i.UserID,
					Identification: identification,
					SyntheticInput: syntheticInput,
				}), nil
			case config.AuthenticationFlowIdentificationPasskey:
				// Cannot create passkey in this step.
				return nil, authflow.ErrIncompatibleInput
			}
		}
	}

	_, identityCreated := authflow.FindMilestone[MilestoneDoCreateIdentity](flows.Nearest)
	_, standardAttributesPopulated := authflow.FindMilestone[MilestoneDoPopulateStandardAttributes](flows.Nearest)
	_, nestedStepHandled := authflow.FindMilestone[MilestoneNestedSteps](flows.Nearest)

	switch {
	case identityCreated && !standardAttributesPopulated && !nestedStepHandled:
		iden := i.identityInfo(flows.Nearest)
		return authflow.NewNodeSimple(&NodeDoPopulateStandardAttributes{
			Identity: iden,
		}), nil
	case identityCreated && standardAttributesPopulated && !nestedStepHandled:
		identification := i.identificationMethod(flows.Nearest)
		return authflow.NewSubFlow(&IntentPromoteFlowSteps{
			JSONPointer: i.jsonPointer(step, identification),
			UserID:      i.UserID,
		}), nil
	default:
		return nil, authflow.ErrIncompatibleInput
	}
}

func (i *IntentPromoteFlowStepIdentify) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	return intentPromoteFlowStepIdentifyData{
		Options: i.Options,
	}, nil
}

func (*IntentPromoteFlowStepIdentify) step(o config.AuthenticationFlowObject) *config.AuthenticationFlowSignupFlowStep {
	step, ok := o.(*config.AuthenticationFlowSignupFlowStep)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}

	return step
}

func (*IntentPromoteFlowStepIdentify) checkIdentificationMethod(deps *authflow.Dependencies, step *config.AuthenticationFlowSignupFlowStep, im config.AuthenticationFlowIdentification) (idx int, err error) {
	idx = -1

	for index, branch := range step.OneOf {
		branch := branch
		if im == branch.Identification {
			idx = index
		}
	}

	if idx >= 0 {
		return
	}

	err = authflow.ErrIncompatibleInput
	return
}

func (*IntentPromoteFlowStepIdentify) identificationMethod(w *authflow.Flow) config.AuthenticationFlowIdentification {
	m, ok := authflow.FindMilestone[MilestoneIdentificationMethod](w)
	if !ok {
		panic(fmt.Errorf("identification method not yet selected"))
	}

	im := m.MilestoneIdentificationMethod()

	return im
}

func (i *IntentPromoteFlowStepIdentify) jsonPointer(step *config.AuthenticationFlowSignupFlowStep, im config.AuthenticationFlowIdentification) jsonpointer.T {
	for idx, branch := range step.OneOf {
		branch := branch
		if branch.Identification == im {
			return authflow.JSONPointerForOneOf(i.JSONPointer, idx)
		}
	}

	panic(fmt.Errorf("selected identification method is not allowed"))
}

func (*IntentPromoteFlowStepIdentify) identityInfo(w *authflow.Flow) *identity.Info {
	m, ok := authflow.FindMilestone[MilestoneDoCreateIdentity](w)
	if !ok {
		panic(fmt.Errorf("MilestoneDoCreateIdentity cannot be found in IntentPromoteFlowStepIdentify"))
	}
	info := m.MilestoneDoCreateIdentity()
	return info
}
