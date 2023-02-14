package latte

import (
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputCheckLoginLinkVerified{})
}

var InputCheckLoginLinkVerifiedSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false
	}
`)

type InputCheckLoginLinkVerified struct {
}

func (*InputCheckLoginLinkVerified) Kind() string {
	return "latte.InputCheckLoginLinkVerified"
}

func (*InputCheckLoginLinkVerified) JSONSchema() *validation.SimpleSchema {
	return InputCheckLoginLinkVerifiedSchema
}

func (i *InputCheckLoginLinkVerified) CheckLoginLinkVerified() {}

type inputCheckLoginLinkVerified interface {
	CheckLoginLinkVerified()
}

var _ inputCheckLoginLinkVerified = &InputCheckLoginLinkVerified{}
