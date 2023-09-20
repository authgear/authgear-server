package authenticationflow

import (
	"errors"
)

func CloneFlow(w *Flow) *Flow {
	nodes := make([]Node, len(w.Nodes))
	for i, node := range w.Nodes {
		node := node
		nodes[i] = *CloneNode(&node)
	}

	return &Flow{
		FlowID:  w.FlowID,
		StateID: "",
		Intent:  w.Intent,
		Nodes:   nodes,
	}
}

func CloneNode(n *Node) *Node {
	cloned := &Node{
		Type: n.Type,
	}

	switch n.Type {
	case NodeTypeSimple:
		cloned.Simple = n.Simple
	case NodeTypeSubFlow:
		clonedFlow := CloneFlow(n.SubFlow)
		cloned.SubFlow = clonedFlow
	default:
		panic(errors.New("unreachable"))
	}

	return cloned
}
