package workflowconfig

import (
	"encoding/json"

	"github.com/authgear/authgear-server/pkg/api/model"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
	"github.com/authgear/authgear-server/pkg/util/slice"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaTakeOOBOTPChannel struct {
	Channels []model.AuthenticatorOOBChannel
}

var _ workflow.InputSchema = &InputSchemaTakeOOBOTPChannel{}

func (s *InputSchemaTakeOOBOTPChannel) SchemaBuilder() validation.SchemaBuilder {
	b := validation.SchemaBuilder{}.
		Type(validation.TypeObject).
		Required("channel")
	b.Properties().Property("channel", validation.SchemaBuilder{}.
		Type(validation.TypeString).
		Enum(slice.Cast[model.AuthenticatorOOBChannel, interface{}](s.Channels)...))

	return b
}

func (s *InputSchemaTakeOOBOTPChannel) MakeInput(rawMessage json.RawMessage) (workflow.Input, error) {
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

var _ workflow.Input = &InputTakeOOBOTPChannel{}
var _ inputTakeOOBOTPChannel = &InputTakeOOBOTPChannel{}

func (*InputTakeOOBOTPChannel) Input() {}

func (i *InputTakeOOBOTPChannel) GetChannel() model.AuthenticatorOOBChannel {
	return i.Channel
}
