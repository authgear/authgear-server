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
	authflow.RegisterIntent(&IntentLoginFlowEnforceAMRConstraints{})
}

type IntentLoginFlowEnforceAMRConstraints struct {
	FlowObject    *config.AuthenticationFlowLoginFlowStep `json:"flow_object"`
	FlowReference authenticationflow.FlowReference        `json:"flow_reference,omitempty"`
	JSONPointer   jsonpointer.T                           `json:"json_pointer,omitempty"`
}

func NewIntentLoginFlowEnforceAMRConstraints(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows, flowRef authenticationflow.FlowReference) (*IntentLoginFlowEnforceAMRConstraints, error) {
	var oneOfs []*config.AuthenticationFlowLoginFlowOneOf

	addOneOf := func(am model.AuthenticationFlowAuthentication, bpGetter func(*config.AppConfig) (*config.AuthenticationFlowBotProtection, bool)) {
		oneOf := &config.AuthenticationFlowLoginFlowOneOf{
			Authentication: am,
		}
		if bpGetter != nil {
			if bp, ok := bpGetter(deps.Config); ok {
				oneOf.BotProtection = bp
			}
		}

		oneOfs = append(oneOfs, oneOf)
	}

	for _, authenticatorType := range *deps.Config.Authentication.SecondaryAuthenticators {
		switch authenticatorType {
		case model.AuthenticatorTypePassword:
			addOneOf(model.AuthenticationFlowAuthenticationSecondaryPassword, getBotProtectionRequirementsPassword)
		case model.AuthenticatorTypeOOBEmail:
			addOneOf(model.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail, getBotProtectionRequirementsOOBOTPEmail)
		case model.AuthenticatorTypeOOBSMS:
			addOneOf(model.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS, getBotProtectionRequirementsOOBOTPSMS)
		case model.AuthenticatorTypeTOTP:
			addOneOf(model.AuthenticationFlowAuthenticationSecondaryTOTP, nil)
		case model.AuthenticatorTypePasskey:
			addOneOf(model.AuthenticationFlowAuthenticationPrimaryPasskey, nil)
		}
	}

	if !*deps.Config.Authentication.RecoveryCode.Disabled {
		oneOfs = append(oneOfs, &config.AuthenticationFlowLoginFlowOneOf{
			Authentication: model.AuthenticationFlowAuthenticationRecoveryCode,
		})
	}

	// Generate a temporary config for this step only
	flowObject := &config.AuthenticationFlowLoginFlowStep{
		Type:  config.AuthenticationFlowLoginFlowStepTypeAuthenticate,
		OneOf: oneOfs,
	}

	return &IntentLoginFlowEnforceAMRConstraints{
		FlowReference: flowRef,
		FlowObject:    flowObject,
		JSONPointer:   jsonpointer.T{},
	}, nil
}

var _ authenticationflow.Intent = &IntentLoginFlowEnforceAMRConstraints{}
var _ authenticationflow.Milestone = &IntentLoginFlowEnforceAMRConstraints{}
var _ MilestoneAuthenticationFlowObjectProvider = &IntentLoginFlowEnforceAMRConstraints{}

func (*IntentLoginFlowEnforceAMRConstraints) Kind() string {
	return "IntentLoginFlowEnforceAMRConstraints"
}

func (i *IntentLoginFlowEnforceAMRConstraints) CanReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows) (authenticationflow.InputSchema, error) {
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

func (i *IntentLoginFlowEnforceAMRConstraints) ReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows, input authenticationflow.Input) (authenticationflow.ReactToResult, error) {
	stepAuthenticate, err := NewIntentLoginFlowStepAuthenticate(ctx, deps, flows, &IntentLoginFlowStepAuthenticate{
		FlowReference: i.FlowReference,
		StepName:      "",
		JSONPointer:   nil,
		UserID:        i.userID(flows),
	}, i)
	if err != nil {
		return nil, err
	}
	remainingAMRs, err := RemainingAMRConstraintsInFlow(ctx, deps, flows)
	if err != nil {
		return nil, err
	}

	// The subflow should only contain options that can fulfill remaining amr
	newOptions := filterAMROptionsByAMRConstraint(stepAuthenticate.Options, remainingAMRs)
	stepAuthenticate.Options = newOptions
	return authflow.NewSubFlow(stepAuthenticate), nil
}

func (*IntentLoginFlowEnforceAMRConstraints) userID(flows authflow.Flows) string {
	userID, err := getUserID(flows)
	if err != nil {
		panic(err)
	}
	return userID
}

func (i *IntentLoginFlowEnforceAMRConstraints) Milestone() {
	return
}

// This is needed so that the child authenticate intents display a correct flow action
func (i *IntentLoginFlowEnforceAMRConstraints) MilestoneAuthenticationFlowObjectProvider() config.AuthenticationFlowObject {
	return i.FlowObject
}
