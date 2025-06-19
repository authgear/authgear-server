package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentSignupFlowEnforceAMRConstraints{})
}

type IntentSignupFlowEnforceAMRConstraints struct {
	UserID        string                                   `json:"user_id"`
	FlowObject    *config.AuthenticationFlowSignupFlowStep `json:"flow_object"`
	FlowReference authenticationflow.FlowReference         `json:"flow_reference,omitempty"`
	JSONPointer   jsonpointer.T                            `json:"json_pointer,omitempty"`
}

type IntentSignupFlowEnforceAMRConstraintsOptions struct {
	UserID        string
	FlowReference authenticationflow.FlowReference
}

func NewIntentSignupFlowEnforceAMRConstraints(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows, opts *IntentSignupFlowEnforceAMRConstraintsOptions) (*IntentSignupFlowEnforceAMRConstraints, error) {
	var oneOfs []*config.AuthenticationFlowSignupFlowOneOf
	recoveryCodeStep := &config.AuthenticationFlowSignupFlowStep{
		Type: config.AuthenticationFlowSignupFlowStepTypeViewRecoveryCode,
	}

	addOneOf := func(am model.AuthenticationFlowAuthentication, bpGetter func(*config.AppConfig) (*config.AuthenticationFlowBotProtection, bool)) {

		oneOf := &config.AuthenticationFlowSignupFlowOneOf{
			Authentication: am,
		}

		if bpGetter != nil {
			if bp, ok := bpGetter(deps.Config); ok {
				oneOf.BotProtection = bp
			}
		}

		if !*deps.Config.Authentication.RecoveryCode.Disabled {
			oneOf.Steps = append(oneOf.Steps, recoveryCodeStep)
		}

		oneOfs = append(oneOfs, oneOf)
	}

	for _, authenticatorType := range *deps.Config.Authentication.SecondaryAuthenticators {
		switch authenticatorType {
		case model.AuthenticatorTypePassword:
			addOneOf(model.AuthenticationFlowAuthenticationSecondaryPassword, nil)
		case model.AuthenticatorTypeOOBEmail:
			addOneOf(model.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail, getBotProtectionRequirementsOOBOTPEmail)
		case model.AuthenticatorTypeOOBSMS:
			addOneOf(model.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS, getBotProtectionRequirementsOOBOTPSMS)
		case model.AuthenticatorTypeTOTP:
			addOneOf(model.AuthenticationFlowAuthenticationSecondaryTOTP, nil)
		case model.AuthenticatorTypePasskey:
			// FIXME(tung): We don't have a step to force user create passkey at the moment
		}
	}

	// Generate a temporary config for this step only
	flowObject := &config.AuthenticationFlowSignupFlowStep{
		Type:  config.AuthenticationFlowSignupFlowStepTypeCreateAuthenticator,
		OneOf: oneOfs,
	}

	return &IntentSignupFlowEnforceAMRConstraints{
		UserID:        opts.UserID,
		FlowReference: opts.FlowReference,
		FlowObject:    flowObject,
		JSONPointer:   jsonpointer.T{},
	}, nil
}

var _ authenticationflow.Intent = &IntentSignupFlowEnforceAMRConstraints{}
var _ authenticationflow.Milestone = &IntentSignupFlowEnforceAMRConstraints{}
var _ MilestoneAuthenticationFlowObjectProvider = &IntentSignupFlowEnforceAMRConstraints{}

func (*IntentSignupFlowEnforceAMRConstraints) Kind() string {
	return "IntentSignupFlowEnforceAMRConstraints"
}

func (i *IntentSignupFlowEnforceAMRConstraints) CanReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows) (authenticationflow.InputSchema, error) {
	remainingAMRs, err := RemainingAMRConstraintsInFlow(ctx, deps, flows)
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

func (i *IntentSignupFlowEnforceAMRConstraints) ReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows, input authenticationflow.Input) (authenticationflow.ReactToResult, error) {
	stepCreateAuthenticator, err := NewIntentSignupFlowStepCreateAuthenticator(ctx, deps, flows, &IntentSignupFlowStepCreateAuthenticator{
		FlowReference: i.FlowReference,
		StepName:      "",
		JSONPointer:   nil,
		UserID:        i.UserID,
	}, i)
	if err != nil {
		return nil, err
	}

	remainingAMRs, err := RemainingAMRConstraintsInFlow(ctx, deps, flows)
	if err != nil {
		return nil, err
	}

	// The subflow should only contain options that can fulfill remaining amr
	newOptions := filterAMROptionsByAMRConstraint(stepCreateAuthenticator.Options, remainingAMRs)
	stepCreateAuthenticator.Options = newOptions

	// This step cannot be skipped to ensure amr constraints are all fulfilled
	stepCreateAuthenticator.CannotBeSkipped = true
	return authflow.NewSubFlow(stepCreateAuthenticator), nil
}

func (i *IntentSignupFlowEnforceAMRConstraints) Milestone() {
	return
}

// This is needed so that the child authenticate intents display a correct flow action
func (i *IntentSignupFlowEnforceAMRConstraints) MilestoneAuthenticationFlowObjectProvider() config.AuthenticationFlowObject {
	return i.FlowObject
}
