package declarative

import (
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

type FlowStep interface {
	GetID() string
	GetJSONPointer() jsonpointer.T
}

func FindTargetStep(w *authflow.Flow, id string) (out *authflow.Flow, err error) {
	err = authflow.TraverseFlow(authflow.Traverser{
		Intent: func(i authflow.Intent, w *authflow.Flow) error {
			if i, ok := i.(FlowStep); ok && id == i.GetID() {
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
