package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaAccountLinkingIdentification struct {
	JSONPointer    jsonpointer.T
	FlowRootObject config.AuthenticationFlowObject
	Options        []AccountLinkingIdentificationOption
}

var _ authflow.InputSchema = &InputSchemaAccountLinkingIdentification{}

func (i *InputSchemaAccountLinkingIdentification) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaAccountLinkingIdentification) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (i *InputSchemaAccountLinkingIdentification) SchemaBuilder() validation.SchemaBuilder {
	b := validation.SchemaBuilder{}
	indices := []interface{}{}
	for idx := range i.Options {
		indices = append(indices, idx)
	}
	b.Properties().Property("index", validation.SchemaBuilder{}.Type(validation.TypeInteger).Enum(indices...))
	b.Required("index")

	return b
}

func (i *InputSchemaAccountLinkingIdentification) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputAccountLinkingIdentification
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputAccountLinkingIdentification struct {
	Index int `json:"index,omitempty"`
}

var _ authflow.Input = &InputAccountLinkingIdentification{}
var _ inputTakeAccountLinkingIdentificationIndex = &InputAccountLinkingIdentification{}

func (*InputAccountLinkingIdentification) Input() {}

func (i *InputAccountLinkingIdentification) GetAccountLinkingIdentificationIndex() int {
	return i.Index
}
