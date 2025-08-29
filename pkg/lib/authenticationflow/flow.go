package authenticationflow

import (
	"fmt"
	"reflect"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

// PublicFlow is a instantiable intent by the public.
type PublicFlow interface {
	Intent
	FlowType() FlowType
	FlowInit(r FlowReference, startFrom jsonpointer.T)
	FlowFlowReference() FlowReference
	FlowRootObject(deps *Dependencies) (config.AuthenticationFlowObject, error)
}

// FlowType denotes the type of the intents.
type FlowType string

const (
	FlowTypeSignup          FlowType = "signup"
	FlowTypePromote         FlowType = "promote"
	FlowTypeLogin           FlowType = "login"
	FlowTypeSignupLogin     FlowType = "signup_login"
	FlowTypeReauth          FlowType = "reauth"
	FlowTypeAccountRecovery FlowType = "account_recovery"
)

var AllFlowTypes []FlowType = []FlowType{
	FlowTypeSignup,
	FlowTypePromote,
	FlowTypeLogin,
	FlowTypeSignupLogin,
	FlowTypeReauth,
	FlowTypeAccountRecovery,
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
	Type           FlowActionType                         `json:"type"`
	Identification model.AuthenticationFlowIdentification `json:"identification,omitempty"`
	Authentication model.AuthenticationFlowAuthentication `json:"authentication,omitempty"`
	Data           Data                                   `json:"data,omitempty"`
}

// FlowResponse is an API object.
type FlowResponse struct {
	StateToken string      `json:"state_token"`
	Type       FlowType    `json:"type,omitempty"`
	Name       string      `json:"name,omitempty"`
	Action     *FlowAction `json:"action,omitempty"`
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
func InstantiateFlow(f FlowReference, startFrom jsonpointer.T) (PublicFlow, error) {
	factory, ok := flowRegistry[f.Type]
	if !ok {
		return nil, ErrUnknownFlow
	}

	flow := factory()
	flow.FlowInit(f, startFrom)
	return flow, nil
}

func FindCurrentFlowReference(flow *Flow) *FlowReference {
	var ref *FlowReference = nil
	_ = TraverseIntentFromEndToRoot(func(intent Intent) error {
		// We only want the first one
		if ref != nil {
			return nil
		}
		if f, ok := intent.(PublicFlow); ok {
			thisref := f.FlowFlowReference()
			ref = &thisref
		}
		return nil
	}, flow)
	return ref
}

// FindStepByType recursively finds the first step in the flow that matches the given stepType
// and returns its JSON pointer.
func FindStepByType(flowRoot config.AuthenticationFlowObject, stepType config.AuthenticationFlowStepType) (jsonpointer.T, bool) {
	var result jsonpointer.T
	var found bool
	var visit func(obj config.AuthenticationFlowObject, p jsonpointer.T)

	visit = func(obj config.AuthenticationFlowObject, p jsonpointer.T) {
		if found {
			return
		}

		switch o := obj.(type) {
		case config.AuthenticationFlowObjectFlowStep:
			if o.GetType() == stepType {
				found = true
				result = p
				return
			}
			for i, oneOf := range o.GetOneOf() {
				oneOfPointer := JSONPointerForOneOf(p, i)
				visit(oneOf, oneOfPointer)
			}
		case config.AuthenticationFlowObjectFlowBranch:
			for i, step := range o.GetSteps() {
				stepPointer := JSONPointerForStep(p, i)
				visit(step, stepPointer)
			}
		case config.AuthenticationFlowStepsObject:
			for i, step := range o.GetSteps() {
				stepPointer := JSONPointerForStep(p, i)
				visit(step, stepPointer)
			}
		}
	}

	visit(flowRoot, jsonpointer.T{})
	return result, found
}
