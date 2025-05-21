package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentLoginFlowStepIdentify{})
}

// IntentLoginFlowStepIdentify
//   IntentUseIdentityLoginID (MilestoneIdentificationMethod, MilestoneFlowUseIdentity)
//     NodeDoUseIdentity (MilestoneDoUseIdentity)
//
//   IntentOAuth (MilestoneIdentificationMethod, MilestoneFlowUseIdentity)
//     NodeOAuth
//     NodeDoUseIdentity (MilestoneDoUseIdentity)
//
//   IntentUseIdentityPasskey (MilestoneIdentificationMethod, MilestoneFlowUseIdentity)
//     NodeDoUseIdentityPasskey (MilestoneDoUseIdentity)
//
//   IntentIdentifyWithIDToken (MilestoneIdentificationMethod, MilestoneFlowUseIdentity)
//     NodeDoUseIDToken (MilestoneDoUseUser, not MilestoneDoUseIdentity!)

type IntentLoginFlowStepIdentify struct {
	FlowReference authflow.FlowReference `json:"flow_reference,omitempty"`
	JSONPointer   jsonpointer.T          `json:"json_pointer,omitempty"`
	StepName      string                 `json:"step_name,omitempty"`
	Options       []IdentificationOption `json:"options"`
}

var _ authflow.TargetStep = &IntentLoginFlowStepIdentify{}

func (i *IntentLoginFlowStepIdentify) GetName() string {
	return i.StepName
}

func (i *IntentLoginFlowStepIdentify) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

var _ IntentLoginFlowStepAuthenticateTarget = &IntentLoginFlowStepIdentify{}

func (*IntentLoginFlowStepIdentify) IntentLoginFlowStepAuthenticateTarget(_ context.Context, _ *authflow.Dependencies, flows authflow.Flows) (*identity.Info, bool) {
	m1, m1Flows, ok := authflow.FindMilestoneInCurrentFlow[MilestoneFlowUseIdentity](flows)
	if !ok {
		panic(fmt.Errorf("MilestoneFlowUseIdentity is absent in IntentLoginFlowStepIdentify"))
	}

	m2, _, ok := m1.MilestoneFlowUseIdentity(m1Flows)
	if !ok {
		// When ID token is used, there is no MilestoneDoUseIdentity
		return nil, false
	}

	info := m2.MilestoneDoUseIdentity()
	return info, true
}

var _ authflow.Intent = &IntentLoginFlowStepIdentify{}
var _ authflow.DataOutputer = &IntentLoginFlowStepIdentify{}

func NewIntentLoginFlowStepIdentify(ctx context.Context, deps *authflow.Dependencies, i *IntentLoginFlowStepIdentify) (*IntentLoginFlowStepIdentify, error) {
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
			c := NewIdentificationOptionLoginID(b.Identification, b.BotProtection, deps.Config.BotProtection)
			options = append(options, c)
		case config.AuthenticationFlowIdentificationOAuth:
			oauthOptions := NewIdentificationOptionsOAuth(
				deps.Config.Identity.OAuth,
				deps.FeatureConfig.Identity.OAuth.Providers, b.BotProtection, deps.Config.BotProtection,
				deps.SSOOAuthDemoCredentials,
			)
			options = append(options, oauthOptions...)
		case config.AuthenticationFlowIdentificationPasskey:
			requestOptions, err := deps.PasskeyRequestOptionsService.MakeModalRequestOptions(ctx)
			if err != nil {
				return nil, err
			}
			c := NewIdentificationOptionPasskey(requestOptions, b.BotProtection, deps.Config.BotProtection)
			options = append(options, c)
		case config.AuthenticationFlowIdentificationLDAP:
			ldapOptions := NewIdentificationOptionLDAP(deps.Config.Identity.LDAP, b.BotProtection, deps.Config.BotProtection)
			options = append(options, ldapOptions...)
			break
		case config.AuthenticationFlowIdentificationIDToken:
			// ID token is an advanced usage, and it inheritly does not support user interaction.
			// Thus bot protection is not supported.
			var botProtection *config.AuthenticationFlowBotProtection = nil
			c := NewIdentificationOptionIDToken(b.Identification, botProtection, deps.Config.BotProtection)
			options = append(options, c)
		}
	}

	i.Options = options
	return i, nil
}

func (*IntentLoginFlowStepIdentify) Kind() string {
	return "IntentLoginFlowStepIdentify"
}

func (i *IntentLoginFlowStepIdentify) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	_, _, identityUsed := authflow.FindMilestoneInCurrentFlow[MilestoneFlowUseIdentity](flows)
	_, _, loginHintChecked := authflow.FindMilestoneInCurrentFlow[MilestoneCheckLoginHint](flows)
	_, _, nestedStepsHandled := authflow.FindMilestoneInCurrentFlow[MilestoneNestedSteps](flows)

	switch {
	case len(flows.Nearest.Nodes) == 0:
		// Let the input to select which identification method to use.
		flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
		if err != nil {
			return nil, err
		}
		shouldBypassBotProtection := ShouldExistingResultBypassBotProtectionRequirement(ctx)
		return &InputSchemaStepIdentify{
			FlowRootObject:            flowRootObject,
			JSONPointer:               i.JSONPointer,
			Options:                   i.Options,
			ShouldBypassBotProtection: shouldBypassBotProtection,
			BotProtectionCfg:          deps.Config.BotProtection,
		}, nil
	case identityUsed && !loginHintChecked:
		// Check login_hint
		return nil, nil
	case identityUsed && !nestedStepsHandled:
		// Handle nested steps.
		return nil, nil
	default:
		return nil, authflow.ErrEOF
	}
}

func (i *IntentLoginFlowStepIdentify) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	current, err := i.currentFlowObject(deps)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	_, _, identityUsed := authflow.FindMilestoneInCurrentFlow[MilestoneFlowUseIdentity](flows)
	_, _, loginHintChecked := authflow.FindMilestoneInCurrentFlow[MilestoneCheckLoginHint](flows)
	_, _, nestedStepsHandled := authflow.FindMilestoneInCurrentFlow[MilestoneNestedSteps](flows)

	switch {
	case len(flows.Nearest.Nodes) == 0:
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
				return authflow.NewSubFlow(&IntentUseIdentityLoginID{
					JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
					Identification: identification,
				}), nil
			case config.AuthenticationFlowIdentificationOAuth:
				return authflow.NewSubFlow(&IntentOAuth{
					JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
					Identification: identification,
				}), nil
			case config.AuthenticationFlowIdentificationPasskey:
				return authflow.NewSubFlow(&IntentUseIdentityPasskey{
					JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
					Identification: identification,
				}), nil
			case config.AuthenticationFlowIdentificationLDAP:
				return authflow.NewSubFlow(&IntentLDAP{
					JSONPointer: authflow.JSONPointerForOneOf(i.JSONPointer, idx),
				}), nil
			case config.AuthenticationFlowIdentificationIDToken:
				return authflow.NewSubFlow(&IntentIdentifyWithIDToken{
					JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
					Identification: identification,
				}), nil
			}
		}
		return nil, authflow.ErrIncompatibleInput
	case identityUsed && !loginHintChecked:
		userID, err := getUserID(flows)
		if err != nil {
			panic("unexpected: identityUsed is true but no userID")
		}
		n, err := NewNodeCheckLoginHint(ctx, deps, userID)
		if err != nil {
			return nil, err
		}
		return authflow.NewNodeSimple(n), nil
	case identityUsed && !nestedStepsHandled:
		identification := i.identificationMethod(flows)
		return authflow.NewSubFlow(&IntentLoginFlowSteps{
			FlowReference: i.FlowReference,
			JSONPointer:   i.jsonPointer(step, identification),
		}), nil
	default:
		return nil, authflow.ErrIncompatibleInput
	}
}

func (i *IntentLoginFlowStepIdentify) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	return NewIdentificationData(IdentificationData{
		Options: i.Options,
	}), nil
}

func (i *IntentLoginFlowStepIdentify) currentFlowObject(deps *authflow.Dependencies) (config.AuthenticationFlowObject, error) {
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

func (*IntentLoginFlowStepIdentify) step(o config.AuthenticationFlowObject) *config.AuthenticationFlowLoginFlowStep {
	step, ok := o.(*config.AuthenticationFlowLoginFlowStep)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}

	return step
}

func (*IntentLoginFlowStepIdentify) checkIdentificationMethod(deps *authflow.Dependencies, step *config.AuthenticationFlowLoginFlowStep, im config.AuthenticationFlowIdentification) (idx int, err error) {
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

func (*IntentLoginFlowStepIdentify) identificationMethod(flows authflow.Flows) config.AuthenticationFlowIdentification {
	m, _, ok := authflow.FindMilestoneInCurrentFlow[MilestoneIdentificationMethod](flows)
	if !ok {
		panic(fmt.Errorf("identification method not yet selected"))
	}

	im := m.MilestoneIdentificationMethod()

	return im
}

func (i *IntentLoginFlowStepIdentify) jsonPointer(step *config.AuthenticationFlowLoginFlowStep, im config.AuthenticationFlowIdentification) jsonpointer.T {
	for idx, branch := range step.OneOf {
		branch := branch
		if branch.Identification == im {
			return authflow.JSONPointerForOneOf(i.JSONPointer, idx)
		}
	}

	panic(fmt.Errorf("selected identification method is not allowed"))
}
