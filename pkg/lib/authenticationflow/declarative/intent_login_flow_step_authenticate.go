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
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
)

type IntentLoginFlowStepAuthenticateTarget interface {
	IntentLoginFlowStepAuthenticateTarget(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (*identity.Info, bool)
}

func init() {
	authflow.RegisterIntent(&IntentLoginFlowStepAuthenticate{})
}

// IntentLoginFlowStepAuthenticate
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
//
//   IntentUseRecoveryCode (MilestoneFlowAuthenticate)
//     NodeDoConsumeRecoveryCode (MilestoneDidAuthenticate)

type IntentLoginFlowStepAuthenticate struct {
	FlowReference      authflow.FlowReference `json:"flow_reference,omitempty"`
	JSONPointer        jsonpointer.T          `json:"json_pointer,omitempty"`
	StepName           string                 `json:"step_name,omitempty"`
	UserID             string                 `json:"user_id,omitempty"`
	Options            []AuthenticateOption   `json:"options"`
	DeviceTokenEnabled bool                   `json:"device_token_enabled"`

	ShowUntilAMRConstraintsFulfilled bool `json:"show_until_amr_constraints_fulfilled,omitempty"`
}

var _ authflow.TargetStep = &IntentLoginFlowStepAuthenticate{}

func (i *IntentLoginFlowStepAuthenticate) GetName() string {
	return i.StepName
}

func (i *IntentLoginFlowStepAuthenticate) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

var _ IntentLoginFlowStepChangePasswordTarget = &IntentLoginFlowStepAuthenticate{}

func (i *IntentLoginFlowStepAuthenticate) GetChangeRequiredPasswordAuthenticator(_ context.Context, _ *authflow.Dependencies, flows authflow.Flows) (info *authenticator.Info, changeRequiredReason PasswordChangeReason) {
	milestones := authflow.FindAllMilestones[MilestoneDoUseAuthenticatorPassword](flows.Nearest)
	var targetMilestone MilestoneDoUseAuthenticatorPassword
	for _, m := range milestones {
		p := authflow.JSONPointerToParent(m.GetJSONPointer())
		if p.String() == i.GetJSONPointer().String() {
			targetMilestone = m
			break
		}
	}

	if targetMilestone != nil {
		n := targetMilestone.MilestoneDoUseAuthenticatorPassword()
		if n.PasswordChangeRequired {
			info = n.Authenticator
			changeRequiredReason = n.PasswordChangeReason
		}
	}

	return
}

var _ authflow.Intent = &IntentLoginFlowStepAuthenticate{}
var _ authflow.DataOutputer = &IntentLoginFlowStepAuthenticate{}

func NewIntentLoginFlowStepAuthenticate(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, i *IntentLoginFlowStepAuthenticate) (*IntentLoginFlowStepAuthenticate, error) {
	current, err := i.currentFlowObject(deps)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	options, deviceTokenEnabled, err := getAuthenticationOptionsForLogin(ctx, deps, flows, i.UserID, step)
	if err != nil {
		return nil, err
	}

	i.Options = options
	i.DeviceTokenEnabled = deviceTokenEnabled

	if step.IsShowUntilAMRConstraintsFulfilled() {
		i.ShowUntilAMRConstraintsFulfilled = true
	}
	return i, nil
}

func (*IntentLoginFlowStepAuthenticate) Kind() string {
	return "IntentLoginFlowStepAuthenticate"
}

func (i *IntentLoginFlowStepAuthenticate) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
	if err != nil {
		return nil, err
	}
	current, err := i.currentFlowObject(deps)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	if i.ShowUntilAMRConstraintsFulfilled {
		remainingAMRs, err := remainingAMRConstraintsInFlow(ctx, deps, flows)
		if err != nil {
			return nil, err
		}
		// No remaining AMRs, end
		if len(remainingAMRs) == 0 {
			return nil, authflow.ErrEOF
		}
		// Let ReactTo create sub-authenticate steps
		return nil, nil
	}

	_, _, deviceTokenInspected := authflow.FindMilestoneInCurrentFlow[MilestoneDeviceTokenInspected](flows)

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

	_, _, deviceTokenCreatedIfRequested := authflow.FindMilestoneInCurrentFlow[MilestoneDoCreateDeviceTokenIfRequested](flows)

	_, _, nestedStepsHandled := authflow.FindMilestoneInCurrentFlow[MilestoneNestedSteps](flows)

	createAuthenticatorMilestones := authflow.FindAllMilestones[MilestoneFlowCreateAuthenticator](flows.Root)

	switch {
	case i.DeviceTokenEnabled && !deviceTokenInspected:
		// Inspect the device token
		return nil, nil
	case !authenticationMethodSelected:
		if len(i.Options) == 0 {
			if step.IsOptional() {
				// Skip this step and any nested step.
				return nil, authflow.ErrEOF
			}

			if len(createAuthenticatorMilestones) > 0 {
				// Skip this step if the user has just created authenticator.
				return nil, authflow.ErrEOF
			}
		}

		shouldBypassBotProtection := ShouldExistingResultBypassBotProtectionRequirement(ctx)
		// Let the input to select which authentication method to use.
		return &InputSchemaLoginFlowStepAuthenticate{
			FlowRootObject:            flowRootObject,
			JSONPointer:               i.JSONPointer,
			Options:                   i.Options,
			DeviceTokenEnabled:        i.DeviceTokenEnabled,
			ShouldBypassBotProtection: shouldBypassBotProtection,
			BotProtectionCfg:          deps.Config.BotProtection,
		}, nil
	case !authenticated:
		// This branch is only reached when there is a programming error.
		// We expect the selected authentication method to be authenticated before this intent becomes input reactor again.
		panic(fmt.Errorf("unauthenticated"))

	case i.DeviceTokenEnabled && !deviceTokenCreatedIfRequested:
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

func (i *IntentLoginFlowStepAuthenticate) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	current, err := i.currentFlowObject(deps)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	if i.ShowUntilAMRConstraintsFulfilled {
		return i.newIntentLoginFlowStepAuthenticateForAMRConstraint(ctx, deps, flows)
	}

	_, _, deviceTokenInspected := authflow.FindMilestoneInCurrentFlow[MilestoneDeviceTokenInspected](flows)

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

	_, _, deviceTokenCreatedIfRequested := authflow.FindMilestoneInCurrentFlow[MilestoneDoCreateDeviceTokenIfRequested](flows)

	_, _, nestedStepsHandled := authflow.FindMilestoneInCurrentFlow[MilestoneNestedSteps](flows)

	switch {
	case i.DeviceTokenEnabled && !deviceTokenInspected:
		return authflow.NewSubFlow(&IntentInspectDeviceToken{
			UserID: i.UserID,
		}), nil

	case !authenticationMethodSelected:
		var inputTakeAuthenticationMethod inputTakeAuthenticationMethod
		if authflow.AsInput(input, &inputTakeAuthenticationMethod) {
			authentication := inputTakeAuthenticationMethod.GetAuthenticationMethod()

			if len(i.Options) == 0 {
				shouldCreateAuthenticator, err := i.canCreateAuthenticator(ctx, step, deps)
				if err != nil {
					return nil, err
				}

				if shouldCreateAuthenticator {
					nextStep := &IntentLoginFlowStepCreateAuthenticator{
						FlowReference:          i.FlowReference,
						StepName:               step.Name,
						JSONPointer:            i.JSONPointer,
						UserID:                 i.UserID,
						IsUpdatingExistingUser: true,
					}
					return authflow.NewSubFlow(nextStep), nil
				} else {
					// Otherwise this step is NON-optional but have no options
					return nil, api.ErrNoAuthenticator
				}
			}

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
			case config.AuthenticationFlowAuthenticationRecoveryCode:
				return authflow.NewSubFlow(&IntentUseRecoveryCode{
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
	case i.DeviceTokenEnabled && !deviceTokenCreatedIfRequested:
		return authflow.NewSubFlow(&IntentCreateDeviceTokenIfRequested{
			JSONPointer: authflow.JSONPointerForOneOf(i.JSONPointer, i.deviceTokenIndex(step)),
			UserID:      i.UserID,
		}), nil
	case !nestedStepsHandled:
		authentication := i.authenticationMethod(flows)
		return authflow.NewSubFlow(&IntentLoginFlowSteps{
			FlowReference: i.FlowReference,
			JSONPointer:   i.jsonPointer(step, authentication),
		}), nil
	default:
		return nil, authflow.ErrIncompatibleInput
	}
}

func (i *IntentLoginFlowStepAuthenticate) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {

	options := []AuthenticateOptionForOutput{}
	for _, o := range i.Options {
		options = append(options, o.ToOutput(ctx))
	}

	return NewStepAuthenticateData(StepAuthenticateData{
		Options:            options,
		DeviceTokenEnabled: i.DeviceTokenEnabled,
	}), nil
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

func (i *IntentLoginFlowStepAuthenticate) canCreateAuthenticator(ctx context.Context, step *config.AuthenticationFlowLoginFlowStep, deps *authflow.Dependencies) (bool, error) {
	authenticationConfig := deps.Config.Authentication
	if authenticationConfig.SecondaryAuthenticationGracePeriod != nil &&
		authenticationConfig.SecondaryAuthenticationGracePeriod.Enabled &&
		(authenticationConfig.SecondaryAuthenticationGracePeriod.EndAt == nil || authenticationConfig.SecondaryAuthenticationGracePeriod.EndAt.After(deps.Clock.NowUTC())) {
		return true, nil
	}

	user, err := deps.Users.Get(ctx, i.UserID, accesscontrol.RoleGreatest)
	if err != nil {
		return false, err
	}
	if user.MFAGracePeriodtEndAt != nil && deps.Clock.NowUTC().Before(*user.MFAGracePeriodtEndAt) {
		return true, nil
	}

	return false, nil
}

func (i *IntentLoginFlowStepAuthenticate) currentFlowObject(deps *authflow.Dependencies) (config.AuthenticationFlowObject, error) {
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

func (i *IntentLoginFlowStepAuthenticate) step(o config.AuthenticationFlowObject) *config.AuthenticationFlowLoginFlowStep {
	step, ok := o.(*config.AuthenticationFlowLoginFlowStep)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}

	return step
}

func (*IntentLoginFlowStepAuthenticate) authenticationMethod(flows authflow.Flows) config.AuthenticationFlowAuthentication {
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

func (i *IntentLoginFlowStepAuthenticate) jsonPointer(step *config.AuthenticationFlowLoginFlowStep, am config.AuthenticationFlowAuthentication) jsonpointer.T {
	for idx, branch := range step.OneOf {
		branch := branch
		if branch.Authentication == am {
			return authflow.JSONPointerForOneOf(i.JSONPointer, idx)
		}
	}

	panic(fmt.Errorf("selected authentication method is not allowed"))
}

func (i *IntentLoginFlowStepAuthenticate) newIntentLoginFlowStepAuthenticateForAMRConstraint(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.ReactToResult, error) {
	current, err := i.currentFlowObject(deps)
	if err != nil {
		return nil, err
	}
	step := i.step(current)
	subintent := i.clone()
	remainingAMRs, err := remainingAMRConstraintsInFlow(ctx, deps, flows)
	if err != nil {
		return nil, err
	}
	// The subflow should not check constraints again
	subintent.ShowUntilAMRConstraintsFulfilled = false
	// To fulfil AMR contraints, device tokens cannot be used
	subintent.DeviceTokenEnabled = false

	options, _, err := getAuthenticationOptionsForLogin(ctx, deps, flows, i.UserID, step)
	if err != nil {
		return nil, err
	}
	// The subflow should only contain options that can fulfill remaining amr
	newOptions := filterAMROptionsByAMRConstraint(options, remainingAMRs)
	subintent.Options = newOptions
	return authflow.NewSubFlow(subintent), nil
}

func (i *IntentLoginFlowStepAuthenticate) clone() *IntentLoginFlowStepAuthenticate {
	s := struct {
		FlowReference                    authflow.FlowReference
		JSONPointer                      jsonpointer.T
		StepName                         string
		UserID                           string
		Options                          []AuthenticateOption
		DeviceTokenEnabled               bool
		ShowUntilAMRConstraintsFulfilled bool
	}{
		FlowReference:                    i.FlowReference,
		JSONPointer:                      i.JSONPointer,
		StepName:                         i.StepName,
		UserID:                           i.UserID,
		Options:                          i.Options,
		DeviceTokenEnabled:               i.DeviceTokenEnabled,
		ShowUntilAMRConstraintsFulfilled: i.ShowUntilAMRConstraintsFulfilled,
	}
	cloned := IntentLoginFlowStepAuthenticate(s)
	return &cloned
}
