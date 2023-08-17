package workflow2

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/util/validation"
)

// Boundary confines the input in Accept.
// Boundary is identified by a string.
// Initially boundary is an empty string.
// Accept records the last boundary it saw.
// When Accept detects a different boundary,
// it stops as if the input reactor does not react to the input.
type Boundary interface {
	InputReactor
	Boundary() string
}

// Accept executes the workflow to the deepest using input.
// In addition to the errors caused by intents and nodes,
// ErrEOF and ErrNoChange can be returned.
func Accept(ctx context.Context, deps *Dependencies, workflows Workflows, rawMessage json.RawMessage) (err error) {
	var lastSeenBoundary string

	var changed bool
	defer func() {
		if changed {
			workflows.Nearest.InstanceID = newInstanceID()
		}
		if !changed && err == nil {
			err = ErrNoChange
		}
	}()

	for {
		var findInputReactorResult *FindInputReactorResult
		findInputReactorResult, err = FindInputReactor(ctx, deps, workflows)
		if err != nil {
			return
		}

		if findInputReactorResult.Boundary != nil {
			b := findInputReactorResult.Boundary.Boundary()
			if lastSeenBoundary == "" {
				lastSeenBoundary = b
			} else if lastSeenBoundary != b {
				// Boundary cross detected.
				// End the loop.
				return
			}
		}

		// Otherwise we found an InputReactor that we can feed input to.

		// input by default is nil.
		var input Input
		if rawMessage != nil && findInputReactorResult.InputSchema != nil {
			input, err = findInputReactorResult.InputSchema.MakeInput(rawMessage)
			// As a special case, if this loop has iterated at least once,
			// then we treat the validation error as ErrIncompatibleInput,
			// by setting err = nil.
			var valiationError *validation.AggregatedError
			if errors.As(err, &valiationError) {
				if changed {
					err = nil
				}
				return
			} else if err != nil {
				return
			}
		}

		var nextNode *Node
		nextNode, err = findInputReactorResult.InputReactor.ReactTo(ctx, deps, findInputReactorResult.Workflows, input)

		// Handle err == ErrIncompatibleInput
		if errors.Is(err, ErrIncompatibleInput) {
			err = nil
			// Since this is a forever loop, changed may have been set to true already.
			return
		}

		// Handle err == ErrSameNode
		if errors.Is(err, ErrSameNode) {
			err = nil
			// We still consider the workflow has something changes.
			changed = true
			// We have to stop and return here because this input reactor will react to this input indefinitely.
			return
		}

		// Handle err == ErrSameNode
		if errors.Is(err, ErrUpdateNode) {
			err = nil
			// We still consider the workflow has something changes.
			changed = true

			nodeToReplace := nextNode

			// precondition: ErrUpdateNode requires at least one node.
			if len(findInputReactorResult.Workflows.Nearest.Nodes) == 0 {
				panic(fmt.Errorf("input reactor %T returned ErrUpdateNode, but there are no nodes", findInputReactorResult.InputReactor))
			}

			// Update the node inplace.
			findInputReactorResult.Workflows.Nearest.Nodes[len(findInputReactorResult.Workflows.Nearest.Nodes)-1] = *nodeToReplace

			// We have to stop and return here because this edge will react to this input indefinitely.
			return
		}

		// Handle other error.
		if err != nil {
			return
		}

		// We need to append the nextNode to the closest workflow.
		err = appendNode(ctx, deps, findInputReactorResult.Workflows, *nextNode)
		if err != nil {
			return
		}
		changed = true
	}
}

func appendNode(ctx context.Context, deps *Dependencies, workflows Workflows, node Node) error {
	workflows.Nearest.Nodes = append(workflows.Nearest.Nodes, node)

	err := TraverseNode(WorkflowTraverser{
		NodeSimple: func(nodeSimple NodeSimple, w *Workflow) error {
			effectGetter, ok := nodeSimple.(EffectGetter)
			if !ok {
				return nil
			}

			effs, err := effectGetter.GetEffects(ctx, deps, workflows.Replace(w))
			if err != nil {
				return err
			}
			for _, eff := range effs {
				if runEff, ok := eff.(RunEffect); ok {
					err = runEff.doNotCallThisDirectly(ctx, deps)
					if err != nil {
						return err
					}
				}
			}
			return nil
		},
		// Intent cannot have run-effect.
		// We do not bother traversing intents here.
	}, workflows.Nearest, &node)
	if err != nil {
		return err
	}

	return nil
}
