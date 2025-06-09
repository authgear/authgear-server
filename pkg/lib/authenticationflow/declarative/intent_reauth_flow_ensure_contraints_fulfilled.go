package declarative

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"
)

func init() {
	authflow.RegisterIntent(&IntentReauthFlowEnsureContraintsFulfilled{})
}

type IntentReauthFlowEnsureContraintsFulfilled struct {
	FlowObject    *config.AuthenticationFlowReauthFlowStep `json:"flow_object"`
	FlowReference authenticationflow.FlowReference         `json:"flow_reference,omitempty"`
	JSONPointer   jsonpointer.T                            `json:"json_pointer,omitempty"`
}

func NewIntentReauthFlowEnsureContraintsFulfilled(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows, flowRef authenticationflow.FlowReference) (*IntentReauthFlowEnsureContraintsFulfilled, error) {
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

	trueValue := true
	// Generate a temporary config for this step only
	flowObject := &config.AuthenticationFlowReauthFlowStep{
		Type:                             config.AuthenticationFlowReauthFlowStepTypeAuthenticate,
		OneOf:                            oneOfs,
		ShowUntilAMRConstraintsFulfilled: &trueValue,
	}

	return &IntentReauthFlowEnsureContraintsFulfilled{
		FlowReference: flowRef,
		FlowObject:    flowObject,
		JSONPointer:   jsonpointer.T{},
	}, nil
}

var _ authenticationflow.Intent = &IntentReauthFlowEnsureContraintsFulfilled{}
var _ authenticationflow.Milestone = &IntentReauthFlowEnsureContraintsFulfilled{}
var _ MilestoneAuthenticationFlowObjectProvider = &IntentReauthFlowEnsureContraintsFulfilled{}

func (*IntentReauthFlowEnsureContraintsFulfilled) Kind() string {
	return "IntentReauthFlowEnsureContraintsFulfilled"
}

func (i *IntentReauthFlowEnsureContraintsFulfilled) CanReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows) (authenticationflow.InputSchema, error) {
	switch len(flows.Nearest.Nodes) {
	case 0:
		return nil, nil
	case 1:
		return nil, authflow.ErrEOF
	}
	panic(fmt.Errorf("unexpected number of nodes"))
}

func (i *IntentReauthFlowEnsureContraintsFulfilled) ReactTo(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows, input authenticationflow.Input) (authenticationflow.ReactToResult, error) {
	stepAuthenticate, err := NewIntentReauthFlowStepAuthenticate(ctx, deps, flows, &IntentReauthFlowStepAuthenticate{
		FlowReference: i.FlowReference,
		StepName:      "",
		JSONPointer:   i.JSONPointer,
		UserID:        i.userID(flows),
	}, i)
	if err != nil {
		return nil, err
	}
	return authflow.NewSubFlow(stepAuthenticate), nil
}

func (*IntentReauthFlowEnsureContraintsFulfilled) userID(flows authflow.Flows) string {
	userID, err := getUserID(flows)
	if err != nil {
		panic(err)
	}
	return userID
}

func (i *IntentReauthFlowEnsureContraintsFulfilled) Milestone() {
	return
}

// This is needed so that the child authenticate intents display a correct flow action
func (i *IntentReauthFlowEnsureContraintsFulfilled) MilestoneAuthenticationFlowObjectProvider() config.AuthenticationFlowObject {
	return i.FlowObject
}
