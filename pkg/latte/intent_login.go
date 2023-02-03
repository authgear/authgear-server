package latte

import (
	"context"
	"encoding/json"

	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterIntent(&IntentLogin{})
}

var IntentLoginSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"some_string": { "type": "string" }
		},
		"required": ["some_string"]
	}
`)

type IntentLogin struct {
	SomeString string `json:"some_string"`
}

func (*IntentLogin) Kind() string {
	return "latte.IntentLogin"
}

func (i *IntentLogin) Instantiate(data json.RawMessage) error {
	return IntentLoginSchema.Validator().ParseJSONRawMessage(data, i)
}

func (*IntentLogin) DeriveEdges(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Edge, error) {
	if len(w.Nodes) == 0 {
		return []workflow.Edge{
			&EdgeTakeLoginID{},
		}, nil
	}
	return nil, workflow.ErrEOF
}

func (*IntentLogin) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (i *IntentLogin) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	return map[string]interface{}{
		"some_string": i.SomeString,
	}, nil
}
