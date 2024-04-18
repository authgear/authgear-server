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

type NodeOrIntent interface {
	Kinder
}

type IntentTraverser = func(intent Intent) error

func TraverseIntentFromEndToRoot(t IntentTraverser, w *Flow) error {
	// First, construct a node to parent mapping
	parentMap := map[NodeOrIntent]Intent{}
	var traverse func(w *Flow)
	traverse = func(w *Flow) {
		for _, node := range w.Nodes {
			node := node
			switch node.Type {
			case NodeTypeSimple:
				parentMap[node.Simple] = w.Intent
			case NodeTypeSubFlow:
				parentMap[node.SubFlow.Intent] = w.Intent
				traverse(node.SubFlow)
			default:
				panic(errors.New("unreachable"))
			}
		}
	}
	traverse(w)

	var findLastNodeInFlow func(w *Flow) NodeOrIntent
	findLastNodeInFlow = func(w *Flow) NodeOrIntent {
		// 1. Special case, no nodes.
		if len(w.Nodes) == 0 {
			return w.Intent
		}

		// 2. At least one node
		var lastNode *Node = &w.Nodes[len(w.Nodes)-1]
		switch lastNode.Type {
		case NodeTypeSimple:
			return lastNode.Simple
		case NodeTypeSubFlow:
			return findLastNodeInFlow(lastNode.SubFlow)
		}
		panic(errors.New("unreachable"))
	}

	lastNode := findLastNodeInFlow(w)
	if lastNode, ok := lastNode.(Intent); ok && lastNode != nil {
		err := t(lastNode)
		if err != nil {
			return err
		}
	}

	var recursivelyTraverseParent func(n NodeOrIntent) error
	recursivelyTraverseParent = func(n NodeOrIntent) error {
		parent := parentMap[n]
		if parent != nil {
			err := t(parent)
			if err != nil {
				return err
			}
			return recursivelyTraverseParent(parent)
		}
		return nil
	}
	return recursivelyTraverseParent(lastNode)
}
