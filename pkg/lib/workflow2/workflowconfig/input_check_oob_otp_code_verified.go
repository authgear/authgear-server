package workflowconfig

import (
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputCheckOOBOTPCodeVerified{})
}

var InputCheckOOBOTPCodeVerifiedSchema = validation.NewSimpleSchema(`
{
	"type": "object",
	"additionalProperties": false
}
`)

type InputCheckOOBOTPCodeVerified struct {
	Code string `json:"code,omitempty"`
}

func (*InputCheckOOBOTPCodeVerified) Kind() string {
	return "workflowconfig.InputCheckOOBOTPCodeVerified"
}

func (*InputCheckOOBOTPCodeVerified) JSONSchema() *validation.SimpleSchema {
	return InputCheckOOBOTPCodeVerifiedSchema
}

func (*InputCheckOOBOTPCodeVerified) CheckOOBOTPCodeVerified() {}

type inputCheckOOBOTPCodeVerified interface {
	CheckOOBOTPCodeVerified()
}

var _ inputCheckOOBOTPCodeVerified = &InputCheckOOBOTPCodeVerified{}
