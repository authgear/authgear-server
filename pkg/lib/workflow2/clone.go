package workflow2

import (
	"errors"
)

func CloneWorkflow(w *Workflow) *Workflow {
	nodes := make([]Node, len(w.Nodes))
	for i, node := range w.Nodes {
		node := node
		nodes[i] = *CloneNode(&node)
	}

	return &Workflow{
		WorkflowID: w.WorkflowID,
		InstanceID: "",
		Intent:     w.Intent,
		Nodes:      nodes,
	}
}

func CloneNode(n *Node) *Node {
	cloned := &Node{
		Type: n.Type,
	}

	switch n.Type {
	case NodeTypeSimple:
		cloned.Simple = n.Simple
	case NodeTypeSubWorkflow:
		clonedWorkflow := CloneWorkflow(n.SubWorkflow)
		cloned.SubWorkflow = clonedWorkflow
	default:
		panic(errors.New("unreachable"))
	}

	return cloned
}
