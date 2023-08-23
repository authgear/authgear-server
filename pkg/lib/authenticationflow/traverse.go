package authenticationflow

import (
	"errors"
)

type Traverser struct {
	Intent     func(intent Intent, w *Flow) error
	NodeSimple func(nodeSimple NodeSimple, w *Flow) error
}

func TraverseFlow(t Traverser, w *Flow) error {
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

func TraverseNode(t Traverser, w *Flow, n *Node) error {
	switch n.Type {
	case NodeTypeSimple:
		if t.NodeSimple != nil {
			err := t.NodeSimple(n.Simple, w)
			if err != nil {
				return err
			}
		}
		return nil
	case NodeTypeSubFlow:
		err := TraverseFlow(t, n.SubFlow)
		if err != nil {
			return err
		}
		return nil
	default:
		panic(errors.New("unreachable"))
	}
}
