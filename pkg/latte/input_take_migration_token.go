package latte

import (
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputTakeMigrationToken{})
}

var InputTakeMigrationTokenSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"migration_token": { "type": "string" }
		},
		"required": ["migration_token"]
	}
`)

type InputTakeMigrationToken struct {
	MigrationToken string `json:"migration_token"`
}

func (*InputTakeMigrationToken) Kind() string {
	return "latte.InputTakeMigrationToken"
}

func (*InputTakeMigrationToken) JSONSchema() *validation.SimpleSchema {
	return InputTakeMigrationTokenSchema
}

func (i *InputTakeMigrationToken) GetMigrationToken() string {
	return i.MigrationToken
}

type inputTakeMigrationToken interface {
	GetMigrationToken() string
}

var _ inputTakeMigrationToken = &InputTakeMigrationToken{}
