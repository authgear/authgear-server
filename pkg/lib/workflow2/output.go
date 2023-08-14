package workflow2

import (
	"context"
	"errors"
)

type DataOutputer interface {
	OutputData(ctx context.Context, deps *Dependencies, workflows Workflows) (interface{}, error)
}

type WorkflowOutput struct {
	WorkflowID string       `json:"workflow_id,omitempty"`
	InstanceID string       `json:"instance_id,omitempty"`
	Intent     IntentOutput `json:"intent"`
	Nodes      []NodeOutput `json:"nodes,omitempty"`
}

type IntentOutput struct {
	Kind string      `json:"kind"`
	Data interface{} `json:"data,omitempty"`
}

type NodeSimpleOutput struct {
	Kind string      `json:"kind"`
	Data interface{} `json:"data,omitempty"`
}

type NodeOutput struct {
	Type        NodeType          `json:"type"`
	Simple      *NodeSimpleOutput `json:"simple,omitempty"`
	SubWorkflow *WorkflowOutput   `json:"workflow,omitempty"`
}

func WorkflowToOutput(ctx context.Context, deps *Dependencies, workflows Workflows) (*WorkflowOutput, error) {
	output := &WorkflowOutput{
		WorkflowID: workflows.Nearest.WorkflowID,
		InstanceID: workflows.Nearest.InstanceID,
	}

	var intentOutputData interface{}
	var err error
	if outputer, ok := workflows.Nearest.Intent.(DataOutputer); ok {
		intentOutputData, err = outputer.OutputData(ctx, deps, workflows)
		if err != nil {
			return nil, err
		}
	}

	output.Intent = IntentOutput{
		Kind: workflows.Nearest.Intent.Kind(),
		Data: intentOutputData,
	}

	for _, node := range workflows.Nearest.Nodes {
		node := node
		nodeOutput, err := NodeToOutput(ctx, deps, workflows, &node)
		if err != nil {
			return nil, err
		}
		output.Nodes = append(output.Nodes, *nodeOutput)
	}

	return output, nil
}

func NodeToOutput(ctx context.Context, deps *Dependencies, workflows Workflows, n *Node) (*NodeOutput, error) {
	output := &NodeOutput{
		Type: n.Type,
	}

	switch n.Type {
	case NodeTypeSimple:
		var nodeSimpleData interface{}
		var err error
		if outputer, ok := n.Simple.(DataOutputer); ok {
			nodeSimpleData, err = outputer.OutputData(ctx, deps, workflows)
			if err != nil {
				return nil, err
			}
		}

		output.Simple = &NodeSimpleOutput{
			Kind: n.Simple.Kind(),
			Data: nodeSimpleData,
		}
		return output, nil
	case NodeTypeSubWorkflow:
		workflowOutput, err := WorkflowToOutput(ctx, deps, workflows.Replace(n.SubWorkflow))
		if err != nil {
			return nil, err
		}

		output.SubWorkflow = workflowOutput
		return output, nil
	default:
		panic(errors.New("unreachable"))
	}
}
