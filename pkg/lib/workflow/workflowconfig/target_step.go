package workflowconfig

import (
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

type WorkflowStep interface {
	GetID() string
	GetJSONPointer() jsonpointer.T
}

func FindTargetStep(w *workflow.Workflow, id string) (out *workflow.Workflow, err error) {
	err = w.Traverse(workflow.WorkflowTraverser{
		Intent: func(i workflow.Intent, w *workflow.Workflow) error {
			if i, ok := i.(WorkflowStep); ok && id == i.GetID() {
				out = w
			}
			return nil
		},
	})
	if err != nil {
		return
	}

	if out == nil {
		err = ErrStepNotFound
	}

	return
}
