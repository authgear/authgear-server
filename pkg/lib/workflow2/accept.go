package workflow2

import (
	"context"
	"errors"
	"fmt"
)

// Accept executes the workflow to the deepest using input.
// In addition to the errors caused by intents and nodes,
// ErrEOF and ErrNoChange can be returned.
func Accept(ctx context.Context, deps *Dependencies, workflows Workflows, input Input) (err error) {
	var newWorkflows Workflows
	var inputReactor InputReactor

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
		newWorkflows, inputReactor, err = FindInputReactorForWorkflow(ctx, deps, workflows)
		if err != nil {
			return
		}

		// Otherwise we found an InputReactor that we can feed input to.
		var nextNode *Node
		nextNode, err = inputReactor.ReactTo(ctx, deps, newWorkflows, input)

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
			if len(newWorkflows.Nearest.Nodes) == 0 {
				panic(fmt.Errorf("input reactor %T returned ErrUpdateNode, but there are no nodes", inputReactor))
			}

			// Update the node inplace.
			newWorkflows.Nearest.Nodes[len(newWorkflows.Nearest.Nodes)-1] = *nodeToReplace

			// We have to stop and return here because this edge will react to this input indefinitely.
			return
		}

		// Handle other error.
		if err != nil {
			return
		}

		// We need to append the nextNode to the closest workflow.
		err = appendNode(ctx, deps, newWorkflows, *nextNode)
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
