package api

import (
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

type WorkflowResponse struct {
	Action   *workflow.WorkflowAction `json:"action"`
	Workflow *workflow.WorkflowOutput `json:"workflow"`
}
