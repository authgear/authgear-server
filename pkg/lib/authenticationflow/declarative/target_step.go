package declarative

import (
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

type FlowTargetStep interface {
	GetName() string
	GetJSONPointer() jsonpointer.T
}

func FindTargetStep(w *authflow.Flow, name string) (out *authflow.Flow, err error) {
	err = authflow.TraverseFlow(authflow.Traverser{
		Intent: func(i authflow.Intent, w *authflow.Flow) error {
			if i, ok := i.(FlowTargetStep); ok && name == i.GetName() {
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
