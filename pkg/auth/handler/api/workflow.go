package api

import (
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

type WorkflowAction struct {
	Type        WorkflowActionType `json:"type"`
	RedirectURI string             `json:"redirect_uri,omitempty"`
}

type WorkflowActionType string

const (
	WorkflowActionTypeContinue WorkflowActionType = "continue"
	WorkflowActionTypeFinish   WorkflowActionType = "finish"
	WorkflowActionTypeRedirect WorkflowActionType = "redirect"
)

type WorkflowResponse struct {
	Action   WorkflowAction           `json:"action"`
	Workflow *workflow.WorkflowOutput `json:"workflow"`
}
