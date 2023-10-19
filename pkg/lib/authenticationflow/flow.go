package authenticationflow

import (
	"fmt"
	"reflect"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

// PublicFlow is a instantiable intent by the public.
type PublicFlow interface {
	Intent
	FlowType() FlowType
	FlowInit(r FlowReference)
	FlowFlowReference() FlowReference
	FlowRootObject(deps *Dependencies) (config.AuthenticationFlowObject, error)
}

// FlowType denotes the type of the intents.
type FlowType string

const (
	FlowTypeSignup      FlowType = "signup"
	FlowTypePromote     FlowType = "promote"
	FlowTypeLogin       FlowType = "login"
	FlowTypeSignupLogin FlowType = "signup_login"
)

var AllFlowTypes []FlowType = []FlowType{
	FlowTypeSignup,
	FlowTypePromote,
	FlowTypeLogin,
	FlowTypeSignupLogin,
}

// FlowReference is an API object.
type FlowReference struct {
	Type FlowType `json:"type"`
	Name string   `json:"name"`
}

type FlowActionType string

const (
	FlowActionTypeFinished FlowActionType = "finished"
)

func FlowActionTypeFromStepType(t config.AuthenticationFlowStepType) FlowActionType {
	return FlowActionType(t)
}

// FlowAction is an API object.
type FlowAction struct {
	Type           FlowActionType                          `json:"type"`
	Identification config.AuthenticationFlowIdentification `json:"identification,omitempty"`
	Authentication config.AuthenticationFlowAuthentication `json:"authentication,omitempty"`
	Data           Data                                    `json:"data,omitempty"`
}

// FlowResponse is an API object.
//
// When the flow finished, `finished` is true.
// In this case, `finish_redirect_uri` may be present.
type FlowResponse struct {
	// StateToken is the StateToken.
	StateToken string `json:"state_token"`

	// ID is the flow ID.
	ID   string   `json:"id"`
	Type FlowType `json:"type,omitempty"`
	Name string   `json:"name,omitempty"`

	Action *FlowAction `json:"action,omitempty"`
}

type flowFactory func() PublicFlow

var flowRegistry = map[FlowType]flowFactory{}

// RegisterFlow is for registering a flow.
func RegisterFlow(flow PublicFlow) {
	flowGoType := reflect.TypeOf(flow).Elem()

	flowType := flow.FlowType()
	factory := flowFactory(func() PublicFlow {
		return reflect.New(flowGoType).Interface().(PublicFlow)
	})

	if _, registered := flowRegistry[flowType]; registered {
		panic(fmt.Errorf("duplicated flow type: %v", flowType))
	}

	flowRegistry[flowType] = factory

	RegisterIntent(flow)
}

// InstantiateFlow is used by the HTTP layer to instantiate a Flow.
func InstantiateFlow(f FlowReference) (PublicFlow, error) {
	factory, ok := flowRegistry[f.Type]
	if !ok {
		return nil, ErrUnknownFlow
	}

	flow := factory()
	flow.FlowInit(f)
	return flow, nil
}
