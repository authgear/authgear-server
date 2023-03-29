package latte

import (
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputTakeCaptchaToken{})
}

var InputTakeCaptchaTokenSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"token": { "type": "string" }
		},
		"required": ["token"]
	}
`)

type InputTakeCaptchaToken struct {
	Token string `json:"token"`
}

func (*InputTakeCaptchaToken) Kind() string {
	return "latte.InputTakeCaptchaToken"
}

func (*InputTakeCaptchaToken) JSONSchema() *validation.SimpleSchema {
	return InputTakeCaptchaTokenSchema
}

func (i *InputTakeCaptchaToken) GetToken() string {
	return i.Token
}

type inputTakeCaptchaToken interface {
	GetToken() string
}

var _ inputTakeCaptchaToken = &InputTakeCaptchaToken{}
