package workflow2

import (
	"errors"
)

type WorkflowTraverser struct {
	Intent     func(intent Intent, w *Workflow) error
	NodeSimple func(nodeSimple NodeSimple, w *Workflow) error
}

func TraverseWorkflow(t WorkflowTraverser, w *Workflow) error {
	for _, node := range w.Nodes {
		node := node
		err := TraverseNode(t, w, &node)
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

func TraverseNode(t WorkflowTraverser, w *Workflow, n *Node) error {
	switch n.Type {
	case NodeTypeSimple:
		if t.NodeSimple != nil {
			err := t.NodeSimple(n.Simple, w)
			if err != nil {
				return err
			}
		}
		return nil
	case NodeTypeSubWorkflow:
		err := TraverseWorkflow(t, n.SubWorkflow)
		if err != nil {
			return err
		}
		return nil
	default:
		panic(errors.New("unreachable"))
	}
}
