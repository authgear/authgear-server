package authenticationflow

import (
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"
)

type TargetStep interface {
	GetName() string
	GetJSONPointer() jsonpointer.T
}

func FindTargetStep(w *Flow, name string) (out *Flow, err error) {
	// Use TraverseFlowIntentFirst here to ensure the intent was found before the nodes belongs created by that intent
	// Because if there are two steps with a same name, the later node should be returned by this function
	// And nodes inside the intent should be considered "later" than the intent itself
	err = TraverseFlowIntentFirst(Traverser{
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
