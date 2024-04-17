package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/slice"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaTakeOOBOTPChannel struct {
	JSONPointer    jsonpointer.T
	FlowRootObject config.AuthenticationFlowObject
	Channels       []model.AuthenticatorOOBChannel
}

var _ authflow.InputSchema = &InputSchemaTakeOOBOTPChannel{}

func (s *InputSchemaTakeOOBOTPChannel) GetJSONPointer() jsonpointer.T {
	return s.JSONPointer
}

func (i *InputSchemaTakeOOBOTPChannel) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (s *InputSchemaTakeOOBOTPChannel) SchemaBuilder() validation.SchemaBuilder {
	b := validation.SchemaBuilder{}.
		Type(validation.TypeObject).
		Required("channel")
	b.Properties().Property("channel", validation.SchemaBuilder{}.
		Type(validation.TypeString).
		Enum(slice.Cast[model.AuthenticatorOOBChannel, interface{}](s.Channels)...))

	return b
}

func (s *InputSchemaTakeOOBOTPChannel) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputTakeOOBOTPChannel
	err := s.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputTakeOOBOTPChannel struct {
	Channel model.AuthenticatorOOBChannel `json:"channel,omitempty"`
}

var _ authflow.Input = &InputTakeOOBOTPChannel{}
var _ inputTakeOOBOTPChannel = &InputTakeOOBOTPChannel{}

func (*InputTakeOOBOTPChannel) Input() {}

func (i *InputTakeOOBOTPChannel) GetChannel() model.AuthenticatorOOBChannel {
	return i.Channel
}
