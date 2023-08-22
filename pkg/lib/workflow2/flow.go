package workflow2

import (
	"fmt"
	"reflect"

	"github.com/authgear/authgear-server/pkg/util/validation"
)

// Flow is a instantiable intent by the public.
type Flow interface {
	Intent
	FlowType() FlowType
	FlowInit(r FlowReference)
}

// FlowType denotes the type of the intents.
type FlowType string

const (
	FlowTypeSignup FlowType = "signup_flow"
	FlowTypeLogin  FlowType = "login_flow"
)

// FlowReference is an API object.
type FlowReference struct {
	Type FlowType `json:"type"`
	ID   string   `json:"id"`
}

// FlowResponse is an API object.
// When json_schema is present, this means the flow is not finished.
// When json_schema is absent, the flow is finished.
// When data contains "redirect_uri", the driver of the flow must perform redirect.
// A very common case is:
// 1. json_schema is absent
// 2. redirect_uri is present in data.
type FlowResponse struct {
	ID         string                   `json:"id"`
	JSONSchema validation.SchemaBuilder `json:"json_schema,omitempty"`
	Data       Data                     `json:"data"`
}

type flowFactory func() Flow

var flowRegistry = map[FlowType]flowFactory{}

// RegisterFlow is for registering a flow.
func RegisterFlow(flow Flow) {
	flowGoType := reflect.TypeOf(flow).Elem()

	flowType := flow.FlowType()
	factory := flowFactory(func() Flow {
		return reflect.New(flowGoType).Interface().(Flow)
	})

	if _, registered := flowRegistry[flowType]; registered {
		panic(fmt.Errorf("workflow: duplicated flow type: %v", flowType))
	}

	flowRegistry[flowType] = factory

	RegisterIntent(flow)
}

// InstantiateFlow is used by the HTTP layer to instantiate a Flow.
func InstantiateFlow(f FlowReference) (Flow, error) {
	factory, ok := flowRegistry[f.Type]
	if !ok {
		return nil, ErrUnknownFlow
	}

	flow := factory()
	flow.FlowInit(f)
	return flow, nil
}
