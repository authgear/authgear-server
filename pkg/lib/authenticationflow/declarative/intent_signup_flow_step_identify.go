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
	authflow.RegisterIntent(&IntentSignupFlowStepIdentify{})
}

type IntentSignupFlowStepIdentify struct {
	FlowReference          authflow.FlowReference `json:"flow_reference,omitempty"`
	JSONPointer            jsonpointer.T          `json:"json_pointer,omitempty"`
	StepName               string                 `json:"step_name,omitempty"`
	UserID                 string                 `json:"user_id,omitempty"`
	Options                []IdentificationOption `json:"options,omitempty"`
	IsUpdatingExistingUser bool                   `json:"is_updating_existing_user,omitempty"`
	IsCreateSkipped        bool                   `json:"is_create_skipped,omitempty"`
}

var _ authflow.TargetStep = &IntentSignupFlowStepIdentify{}
var _ authflow.Milestone = &IntentSignupFlowStepIdentify{}
var _ MilestoneSwitchToExistingUser = &IntentSignupFlowStepIdentify{}

func (*IntentSignupFlowStepIdentify) Milestone() {}
func (i *IntentSignupFlowStepIdentify) MilestoneSwitchToExistingUser(deps *authflow.Dependencies, flow *authflow.Flow, newUserID string) error {
	i.IsUpdatingExistingUser = true
	i.UserID = newUserID

	milestoneDoCreateIdentity, ok := authflow.FindFirstMilestone[MilestoneDoCreateIdentity](flow)
	if ok {
		iden := milestoneDoCreateIdentity.MilestoneDoCreateIdentity()
		idenSpec := iden.ToSpec()
		idenWithSameType, err := i.findIdentityOfSameType(deps, &idenSpec)
		if err != nil {
			return err
		}
		if idenWithSameType != nil {
			milestoneDoCreateIdentity.MilestoneDoCreateIdentitySkipCreate()
			i.IsCreateSkipped = true
		} else {
			milestoneDoCreateIdentity.MilestoneDoCreateIdentityUpdate(iden.UpdateUserID(newUserID))
		}
	}
	milestoneDoPopulateStandardAttributes, ok := authflow.FindFirstMilestone[MilestoneDoPopulateStandardAttributes](flow)
	if ok {
		// Always skip population
		milestoneDoPopulateStandardAttributes.MilestoneDoPopulateStandardAttributesSkip()
	}
	return nil
}

func (i *IntentSignupFlowStepIdentify) GetName() string {
	return i.StepName
}

func (i *IntentSignupFlowStepIdentify) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

var _ IntentSignupFlowStepVerifyTarget = &IntentSignupFlowStepIdentify{}

func (i *IntentSignupFlowStepIdentify) GetVerifiableClaims(_ context.Context, _ *authflow.Dependencies, flows authflow.Flows) (map[model.ClaimName]string, error) {
	if i.IsCreateSkipped {
		return nil, nil
	}

	m, ok := authflow.FindMilestoneInCurrentFlow[MilestoneDoCreateIdentity](flows.Nearest)
	if !ok {
		return nil, fmt.Errorf("MilestoneDoCreateIdentity cannot be found in IntentSignupFlowStepIdentify")
	}
	info := m.MilestoneDoCreateIdentity()

	return info.IdentityAwareStandardClaims(), nil
}

func (*IntentSignupFlowStepIdentify) GetPurpose(_ context.Context, _ *authflow.Dependencies, _ authflow.Flows) otp.Purpose {
	return otp.PurposeVerification
}

func (*IntentSignupFlowStepIdentify) GetMessageType(_ context.Context, _ *authflow.Dependencies, _ authflow.Flows) otp.MessageType {
	return otp.MessageTypeVerification
}

var _ IntentSignupFlowStepCreateAuthenticatorTarget = &IntentSignupFlowStepIdentify{}

func (n *IntentSignupFlowStepIdentify) GetOOBOTPClaims(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (map[model.ClaimName]string, error) {
	return n.GetVerifiableClaims(ctx, deps, flows)
}

func (n *IntentSignupFlowStepIdentify) IsSkipped() bool {
	return n.IsCreateSkipped
}

var _ authflow.Intent = &IntentSignupFlowStepIdentify{}
var _ authflow.DataOutputer = &IntentSignupFlowStepIdentify{}

func NewIntentSignupFlowStepIdentify(ctx context.Context, deps *authflow.Dependencies, i *IntentSignupFlowStepIdentify) (*IntentSignupFlowStepIdentify, error) {
	current, err := i.currentFlowObject(deps)
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

func (*IntentSignupFlowStepIdentify) Kind() string {
	return "IntentSignupFlowStepIdentify"
}

func (i *IntentSignupFlowStepIdentify) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	if len(flows.Nearest.Nodes) == 0 && i.IsUpdatingExistingUser {
		option, _, _, err := i.findSkippableOption(ctx, deps, flows)
		if err != nil {
			return nil, err
		}
		if option != nil {
			// Proceed without user input to use the existing identity automatically
			return nil, nil
		}
	}

	// Let the input to select which identification method to use.
	if len(flows.Nearest.Nodes) == 0 {
		flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
		if err != nil {
			return nil, err
		}
		return &InputSchemaStepIdentify{
			FlowRootObject: flowRootObject,
			JSONPointer:    i.JSONPointer,
			Options:        i.Options,
		}, nil
	}

	_, identityCreated := authflow.FindMilestone[MilestoneDoCreateIdentity](flows.Nearest)
	_, standardAttributesPopulated := authflow.FindMilestone[MilestoneDoPopulateStandardAttributes](flows.Nearest)
	_, nestedStepHandled := authflow.FindMilestoneInCurrentFlow[MilestoneNestedSteps](flows.Nearest)

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

func (i *IntentSignupFlowStepIdentify) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	if len(flows.Nearest.Nodes) == 0 && i.IsUpdatingExistingUser {
		option, idx, identity, err := i.findSkippableOption(ctx, deps, flows)
		if err != nil {
			return nil, err
		}
		if option != nil {
			return i.reactToExistingIdentity(ctx, deps, flows, *option, identity, idx)
		}
	}

	current, err := i.currentFlowObject(deps)
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

			switch identification {
			case config.AuthenticationFlowIdentificationEmail:
				fallthrough
			case config.AuthenticationFlowIdentificationPhone:
				fallthrough
			case config.AuthenticationFlowIdentificationUsername:
				return authflow.NewNodeSimple(&NodeCreateIdentityLoginID{
					JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
					UserID:         i.UserID,
					Identification: identification,
				}), nil
			case config.AuthenticationFlowIdentificationOAuth:
				return authflow.NewSubFlow(&IntentOAuth{
					JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
					NewUserID:      i.UserID,
					Identification: identification,
				}), nil
			case config.AuthenticationFlowIdentificationPasskey:
				// Cannot create passkey in this step.
				return nil, authflow.ErrIncompatibleInput
			}
		}
		return nil, authflow.ErrIncompatibleInput
	}

	_, identityCreated := authflow.FindMilestone[MilestoneDoCreateIdentity](flows.Nearest)
	_, standardAttributesPopulated := authflow.FindMilestone[MilestoneDoPopulateStandardAttributes](flows.Nearest)
	_, nestedStepHandled := authflow.FindMilestoneInCurrentFlow[MilestoneNestedSteps](flows.Nearest)

	switch {
	case identityCreated && !standardAttributesPopulated && !nestedStepHandled:
		iden := i.identityInfo(flows.Nearest)
		return authflow.NewNodeSimple(&NodeDoPopulateStandardAttributesInSignup{
			Identity:   iden,
			SkipUpdate: i.IsUpdatingExistingUser,
		}), nil
	case identityCreated && standardAttributesPopulated && !nestedStepHandled:
		identification := i.identificationMethod(flows.Nearest)
		return authflow.NewSubFlow(&IntentSignupFlowSteps{
			FlowReference:          i.FlowReference,
			JSONPointer:            i.jsonPointer(step, identification),
			UserID:                 i.UserID,
			IsUpdatingExistingUser: i.IsUpdatingExistingUser,
		}), nil
	default:
		return nil, authflow.ErrIncompatibleInput
	}
}

func (i *IntentSignupFlowStepIdentify) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	return NewIdentificationData(IdentificationData{
		Options: i.Options,
	}), nil
}

func (*IntentSignupFlowStepIdentify) step(o config.AuthenticationFlowObject) *config.AuthenticationFlowSignupFlowStep {
	step, ok := o.(*config.AuthenticationFlowSignupFlowStep)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}

	return step
}

func (*IntentSignupFlowStepIdentify) checkIdentificationMethod(deps *authflow.Dependencies, step *config.AuthenticationFlowSignupFlowStep, im config.AuthenticationFlowIdentification) (idx int, err error) {
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

func (*IntentSignupFlowStepIdentify) identificationMethod(w *authflow.Flow) config.AuthenticationFlowIdentification {
	m, ok := authflow.FindMilestone[MilestoneIdentificationMethod](w)
	if !ok {
		panic(fmt.Errorf("identification method not yet selected"))
	}

	im := m.MilestoneIdentificationMethod()

	return im
}

func (i *IntentSignupFlowStepIdentify) jsonPointer(step *config.AuthenticationFlowSignupFlowStep, im config.AuthenticationFlowIdentification) jsonpointer.T {
	for idx, branch := range step.OneOf {
		branch := branch
		if branch.Identification == im {
			return authflow.JSONPointerForOneOf(i.JSONPointer, idx)
		}
	}

	panic(fmt.Errorf("selected identification method is not allowed"))
}

func (*IntentSignupFlowStepIdentify) identityInfo(w *authflow.Flow) *identity.Info {
	m, ok := authflow.FindMilestone[MilestoneDoCreateIdentity](w)
	if !ok {
		panic(fmt.Errorf("MilestoneDoCreateIdentity cannot be found in IntentSignupFlowStepIdentify"))
	}
	info := m.MilestoneDoCreateIdentity()
	return info
}

func (i *IntentSignupFlowStepIdentify) currentFlowObject(deps *authflow.Dependencies) (config.AuthenticationFlowObject, error) {
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

func (i *IntentSignupFlowStepIdentify) findIdentityOfSameType(deps *authflow.Dependencies, spec *identity.Spec) (*identity.Info, error) {

	userIdens, err := deps.Identities.ListByUser(i.UserID)
	if err != nil {
		return nil, err
	}

	var idenWithSameType *identity.Info

	for _, uiden := range userIdens {
		uiden := uiden
		if uiden.Type == spec.Type {
			if spec.Type == model.IdentityTypeLoginID {
				// Only login id needs to check the key
				if spec.LoginID.Key == uiden.LoginID.LoginIDKey {
					idenWithSameType = uiden
				}
			} else {
				// For others, just check they are same type
				idenWithSameType = uiden
			}
		}
	}

	return idenWithSameType, nil
}

func (i *IntentSignupFlowStepIdentify) reactToExistingIdentity(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, option IdentificationOption, identity *identity.Info, idx int) (*authflow.Node, error) {
	if len(flows.Nearest.Nodes) == 0 {
		return authflow.NewNodeSimple(&NodeSkipCreationByExistingIdentity{
			Identity:       identity,
			JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
			Identification: option.Identification,
		}), nil
	}

	_, identityCreated := authflow.FindMilestone[MilestoneDoCreateIdentity](flows.Nearest)
	_, nestedStepHandled := authflow.FindMilestone[MilestoneNestedSteps](flows.Nearest)

	current, err := i.currentFlowObject(deps)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	switch {
	case identityCreated && !nestedStepHandled:
		identification := i.identificationMethod(flows.Nearest)
		return authflow.NewSubFlow(&IntentSignupFlowSteps{
			FlowReference:          i.FlowReference,
			JSONPointer:            i.jsonPointer(step, identification),
			UserID:                 i.UserID,
			IsUpdatingExistingUser: i.IsUpdatingExistingUser,
		}), nil
	default:
		return nil, authflow.ErrIncompatibleInput
	}
}

func (i *IntentSignupFlowStepIdentify) findSkippableOption(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (option *IdentificationOption, idx int, info *identity.Info, err error) {
	userIdens, err := deps.Identities.ListByUser(i.UserID)
	if err != nil {
		return nil, -1, nil, err
	}
	// For each option, see if any existing identities can be reused
	for idx, option := range i.Options {
		option := option
		existingIden := i.findIdentityByOption(userIdens, option)
		if existingIden != nil {
			return &option, idx, existingIden, nil
		}
	}
	return nil, -1, nil, nil
}

func (i *IntentSignupFlowStepIdentify) findIdentityByOption(in []*identity.Info, option IdentificationOption) *identity.Info {
	findLoginID := func(typ model.LoginIDKeyType) *identity.Info {
		for _, iden := range in {
			iden := iden
			if iden.Type != model.IdentityTypeLoginID {
				continue
			}
			if iden.LoginID.LoginIDType == typ {
				return iden
			}
		}
		return nil
	}

	findOAuth := func(alias string) *identity.Info {
		for _, iden := range in {
			iden := iden
			if iden.Type != model.IdentityTypeOAuth {
				continue
			}
			if iden.OAuth.ProviderAlias == alias {
				return iden
			}
		}
		return nil
	}

	switch option.Identification {
	case config.AuthenticationFlowIdentificationEmail:
		return findLoginID(model.LoginIDKeyTypeEmail)
	case config.AuthenticationFlowIdentificationPhone:
		return findLoginID(model.LoginIDKeyTypePhone)
	case config.AuthenticationFlowIdentificationUsername:
		return findLoginID(model.LoginIDKeyTypeUsername)
	case config.AuthenticationFlowIdentificationOAuth:
		return findOAuth(option.Alias)
	case config.AuthenticationFlowIdentificationPasskey:
		// Cannot create passkey in this step.
		return nil
	}
	return nil
}
