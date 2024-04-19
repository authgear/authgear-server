package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

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

type IntentSignupFlowStepCreateAuthenticatorData struct {
	TypedData
	Options []CreateAuthenticatorOption `json:"options,omitempty"`
}

func NewIntentSignupFlowStepCreateAuthenticatorData(d IntentSignupFlowStepCreateAuthenticatorData) IntentSignupFlowStepCreateAuthenticatorData {
	d.Type = DataTypeCreateAuthenticatorData
	return d
}

var _ authflow.Data = &IntentSignupFlowStepCreateAuthenticatorData{}

func (m IntentSignupFlowStepCreateAuthenticatorData) Data() {}

type IntentSignupFlowStepCreateAuthenticator struct {
	FlowReference          authflow.FlowReference `json:"flow_reference,omitempty"`
	JSONPointer            jsonpointer.T          `json:"json_pointer,omitempty"`
	StepName               string                 `json:"step_name,omitempty"`
	UserID                 string                 `json:"user_id,omitempty"`
	IsUpdatingExistingUser bool                   `json:"is_updating_existing_user,omitempty"`
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
func (i *IntentSignupFlowStepCreateAuthenticator) MilestoneSwitchToExistingUser(deps *authflow.Dependencies, flow *authflow.Flow, newUserID string) error {
	i.UserID = newUserID
	i.IsUpdatingExistingUser = true

	milestone, ok := authflow.FindFirstMilestone[MilestoneDoCreateAuthenticator](flow)
	if ok {
		authn := milestone.MilestoneDoCreateAuthenticator()
		existing, err := i.findAuthenticatorOfSameType(deps, authn.Type)
		if err != nil {
			return err
		}
		if existing != nil {
			milestone.MilestoneDoCreateAuthenticatorSkipCreate()
		}
	}

	return nil
}

func (*IntentSignupFlowStepCreateAuthenticator) Kind() string {
	return "IntentSignupFlowStepCreateAuthenticator"
}

func (i *IntentSignupFlowStepCreateAuthenticator) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	internalOptions, err := i.getOptions(ctx, deps, flows)
	if err != nil {
		return nil, err
	}
	if len(internalOptions) == 0 {
		// Nothing can be selected, skip this step.
		return nil, authflow.ErrEOF
	}

	if i.IsUpdatingExistingUser {
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
	// TODO(tung): Auto select the authentication method when possible if IsUpdatingExistingUser
	if len(flows.Nearest.Nodes) == 0 {
		flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
		if err != nil {
			return nil, err
		}
		current, err := authflow.FlowObject(flowRootObject, i.JSONPointer)
		if err != nil {
			return nil, err
		}
		step := i.step(current)
		return &InputSchemaSignupFlowStepCreateAuthenticator{
			FlowRootObject: flowRootObject,
			JSONPointer:    i.JSONPointer,
			OneOf:          step.OneOf,
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

func (i *IntentSignupFlowStepCreateAuthenticator) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {

	current, err := i.currentFlowObject(deps)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	if i.IsUpdatingExistingUser {
		option, idx, authn, err := i.findSkippableOption(ctx, deps, flows)
		if err != nil {
			return nil, err
		}
		if option != nil {
			return i.reactToExistingAuthenticator(ctx, deps, flows, *option, authn, idx)
		}
	}

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
				return authflow.NewSubFlow(&IntentCreateAuthenticatorOOBOTP{
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
	internalOptions, err := i.getOptions(ctx, deps, flows)
	if err != nil {
		return nil, err
	}

	return NewIntentSignupFlowStepCreateAuthenticatorData(IntentSignupFlowStepCreateAuthenticatorData{
		Options: slice.Map(internalOptions, func(o CreateAuthenticatorOptionInternal) CreateAuthenticatorOption {
			return o.CreateAuthenticatorOption
		}),
	}), nil
}

func (*IntentSignupFlowStepCreateAuthenticator) step(o config.AuthenticationFlowObject) *config.AuthenticationFlowSignupFlowStep {
	step, ok := o.(*config.AuthenticationFlowSignupFlowStep)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}

	return step
}

func (i *IntentSignupFlowStepCreateAuthenticator) checkAuthenticationMethod(deps *authflow.Dependencies, step *config.AuthenticationFlowSignupFlowStep, am config.AuthenticationFlowAuthentication) (idx int, err error) {
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

func (*IntentSignupFlowStepCreateAuthenticator) authenticationMethod(flows authflow.Flows) config.AuthenticationFlowAuthentication {
	m, ok := authflow.FindMilestone[MilestoneAuthenticationMethod](flows.Nearest)
	if !ok {
		panic(fmt.Errorf("authentication method not yet selected"))
	}

	am := m.MilestoneAuthenticationMethod()

	return am
}

func (i *IntentSignupFlowStepCreateAuthenticator) jsonPointer(step *config.AuthenticationFlowSignupFlowStep, am config.AuthenticationFlowAuthentication) jsonpointer.T {
	for idx, branch := range step.OneOf {
		branch := branch
		if branch.Authentication == am {
			return authflow.JSONPointerForOneOf(i.JSONPointer, idx)
		}
	}

	panic(fmt.Errorf("selected identification method is not allowed"))
}

func (i *IntentSignupFlowStepCreateAuthenticator) currentFlowObject(deps *authflow.Dependencies) (config.AuthenticationFlowObject, error) {
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

func (i *IntentSignupFlowStepCreateAuthenticator) findAuthenticatorOfSameType(deps *authflow.Dependencies, typ model.AuthenticatorType) (*authenticator.Info, error) {

	userAuthns, err := deps.Authenticators.List(i.UserID)
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

func (i *IntentSignupFlowStepCreateAuthenticator) getOptions(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) ([]CreateAuthenticatorOptionInternal, error) {
	current, err := i.currentFlowObject(deps)
	if err != nil {
		return nil, err
	}
	step := i.step(current)
	options, err := NewCreateAuthenticationOptions(ctx, deps, flows, step, i.UserID)
	if err != nil {
		return nil, err
	}
	return options, nil
}

func (i *IntentSignupFlowStepCreateAuthenticator) reactToExistingAuthenticator(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, option CreateAuthenticatorOptionInternal, authn *authenticator.Info, idx int) (*authflow.Node, error) {
	if len(flows.Nearest.Nodes) == 0 {
		return authflow.NewNodeSimple(&NodeSkipCreationByExistingAuthenticator{
			Authenticator:  authn,
			JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
			Authentication: option.Authentication,
		}), nil
	}

	_, authenticatorCreated := authflow.FindMilestone[MilestoneDoCreateAuthenticator](flows.Nearest)
	_, nestedStepsHandled := authflow.FindMilestone[MilestoneNestedSteps](flows.Nearest)

	current, err := i.currentFlowObject(deps)
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
	userAuthns, err := deps.Authenticators.List(i.UserID)
	if err != nil {
		return nil, -1, nil, err
	}
	// For each option, see if any existing identities can be reused
	options, err := i.getOptions(ctx, deps, flows)
	if err != nil {
		return nil, -1, nil, err
	}
	for idx, option := range options {
		option := option
		existingAuthn := i.findAuthenticatorByOption(userAuthns, option)
		if existingAuthn != nil {
			return &option, idx, existingAuthn, nil
		}
	}
	return nil, -1, nil, nil
}

func (i *IntentSignupFlowStepCreateAuthenticator) findAuthenticatorByOption(in []*authenticator.Info, option CreateAuthenticatorOptionInternal) *authenticator.Info {
	findPassword := func(kind model.AuthenticatorKind) *authenticator.Info {
		for _, authn := range in {
			authn := authn
			if authn.Type != model.AuthenticatorTypePassword {
				continue
			}
			if authn.Kind == kind {
				return authn
			}
		}
		return nil
	}

	findPrimaryPasskey := func(kind model.AuthenticatorKind) *authenticator.Info {
		for _, authn := range in {
			authn := authn
			if authn.Type != model.AuthenticatorTypePasskey {
				continue
			}
			if authn.Kind == kind {
				return authn
			}
		}
		return nil
	}

	findEmailOOB := func(kind model.AuthenticatorKind, target string) *authenticator.Info {
		for _, authn := range in {
			authn := authn
			if authn.Type != model.AuthenticatorTypeOOBEmail {
				continue
			}
			if authn.Kind == kind && authn.OOBOTP.Email == target {
				return authn
			}
		}
		return nil
	}

	findSMSOOB := func(kind model.AuthenticatorKind, target string) *authenticator.Info {
		for _, authn := range in {
			authn := authn
			if authn.Type != model.AuthenticatorTypeOOBSMS {
				continue
			}
			if authn.Kind == kind && authn.OOBOTP.Phone == target {
				return authn
			}
		}
		return nil
	}

	findTOTP := func(kind model.AuthenticatorKind) *authenticator.Info {
		for _, authn := range in {
			authn := authn
			if authn.Type != model.AuthenticatorTypeTOTP {
				continue
			}
			if authn.Kind == kind {
				return authn
			}
		}
		return nil
	}

	switch option.Authentication {
	case config.AuthenticationFlowAuthenticationPrimaryPassword:
		return findPassword(authenticator.KindPrimary)
	case config.AuthenticationFlowAuthenticationSecondaryPassword:
		return findPassword(authenticator.KindSecondary)
	case config.AuthenticationFlowAuthenticationPrimaryPasskey:
		return findPrimaryPasskey(authenticator.KindPrimary)
	case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
		return findEmailOOB(authenticator.KindPrimary, option.UnmaskedTarget)
	case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
		return findEmailOOB(authenticator.KindSecondary, option.UnmaskedTarget)
	case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
		return findSMSOOB(authenticator.KindPrimary, option.UnmaskedTarget)
	case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
		return findSMSOOB(authenticator.KindSecondary, option.UnmaskedTarget)
	case config.AuthenticationFlowAuthenticationSecondaryTOTP:
		return findTOTP(authenticator.KindSecondary)
	}
	return nil
}
