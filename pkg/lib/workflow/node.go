package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

type nodeJSON struct {
	Type        NodeType        `json:"type"`
	Simple      *nodeSimpleJSON `json:"simple,omitempty"`
	SubWorkflow *Workflow       `json:"workflow,omitempty"`
}

type nodeSimpleJSON struct {
	Kind string          `json:"kind"`
	Data json.RawMessage `json:"data"`
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

type NodeType string

const (
	NodeTypeSimple      NodeType = "SIMPLE"
	NodeTypeSubWorkflow NodeType = "SUB_WORKFLOW"
)

type Node struct {
	Type        NodeType   `json:"type"`
	Simple      NodeSimple `json:"simple,omitempty"`
	SubWorkflow *Workflow  `json:"workflow,omitempty"`
}

func NewNodeSimple(simple NodeSimple) *Node {
	return &Node{
		Type:   NodeTypeSimple,
		Simple: simple,
	}
}

func NewSubWorkflow(intent Intent) *Node {
	return &Node{
		Type: NodeTypeSubWorkflow,
		SubWorkflow: &Workflow{
			// WorkflowID and InstanceID do not matter here.
			Intent: intent,
		},
	}
}

func (n *Node) Traverse(t WorkflowTraverser, w *Workflow) error {
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
		err := n.SubWorkflow.Traverse(t)
		if err != nil {
			return err
		}
		return nil
	default:
		panic(errors.New("unreachable"))
	}
}

func (n *Node) FindInputReactor(ctx context.Context, deps *Dependencies, w *Workflow) (*Workflow, InputReactor, error) {
	switch n.Type {
	case NodeTypeSimple:
		inputs, err := n.Simple.CanReactTo(ctx, deps, w)
		if err == nil {
			if len(inputs) == 0 {
				panic(fmt.Errorf("node %T react to no input without error", n.Simple))
			}
			return w, n.Simple, nil
		}
		return nil, nil, err
	case NodeTypeSubWorkflow:
		return n.SubWorkflow.FindInputReactor(ctx, deps)
	default:
		panic(errors.New("unreachable"))
	}
}

func (n *Node) Clone() *Node {
	cloned := &Node{
		Type: n.Type,
	}

	switch n.Type {
	case NodeTypeSimple:
		cloned.Simple = n.Simple
	case NodeTypeSubWorkflow:
		clonedWorkflow := n.SubWorkflow.Clone()
		cloned.SubWorkflow = clonedWorkflow
	default:
		panic(errors.New("unreachable"))
	}

	return cloned
}

func (n *Node) MarshalJSON() ([]byte, error) {
	var err error

	nodeJSON := nodeJSON{
		Type: n.Type,
	}

	switch n.Type {
	case NodeTypeSimple:
		var nodeSimpleBytes []byte
		nodeSimpleBytes, err = json.Marshal(n.Simple)
		if err != nil {
			return nil, err
		}

		nodeSimpleJSON := nodeSimpleJSON{
			Kind: n.Simple.Kind(),
			Data: nodeSimpleBytes,
		}
		nodeJSON.Simple = &nodeSimpleJSON
	case NodeTypeSubWorkflow:
		nodeJSON.SubWorkflow = n.SubWorkflow
	default:
		panic(errors.New("unreachable"))
	}

	return json.Marshal(nodeJSON)
}

func (n *Node) UnmarshalJSON(d []byte) (err error) {
	nodeJSON := nodeJSON{}
	// nodeJSON contains *Workflow
	// json.Unmarshal will call UnmarshalJSON of Workflow for us.
	err = json.Unmarshal(d, &nodeJSON)
	if err != nil {
		return
	}

	n.Type = nodeJSON.Type

	switch nodeJSON.Type {
	case NodeTypeSimple:
		var nodeSimple NodeSimple
		nodeSimple, err = InstantiateNode(nodeJSON.Simple.Kind)
		if err != nil {
			return
		}

		err = json.Unmarshal(nodeJSON.Simple.Data, nodeSimple)
		if err != nil {
			return
		}
		n.Simple = nodeSimple
	case NodeTypeSubWorkflow:
		n.SubWorkflow = nodeJSON.SubWorkflow
	default:
		panic(errors.New("unreachable"))
	}

	return nil
}

func (n *Node) ToOutput(ctx context.Context, deps *Dependencies, w *Workflow) (*NodeOutput, error) {
	output := &NodeOutput{
		Type: n.Type,
	}

	switch n.Type {
	case NodeTypeSimple:
		nodeSimpleData, err := n.Simple.OutputData(ctx, deps, w)
		if err != nil {
			return nil, err
		}
		output.Simple = &NodeSimpleOutput{
			Kind: n.Simple.Kind(),
			Data: nodeSimpleData,
		}
		return output, nil
	case NodeTypeSubWorkflow:
		workflowOutput, err := n.SubWorkflow.ToOutput(ctx, deps)
		if err != nil {
			return nil, err
		}
		output.SubWorkflow = workflowOutput
		return output, nil
	default:
		panic(errors.New("unreachable"))
	}
}

// NodeSimple can optionally implement CookieGetter.
type NodeSimple interface {
	InputReactor
	EffectGetter
	Kind() string
	OutputData(ctx context.Context, deps *Dependencies, workflow *Workflow) (interface{}, error)
}

type NodeFactory func() NodeSimple

var nodeRegistry = map[string]NodeFactory{}

func RegisterNode(node NodeSimple) {
	nodeType := reflect.TypeOf(node).Elem()

	nodeKind := node.Kind()
	factory := NodeFactory(func() NodeSimple {
		return reflect.New(nodeType).Interface().(NodeSimple)
	})

	if _, hasKind := nodeRegistry[nodeKind]; hasKind {
		panic(fmt.Errorf("workflow: duplicated node kind: %v", nodeKind))
	}
	nodeRegistry[nodeKind] = factory
}

func InstantiateNode(kind string) (NodeSimple, error) {
	factory, ok := nodeRegistry[kind]
	if !ok {
		return nil, fmt.Errorf("workflow: unknown node kind: %v", kind)
	}
	return factory(), nil
}
