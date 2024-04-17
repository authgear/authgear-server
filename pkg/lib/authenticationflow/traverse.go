package authenticationflow

import (
	"errors"
)

type Traverser struct {
	Intent     func(intent Intent, w *Flow) error
	NodeSimple func(nodeSimple NodeSimple, w *Flow) error
}

// TraverseFlow traverse the flow, and intent of the flow is treated as the last node of the flow
func TraverseFlow(t Traverser, w *Flow) error {
	return traverseFlow(t, w, false)
}

func TraverseNode(t Traverser, w *Flow, n *Node) error {
	return traverseNode(t, w, n, false)
}

// TraverseFlowIntentFirst is same as TraverseFlow,
// except that it ensures all nodes and intents must be traversed in the order they are inserted to the flow
// i.e. The intent will invoke the Traverser before the nodes belongs that intent
func TraverseFlowIntentFirst(t Traverser, w *Flow) error {
	return traverseFlow(t, w, true)
}

func TraverseNodeIntentFirst(t Traverser, w *Flow, n *Node) error {
	return traverseNode(t, w, n, true)
}

func traverseFlow(t Traverser, w *Flow, intentFirst bool) error {
	if t.Intent != nil && intentFirst {
		err := t.Intent(w.Intent, w)
		if err != nil {
			return err
		}
	}
	for _, node := range w.Nodes {
		node := node
		err := traverseNode(t, w, &node, intentFirst)
		if err != nil {
			return err
		}
	}
	if t.Intent != nil && !intentFirst {
		err := t.Intent(w.Intent, w)
		if err != nil {
			return err
		}
	}
	return nil
}

func traverseNode(t Traverser, w *Flow, n *Node, intentFirst bool) error {
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
		err := traverseFlow(t, n.SubFlow, intentFirst)
		if err != nil {
			return err
		}
		return nil
	default:
		panic(errors.New("unreachable"))
	}
}

func TraverseFlowReverse(t Traverser, w *Flow) error {
	for i := len(w.Nodes) - 1; i >= 0; i-- {
		node := &w.Nodes[i]
		traverseNodeReverse(t, w, node)
	}

	if t.Intent != nil {
		err := t.Intent(w.Intent, w)
		if err != nil {
			return err
		}
	}
	return nil
}

func traverseNodeReverse(t Traverser, w *Flow, n *Node) error {
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
		err := TraverseFlowReverse(t, n.SubFlow)
		if err != nil {
			return err
		}
		return nil
	default:
		panic(errors.New("unreachable"))
	}
}
