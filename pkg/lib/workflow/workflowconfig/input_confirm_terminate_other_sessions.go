package workflowconfig

import (
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicInput(&InputConfirmTerminateOtherSessions{})
}

var InputConfirmTerminateOtherSessionsSchema = validation.NewSimpleSchema(`
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
	}
}
`)

type InputConfirmTerminateOtherSessions struct {
}

func (*InputConfirmTerminateOtherSessions) Kind() string {
	return "workflowconfig.InputConfirmTerminateOtherSessions"
}

func (*InputConfirmTerminateOtherSessions) JSONSchema() *validation.SimpleSchema {
	return InputConfirmTerminateOtherSessionsSchema
}

func (*InputConfirmTerminateOtherSessions) ConfirmTerminateOtherSessions() {}

type inputConfirmTerminateOtherSessions interface {
	ConfirmTerminateOtherSessions()
}

var _ inputConfirmTerminateOtherSessions = &InputConfirmTerminateOtherSessions{}
