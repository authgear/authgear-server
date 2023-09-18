package authenticationflow

import (
	"fmt"
	"reflect"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
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
	FlowTypeSignup FlowType = "signup_flow"
	FlowTypeLogin  FlowType = "login_flow"
)

// FlowReference is an API object.
type FlowReference struct {
	Type FlowType `json:"type"`
	Name string   `json:"name"`
}

// FlowStep is an API object.
type FlowStep struct {
	Type           string                                  `json:"type"`
	Identification config.AuthenticationFlowIdentification `json:"identification,omitempty"`
	Authentication config.AuthenticationFlowAuthentication `json:"authentication,omitempty"`
}

// FlowResponse is an API object.
// When the flow finished, `json_schema` is absent and `finished` is true.
// When data contains "redirect_uri", the driver of the flow must perform redirect.
type FlowResponse struct {
	// ID is the instance ID.
	ID string `json:"id"`
	// WebsocketID is actually the flow ID.
	WebsocketID string `json:"websocket_id"`

	Finished   bool                     `json:"finished,omitempty"`
	JSONSchema validation.SchemaBuilder `json:"json_schema,omitempty"`

	FlowType FlowType `json:"flow_type,omitempty"`
	FlowName string   `json:"flow_name,omitempty"`

	FlowStep *FlowStep `json:"flow_step,omitempty"`

	Data Data `json:"data"`
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
