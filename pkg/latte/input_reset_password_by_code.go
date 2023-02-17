package latte

import (
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputResetPasswordByCode{})
}

var InputResetPasswordByCodeSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"code": { "type": "string" },
			"new_password": { "type": "string" }
		},
		"required": ["code", "new_password"]
	}
`)

type InputResetPasswordByCode struct {
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

func (*InputResetPasswordByCode) Kind() string {
	return "latte.InputResetPasswordByCode"
}

func (*InputResetPasswordByCode) JSONSchema() *validation.SimpleSchema {
	return InputResetPasswordByCodeSchema
}

func (i *InputResetPasswordByCode) GetCode() string {
	return i.Code
}

func (i *InputResetPasswordByCode) GetNewPassword() string {
	return i.NewPassword
}

type inputResetPasswordByCode interface {
	GetCode() string
	GetNewPassword() string
}

var _ inputResetPasswordByCode = &InputResetPasswordByCode{}
