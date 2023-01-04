package workflow

import (
	"encoding/json"
	"errors"
	"fmt"
)

type workflowJSON struct {
	WorkflowID string     `json:"workflow_id,omitempty"`
	InstanceID string     `json:"instance_id,omitempty"`
	Intent     intentJSON `json:"intent"`
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

func NewWorkflow(intent Intent) *Workflow {
	return &Workflow{
		WorkflowID: newWorkflowID(),
		InstanceID: newInstanceID(),
		Intent:     intent,
	}
}

// Accept executes the workflow to the deepest using input.
// In addition to the errors caused by intent, nodes and edges,
// ErrEOF and ErrNoChange can be returned.
func (w *Workflow) Accept(ctx *Context, input interface{}) (err error) {
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
		workflowForTheEdges, edges, err = w.DeriveEdges(ctx)
		if err != nil {
			return
		}

		// Otherwise we have some edges that we can feed input to.
		var nextNode *Node
		for _, edge := range edges {
			nextNode, err = edge.Instantiate(ctx, workflowForTheEdges, input)

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
		err = workflowForTheEdges.appendNode(ctx, *nextNode)
		if err != nil {
			return
		}
		changed = true
		nextNode = nil
	}
}

func (w *Workflow) appendNode(ctx *Context, node Node) error {
	w.Nodes = append(w.Nodes, node)

	effs, err := node.GetEffects(ctx)
	if err != nil {
		return err
	}
	for _, eff := range effs {
		if runEff, ok := eff.(RunEffect); ok {
			err = applyRunEffect(ctx, runEff)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (w *Workflow) ApplyRunEffects(ctx *Context) error {
	for _, node := range w.Nodes {
		effs, err := node.GetEffects(ctx)
		if err != nil {
			return err
		}
		for _, eff := range effs {
			if runEff, ok := eff.(RunEffect); ok {
				err = applyRunEffect(ctx, runEff)
				if err != nil {
					return err
				}
			}
		}
	}

	// Intent cannot have run-effects.
	effs, err := w.Intent.GetEffects(ctx, w)
	if err != nil {
		return err
	}
	for _, eff := range effs {
		if _, ok := eff.(RunEffect); ok {
			panic(fmt.Errorf("%T has RunEffect, which is disallowed", w.Intent))
		}
	}

	return nil
}

func (w *Workflow) ApplyAllEffects(ctx *Context) error {
	err := w.ApplyRunEffects(ctx)
	if err != nil {
		return err
	}

	for _, node := range w.Nodes {
		effs, err := node.GetEffects(ctx)
		if err != nil {
			return err
		}
		for _, eff := range effs {
			if onCommitEff, ok := eff.(OnCommitEffect); ok {
				err = applyOnCommitEffect(ctx, onCommitEff)
				if err != nil {
					return err
				}
			}
		}
	}

	effs, err := w.Intent.GetEffects(ctx, w)
	if err != nil {
		return err
	}
	for _, eff := range effs {
		if onCommitEff, ok := eff.(OnCommitEffect); ok {
			err = applyOnCommitEffect(ctx, onCommitEff)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (w *Workflow) DeriveEdges(ctx *Context) (*Workflow, []Edge, error) {
	if len(w.Nodes) > 0 {
		// We ask the last node to derive edges first.
		lastNode := w.Nodes[len(w.Nodes)-1]
		workflow, edges, err := lastNode.DeriveEdges(ctx, w)
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
	edges, err := w.Intent.DeriveEdges(ctx, w)
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

	intentJSON := intentJSON{
		Kind: IntentKind(w.Intent),
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

	intent := InstantiateIntent(workflowJSON.Intent.Kind)
	err = json.Unmarshal(workflowJSON.Intent.Data, intent)
	if err != nil {
		return
	}

	w.WorkflowID = workflowJSON.WorkflowID
	w.InstanceID = workflowJSON.InstanceID
	w.Intent = intent
	w.Nodes = workflowJSON.Nodes
	return nil
}

func (w *Workflow) ToOutput(ctx *Context) (*WorkflowOutput, error) {
	output := &WorkflowOutput{
		WorkflowID: w.WorkflowID,
		InstanceID: w.InstanceID,
	}

	intentOutputData, err := w.Intent.OutputData(ctx, w)
	if err != nil {
		return nil, err
	}
	output.Intent = IntentOutput{
		Kind: IntentKind(w.Intent),
		Data: intentOutputData,
	}

	for _, node := range w.Nodes {
		nodeOutput, err := node.ToOutput(ctx)
		if err != nil {
			return nil, err
		}
		output.Nodes = append(output.Nodes, *nodeOutput)
	}

	return output, nil
}
