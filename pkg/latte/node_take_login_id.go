package latte

import (
	"context"
	"encoding/json"

	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterInput(&InputTakeLoginID{})
}

var InputTakeLoginIDSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"login_id": { "type": "string" }
		},
		"required": ["login_id"]
	}
`)

type InputTakeLoginID struct {
	LoginID string `json:"login_id"`
}

func (*InputTakeLoginID) Kind() string {
	return "latte.InputTakeLoginID"
}

func (i *InputTakeLoginID) Instantiate(data json.RawMessage) error {
	return InputTakeLoginIDSchema.Validator().ParseJSONRawMessage(data, i)
}

func (i *InputTakeLoginID) GetLoginID() string {
	return i.LoginID
}

type inputTakeLoginID interface {
	GetLoginID() string
}

type EdgeTakeLoginID struct{}

func (e *EdgeTakeLoginID) Instantiate(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	var loginIDInput inputTakeLoginID
	ok := workflow.AsInput(input, &loginIDInput)
	if !ok {
		return nil, workflow.ErrIncompatibleInput
	}

	return workflow.NewNodeSimple(&NodeTakeLoginID{
		LoginID: loginIDInput.GetLoginID(),
	}), nil
}

type NodeTakeLoginID struct {
	LoginID string
}

func (n *NodeTakeLoginID) Kind() string {
	return "latte.TakeLoginID"
}

func (n *NodeTakeLoginID) GetEffects(ctx context.Context, deps *workflow.Dependencies) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (n *NodeTakeLoginID) DeriveEdges(ctx context.Context, deps *workflow.Dependencies) ([]workflow.Edge, error) {
	return nil, workflow.ErrEOF
}

func (n *NodeTakeLoginID) OutputData(ctx context.Context, deps *workflow.Dependencies) (interface{}, error) {
	return map[string]interface{}{
		"login_id": n.LoginID,
	}, nil
}
