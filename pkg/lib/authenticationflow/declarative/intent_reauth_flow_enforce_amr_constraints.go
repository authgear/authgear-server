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
	authflow.RegisterIntent(&IntentReauthFlowEnforceAMRConstraints{})
}

type IntentReauthFlowEnforceAMRConstraints struct {
	FlowObject    *config.AuthenticationFlowReauthFlowStep `json:"flow_object"`
	FlowReference authenticationflow.FlowReference         `json:"flow_reference,omitempty"`
	JSONPointer   jsonpointer.T                            `json:"json_pointer,omitempty"`
}

func NewIntentReauthFlowEnforceAMRConstraints(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows, flowRef authenticationflow.FlowReference) (*IntentReauthFlowEnforceAMRConstraints, error) {
	var oneOfs []*config.AuthenticationFlowReauthFlowOneOf

	addOneOf := func(am config.AuthenticationFlowAuthentication) {
		oneOf := &config.AuthenticationFlowReauthFlowOneOf{
			Authentication: am,
		}

		oneOfs = append(oneOfs, oneOf)
	}

	for _, authenticatorType := range *deps.Config.Authentication.SecondaryAuthenticators {
		switch authenticatorType {
		case model.AuthenticatorTypePassword:
			addOneOf(config.AuthenticationFlowAuthenticationSecondaryPassword)
		case model.AuthenticatorTypeOOBEmail:
			addOneOf(config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail)
		case model.AuthenticatorTypeOOBSMS:
			addOneOf(config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS)
		case model.AuthenticatorTypeTOTP:
			addOneOf(config.AuthenticationFlowAuthenticationSecondaryTOTP)
		case model.AuthenticatorTypePasskey:
			addOneOf(config.AuthenticationFlowAuthenticationPrimaryPasskey)
		}
	}

	flowObject := &config.AuthenticationFlowReauthFlowStep{
		Type:  config.AuthenticationFlowReauthFlowStepTypeAuthenticate,
		OneOf: oneOfs,
	}

	return &IntentReauthFlowEnforceAMRConstraints{
		FlowReference: flowRef,
		FlowObject:    flowObject,
		JSONPointer:   jsonpointer.T{},
	}, nil
}

var _ authenticationflow.Intent = &IntentReauthFlowEnforceAMRConstraints{}
var _ authenticationflow.Milestone = &IntentReauthFlowEnforceAMRConstraints{}
var _ MilestoneAuthenticationFlowObjectProvider = &IntentReauthFlowEnforceAMRConstraints{}

func (*IntentReauthFlowEnforceAMRConstraints) Kind() string {
	return "IntentReauthFlowEnforceAMRConstraints"
}

func (i *IntentReauthFlowEnforceAMRConstraints) CanReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows) (authenticationflow.InputSchema, error) {
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

func (i *IntentReauthFlowEnforceAMRConstraints) ReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows, input authenticationflow.Input) (authenticationflow.ReactToResult, error) {
	stepAuthenticate, err := NewIntentReauthFlowStepAuthenticate(ctx, deps, flows, &IntentReauthFlowStepAuthenticate{
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

func (*IntentReauthFlowEnforceAMRConstraints) userID(flows authflow.Flows) string {
	userID, err := getUserID(flows)
	if err != nil {
		panic(err)
	}
	return userID
}

func (i *IntentReauthFlowEnforceAMRConstraints) Milestone() {
	return
}

// This is needed so that the child authenticate intents display a correct flow action
func (i *IntentReauthFlowEnforceAMRConstraints) MilestoneAuthenticationFlowObjectProvider() config.AuthenticationFlowObject {
	return i.FlowObject
}
