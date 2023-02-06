package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

type workflowJSON struct {
	WorkflowID string     `json:"workflow_id,omitempty"`
	InstanceID string     `json:"instance_id,omitempty"`
	Intent     IntentJSON `json:"intent"`
	Nodes      []Node     `json:"nodes,omitempty"`
}

type WorkflowOutput struct {
	WorkflowID string       `json:"workflow_id,omitempty"`
	InstanceID string       `json:"instance_id,omitempty"`
	Intent     IntentOutput `json:"intent"`
	Nodes      []NodeOutput `json:"nodes,omitempty"`
}

type Workflow struct {
	WorkflowID string
	InstanceID string
	Intent     Intent
	Nodes      []Node
}

type WorkflowTraverser struct {
	Intent     func(intent Intent, w *Workflow) error
	NodeSimple func(nodeSimple NodeSimple) error
}

func NewWorkflow(workflowID string, intent Intent) *Workflow {
	return &Workflow{
		WorkflowID: workflowID,
		InstanceID: newInstanceID(),
		Intent:     intent,
	}
}

// Accept executes the workflow to the deepest using input.
// In addition to the errors caused by intent, nodes and edges,
// ErrEOF and ErrNoChange can be returned.
func (w *Workflow) Accept(ctx context.Context, deps *Dependencies, input Input) (err error) {
	var workflowForTheEdges *Workflow
	var edges []Edge

	var changed bool
	defer func() {
		if changed {
			w.InstanceID = newInstanceID()
		}
		if !changed && err == nil {
			err = ErrNoChange
		}
	}()

	for {
		workflowForTheEdges, edges, err = w.DeriveEdges(ctx, deps)
		if err != nil {
			return
		}

		// Otherwise we have some edges that we can feed input to.
		var nextNode *Node
		for _, edge := range edges {
			nextNode, err = edge.Instantiate(ctx, deps, workflowForTheEdges, input)

			// Continue to check the next edge.
			if errors.Is(err, ErrIncompatibleInput) {
				err = nil
				continue
			}

			if errors.Is(err, ErrSameNode) {
				err = nil
				// We still consider the workflow has something changes.
				changed = true
				// We have to stop and return here because this edge will react to this input indefinitely.
				return
			}

			if errors.Is(err, ErrUpdateNode) {
				err = nil
				// We still consider the workflow has something changes.
				changed = true

				nodeToReplace := nextNode

				// precondition: ErrUpdateNode must be returned by edges that were derived by a node.
				if len(workflowForTheEdges.Nodes) == 0 {
					panic(fmt.Errorf("edge %T returned ErrUpdateNode, but the edge was not derived by a node", edge))
				}

				// Update the node inplace.
				workflowForTheEdges.Nodes[len(workflowForTheEdges.Nodes)-1] = *nodeToReplace

				// We have to stop and return here because this edge will react to this input indefinitely.
				return
			}

			if err != nil {
				return
			}

			// So we have a non-nil nextNode here.
			// We can break the loop.
			break
		}

		// No edges are followed, input is required.
		if nextNode == nil {
			return
		}

		// We need to append the nextNode to the closest workflow.
		err = workflowForTheEdges.appendNode(ctx, deps, *nextNode)
		if err != nil {
			return
		}
		changed = true
		nextNode = nil
	}
}

func (w *Workflow) appendNode(ctx context.Context, deps *Dependencies, node Node) error {
	w.Nodes = append(w.Nodes, node)

	err := node.Traverse(WorkflowTraverser{
		NodeSimple: func(nodeSimple NodeSimple) error {
			effs, err := nodeSimple.GetEffects(ctx, deps)
			if err != nil {
				return err
			}
			for _, eff := range effs {
				if runEff, ok := eff.(RunEffect); ok {
					err = applyRunEffect(ctx, deps, runEff)
					if err != nil {
						return err
					}
				}
			}
			return nil
		},
		// Intent cannot have run-effect.
		// We do not bother traversing intents here.
	})
	if err != nil {
		return err
	}

	return nil
}

func (w *Workflow) ApplyRunEffects(ctx context.Context, deps *Dependencies) error {
	err := w.Traverse(WorkflowTraverser{
		NodeSimple: func(nodeSimple NodeSimple) error {
			effs, err := nodeSimple.GetEffects(ctx, deps)
			if err != nil {
				return err
			}
			for _, eff := range effs {
				if runEff, ok := eff.(RunEffect); ok {
					err = applyRunEffect(ctx, deps, runEff)
					if err != nil {
						return err
					}
				}
			}
			return nil
		},
		Intent: func(intent Intent, w *Workflow) error {
			effs, err := intent.GetEffects(ctx, deps, w)
			if err != nil {
				return err
			}
			for _, eff := range effs {
				if _, ok := eff.(RunEffect); ok {
					// Intent cannot have run-effects.
					panic(fmt.Errorf("%T has RunEffect, which is disallowed", w.Intent))
				}
			}
			return nil
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (w *Workflow) ApplyAllEffects(ctx context.Context, deps *Dependencies) error {
	err := w.ApplyRunEffects(ctx, deps)
	if err != nil {
		return err
	}

	err = w.Traverse(WorkflowTraverser{
		NodeSimple: func(nodeSimple NodeSimple) error {
			effs, err := nodeSimple.GetEffects(ctx, deps)
			if err != nil {
				return err
			}
			for _, eff := range effs {
				if onCommitEff, ok := eff.(OnCommitEffect); ok {
					err = applyOnCommitEffect(ctx, deps, onCommitEff)
					if err != nil {
						return err
					}
				}
			}
			return nil
		},
		Intent: func(intent Intent, w *Workflow) error {
			effs, err := intent.GetEffects(ctx, deps, w)
			if err != nil {
				return err
			}
			for _, eff := range effs {
				if onCommitEff, ok := eff.(OnCommitEffect); ok {
					err = applyOnCommitEffect(ctx, deps, onCommitEff)
					if err != nil {
						return err
					}
				}
			}
			return nil
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (w *Workflow) Traverse(t WorkflowTraverser) error {
	for _, node := range w.Nodes {
		err := node.Traverse(t)
		if err != nil {
			return err
		}
	}
	if t.Intent != nil {
		err := t.Intent(w.Intent, w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *Workflow) DeriveEdges(ctx context.Context, deps *Dependencies) (*Workflow, []Edge, error) {
	if len(w.Nodes) > 0 {
		// We ask the last node to derive edges first.
		lastNode := w.Nodes[len(w.Nodes)-1]
		workflow, edges, err := lastNode.DeriveEdges(ctx, deps, w)
		if err == nil {
			if len(edges) == 0 {
				panic(fmt.Errorf("node %T derives no edges without error", lastNode))
			}
			return workflow, edges, nil
		}

		// err != nil here.
		if !errors.Is(err, ErrEOF) {
			return nil, nil, err
		}

		// err is ErrEOF, fallthrough.
	}

	// Otherwise we ask the intent to derive edges.
	edges, err := w.Intent.DeriveEdges(ctx, deps, w)
	if err == nil {
		if len(edges) == 0 {
			panic(fmt.Errorf("intent %T derives no edges without error", w.Intent))
		}
		return w, edges, nil
	}

	// err != nil here.
	// Regardless of whether err is ErrEOF, we return err.
	return nil, nil, err
}

func (w *Workflow) IsEOF(ctx context.Context, deps *Dependencies) (bool, error) {
	_, _, err := w.DeriveEdges(ctx, deps)
	if err != nil {
		if errors.Is(err, ErrEOF) {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

func (w *Workflow) Clone() *Workflow {
	nodes := make([]Node, len(w.Nodes))
	for i, node := range w.Nodes {
		nodes[i] = *node.Clone()
	}

	return &Workflow{
		WorkflowID: w.WorkflowID,
		InstanceID: "",
		Intent:     w.Intent,
		Nodes:      nodes,
	}
}

func (w *Workflow) MarshalJSON() ([]byte, error) {
	var err error

	intentBytes, err := json.Marshal(w.Intent)
	if err != nil {
		return nil, err
	}

	intentJSON := IntentJSON{
		Kind: w.Intent.Kind(),
		Data: intentBytes,
	}

	workflowJSON := workflowJSON{
		WorkflowID: w.WorkflowID,
		InstanceID: w.InstanceID,
		Intent:     intentJSON,
		Nodes:      w.Nodes,
	}

	return json.Marshal(workflowJSON)
}

func (w *Workflow) UnmarshalJSON(d []byte) (err error) {
	workflowJSON := workflowJSON{}
	// workflowJSON contains []Node.
	// json.Unmarshal will call UnmarshalJSON of Node for us.
	err = json.Unmarshal(d, &workflowJSON)
	if err != nil {
		return
	}

	intent, err := InstantiateIntent(workflowJSON.Intent)
	if err != nil {
		return
	}

	w.WorkflowID = workflowJSON.WorkflowID
	w.InstanceID = workflowJSON.InstanceID
	w.Intent = intent
	w.Nodes = workflowJSON.Nodes
	return nil
}

func (w *Workflow) ToOutput(ctx context.Context, deps *Dependencies) (*WorkflowOutput, error) {
	output := &WorkflowOutput{
		WorkflowID: w.WorkflowID,
		InstanceID: w.InstanceID,
	}

	intentOutputData, err := w.Intent.OutputData(ctx, deps, w)
	if err != nil {
		return nil, err
	}
	output.Intent = IntentOutput{
		Kind: w.Intent.Kind(),
		Data: intentOutputData,
	}

	for _, node := range w.Nodes {
		nodeOutput, err := node.ToOutput(ctx, deps)
		if err != nil {
			return nil, err
		}
		output.Nodes = append(output.Nodes, *nodeOutput)
	}

	return output, nil
}
