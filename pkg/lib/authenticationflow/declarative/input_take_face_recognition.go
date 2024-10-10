package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaTakeFaceRecognition struct {
	JSONPointer    jsonpointer.T
	FlowRootObject config.AuthenticationFlowObject
}

var _ authflow.InputSchema = &InputSchemaTakeFaceRecognition{}

func (i *InputSchemaTakeFaceRecognition) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaTakeFaceRecognition) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (i *InputSchemaTakeFaceRecognition) SchemaBuilder() validation.SchemaBuilder {
	inputTakeFaceRecognitionSchemaBuilder := validation.SchemaBuilder{}.
		Type(validation.TypeObject).
		Required("b64_image")

	inputTakeFaceRecognitionSchemaBuilder.Properties().Property(
		"b64_image",
		validation.SchemaBuilder{}.Type(validation.TypeString),
	)

	return inputTakeFaceRecognitionSchemaBuilder
}

func (i *InputSchemaTakeFaceRecognition) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputTakeFaceRecognition
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputTakeFaceRecognition struct {
	B64Image string `json:"b64_image"`
}

var _ authflow.Input = &InputTakeFaceRecognition{}
var _ inputTakeFaceRecognition = &InputTakeFaceRecognition{}

func (*InputTakeFaceRecognition) Input() {}

func (i *InputTakeFaceRecognition) GetB64Image() string {
	return i.B64Image
}
