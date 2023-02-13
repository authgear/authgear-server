package latte

import (
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputSelectEmailLoginLink{})
}

var InputSelectEmailLoginLinkSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false
	}
`)

type InputSelectEmailLoginLink struct{}

func (*InputSelectEmailLoginLink) Kind() string {
	return "latte.InputSelectEmailLoginLink"
}

func (*InputSelectEmailLoginLink) JSONSchema() *validation.SimpleSchema {
	return InputSelectEmailLoginLinkSchema
}

func (i *InputSelectEmailLoginLink) SelectEmailLoginLink() {}

type inputSelectEmailLoginLink interface {
	SelectEmailLoginLink()
}

var _ inputSelectEmailLoginLink = &InputSelectEmailLoginLink{}
