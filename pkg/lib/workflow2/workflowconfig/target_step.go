package workflowconfig

import (
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

type WorkflowStep interface {
	GetID() string
	GetJSONPointer() jsonpointer.T
}

func FindTargetStep(w *workflow.Workflow, id string) (out *workflow.Workflow, err error) {
	err = workflow.TraverseWorkflow(workflow.WorkflowTraverser{
		Intent: func(i workflow.Intent, w *workflow.Workflow) error {
			if i, ok := i.(WorkflowStep); ok && id == i.GetID() {
				out = w
			}
			return nil
		},
	}, w)
	if err != nil {
		return
	}

	if out == nil {
		err = ErrStepNotFound
	}

	return
}
