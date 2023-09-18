package authenticationflow

import (
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"
)

type TargetStep interface {
	GetName() string
	GetJSONPointer() jsonpointer.T
}

func FindTargetStep(w *Flow, name string) (out *Flow, err error) {
	err = TraverseFlow(Traverser{
		Intent: func(i Intent, w *Flow) error {
			if i, ok := i.(TargetStep); ok && name == i.GetName() {
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
