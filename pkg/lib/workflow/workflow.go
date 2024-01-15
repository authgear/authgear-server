package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
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
	NodeSimple func(nodeSimple NodeSimple, w *Workflow) error
}

type Workflows struct {
	Root    *Workflow
	Nearest *Workflow
}

func NewWorkflows(root *Workflow) Workflows {
	return Workflows{
		Root:    root,
		Nearest: root,
	}
}

func (w Workflows) Replace(nearest *Workflow) Workflows {
	w.Nearest = nearest
	return w
}

func NewWorkflow(workflowID string, intent Intent) *Workflow {
	return &Workflow{
		WorkflowID: workflowID,
		InstanceID: newInstanceID(),
		Intent:     intent,
	}
}

// Accept executes the workflow to the deepest using input.
// In addition to the errors caused by intents and nodes,
// ErrEOF and ErrNoChange can be returned.
func (w *Workflow) Accept(ctx context.Context, deps *Dependencies, workflows Workflows, input Input) (err error) {
	var workflowInQuestion *Workflow
	var inputReactor InputReactor

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
		workflowInQuestion, inputReactor, err = w.FindInputReactor(ctx, deps, workflows.Replace(w))
		if err != nil {
			return
		}

		// Otherwise we found an InputReactor that we can feed input to.
		var nextNode *Node
		nextNode, err = inputReactor.ReactTo(ctx, deps, workflows.Replace(workflowInQuestion), input)

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
			if len(workflowInQuestion.Nodes) == 0 {
				panic(fmt.Errorf("input reactor %T returned ErrUpdateNode, but there are no nodes", inputReactor))
			}

			// Update the node inplace.
			workflowInQuestion.Nodes[len(workflowInQuestion.Nodes)-1] = *nodeToReplace

			// We have to stop and return here because this edge will react to this input indefinitely.
			return
		}

		// Handle other error.
		if err != nil {
			return
		}

		// We need to append the nextNode to the closest workflow.
		err = workflowInQuestion.appendNode(ctx, deps, workflows.Replace(workflowInQuestion), *nextNode)
		if err != nil {
			return
		}
		changed = true
	}
}

func (w *Workflow) appendNode(ctx context.Context, deps *Dependencies, workflows Workflows, node Node) error {
	w.Nodes = append(w.Nodes, node)

	err := node.Traverse(WorkflowTraverser{
		NodeSimple: func(nodeSimple NodeSimple, w *Workflow) error {
			effs, err := nodeSimple.GetEffects(ctx, deps, workflows.Replace(w))
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
	}, w)
	if err != nil {
		return err
	}

	return nil
}

func (w *Workflow) ApplyRunEffects(ctx context.Context, deps *Dependencies, workflows Workflows) error {
	err := w.Traverse(WorkflowTraverser{
		NodeSimple: func(nodeSimple NodeSimple, w *Workflow) error {
			effs, err := nodeSimple.GetEffects(ctx, deps, workflows.Replace(w))
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
			effs, err := intent.GetEffects(ctx, deps, workflows.Replace(w))
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

func (w *Workflow) ApplyAllEffects(ctx context.Context, deps *Dependencies, workflows Workflows) error {
	err := w.ApplyRunEffects(ctx, deps, workflows.Replace(w))
	if err != nil {
		return err
	}

	err = w.Traverse(WorkflowTraverser{
		NodeSimple: func(nodeSimple NodeSimple, w *Workflow) error {
			effs, err := nodeSimple.GetEffects(ctx, deps, workflows.Replace(w))
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
			effs, err := intent.GetEffects(ctx, deps, workflows.Replace(w))
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

func (w *Workflow) CollectCookies(ctx context.Context, deps *Dependencies, workflows Workflows) (cookies []*http.Cookie, err error) {
	err = w.Traverse(WorkflowTraverser{
		NodeSimple: func(nodeSimple NodeSimple, w *Workflow) error {
			if n, ok := nodeSimple.(CookieGetter); ok {
				c, err := n.GetCookies(ctx, deps, workflows.Replace(w))
				if err != nil {
					return err
				}
				cookies = append(cookies, c...)
			}

			return nil
		},
		Intent: func(intent Intent, w *Workflow) error {
			if i, ok := intent.(CookieGetter); ok {
				c, err := i.GetCookies(ctx, deps, workflows.Replace(w))
				if err != nil {
					return err
				}
				cookies = append(cookies, c...)
			}

			return nil
		},
	})
	if err != nil {
		return
	}

	return
}

func (w *Workflow) GetAuthenticationInfoEntry(ctx context.Context, deps *Dependencies, workflows Workflows) (*authenticationinfo.Entry, bool) {
	var e *authenticationinfo.Entry
	_ = w.Traverse(WorkflowTraverser{
		NodeSimple: func(nodeSimple NodeSimple, w *Workflow) error {
			if n, ok := nodeSimple.(AuthenticationInfoEntryGetter); ok {
				e = n.GetAuthenticationInfoEntry(ctx, deps, workflows)
			}
			return nil
		},
		Intent: func(intent Intent, w *Workflow) error {
			if i, ok := intent.(AuthenticationInfoEntryGetter); ok {
				e = i.GetAuthenticationInfoEntry(ctx, deps, workflows)
			}
			return nil
		},
	})
	if e != nil {
		return e, true
	}
	return nil, false
}

func (w *Workflow) Traverse(t WorkflowTraverser) error {
	for _, node := range w.Nodes {
		err := node.Traverse(t, w)
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

func (w *Workflow) FindInputReactor(ctx context.Context, deps *Dependencies, workflows Workflows) (*Workflow, InputReactor, error) {
	if len(w.Nodes) > 0 {
		// We check the last node if it can react to input first.
		lastNode := w.Nodes[len(w.Nodes)-1]
		workflow, inputReactor, err := lastNode.FindInputReactor(ctx, deps, workflows.Replace(w))
		if err == nil {
			return workflow, inputReactor, nil
		}
		// Return non ErrEOF error.
		if !errors.Is(err, ErrEOF) {
			return nil, nil, err
		}
		// err is ErrEOF, fallthrough
	}

	// Otherwise we check if the intent can react to input.
	_, err := w.Intent.CanReactTo(ctx, deps, workflows.Replace(w))
	if err == nil {
		return w, w.Intent, nil
	}

	// err != nil here.
	// Regardless of whether err is ErrEOF, we return err.
	return nil, nil, err
}

func (w *Workflow) IsEOF(ctx context.Context, deps *Dependencies, workflows Workflows) (bool, error) {
	_, _, err := w.FindInputReactor(ctx, deps, workflows.Replace(w))
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

	intent, err := InstantiateIntentFromPrivateRegistry(workflowJSON.Intent)
	if err != nil {
		return
	}

	w.WorkflowID = workflowJSON.WorkflowID
	w.InstanceID = workflowJSON.InstanceID
	w.Intent = intent
	w.Nodes = workflowJSON.Nodes
	return nil
}

func (w *Workflow) ToOutput(ctx context.Context, deps *Dependencies, workflows Workflows) (*WorkflowOutput, error) {
	output := &WorkflowOutput{
		WorkflowID: w.WorkflowID,
		InstanceID: w.InstanceID,
	}

	intentOutputData, err := w.Intent.OutputData(ctx, deps, workflows.Replace(w))
	if err != nil {
		return nil, err
	}
	output.Intent = IntentOutput{
		Kind: w.Intent.Kind(),
		Data: intentOutputData,
	}

	for _, node := range w.Nodes {
		nodeOutput, err := node.ToOutput(ctx, deps, workflows.Replace(w))
		if err != nil {
			return nil, err
		}
		output.Nodes = append(output.Nodes, *nodeOutput)
	}

	return output, nil
}

func FindSingleNode[T NodeSimple](w *Workflow) (T, bool) {
	for _, n := range w.Nodes {
		if n.Type != NodeTypeSimple {
			continue
		}
		if node, ok := n.Simple.(T); ok {
			return node, true
		}
	}

	return *new(T), false
}

func FindSubWorkflows[T Intent](w *Workflow) []*Workflow {
	var workflows []*Workflow
	for _, n := range w.Nodes {
		if n.Type != NodeTypeSubWorkflow {
			continue
		}
		if _, ok := n.SubWorkflow.Intent.(T); ok {
			workflows = append(workflows, n.SubWorkflow)
		}
	}

	return workflows
}

func MustFindSubWorkflow[T Intent](w *Workflow) (T, *Workflow) {
	workflows := FindSubWorkflows[T](w)
	if len(workflows) == 0 {
		panic(fmt.Sprintf("workflow: cannot find workflow %T", *new(T)))
	} else if len(workflows) > 1 {
		panic(fmt.Sprintf("workflow: multiple workflow %T found", *new(T)))
	}
	return workflows[0].Intent.(T), workflows[0]
}
