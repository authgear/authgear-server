package workflow

import (
	"encoding/json"
	"errors"
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

func (n *Node) GetEffects(ctx *Context) ([]Effect, error) {
	switch n.Type {
	case NodeTypeSimple:
		return n.Simple.GetEffects(ctx)
	case NodeTypeSubWorkflow:
		var allEffects []Effect
		for _, node := range n.SubWorkflow.Nodes {
			effs, err := node.GetEffects(ctx)
			if err != nil {
				return nil, err
			}
			allEffects = append(allEffects, effs...)
		}

		return allEffects, nil
	default:
		panic(errors.New("unreachable"))
	}
}

func (n *Node) DeriveEdges(ctx *Context, workflow *Workflow) (*Workflow, []Edge, error) {
	switch n.Type {
	case NodeTypeSimple:
		edges, err := n.Simple.DeriveEdges(ctx)
		if err != nil {
			return nil, nil, err
		}
		return workflow, edges, nil
	case NodeTypeSubWorkflow:
		return n.SubWorkflow.DeriveEdges(ctx)
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
			Kind: NodeKind(n.Simple),
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
		nodeSimple := InstantiateNode(nodeJSON.Simple.Kind)
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

func (n *Node) ToOutput(ctx *Context) (*NodeOutput, error) {
	output := &NodeOutput{
		Type: n.Type,
	}

	switch n.Type {
	case NodeTypeSimple:
		nodeSimpleData, err := n.Simple.OutputData(ctx)
		if err != nil {
			return nil, err
		}
		output.Simple = &NodeSimpleOutput{
			Kind: NodeKind(n.Simple),
			Data: nodeSimpleData,
		}
		return output, nil
	case NodeTypeSubWorkflow:
		workflowOutput, err := n.SubWorkflow.ToOutput(ctx)
		if err != nil {
			return nil, err
		}
		output.SubWorkflow = workflowOutput
		return output, nil
	default:
		panic(errors.New("unreachable"))
	}
}

type NodeSimple interface {
	GetEffects(ctx *Context) (effs []Effect, err error)
	DeriveEdges(ctx *Context) ([]Edge, error)
	OutputData(ctx *Context) (interface{}, error)
}

type Edge interface {
	Instantiate(ctx *Context, workflow *Workflow, input interface{}) (*Node, error)
}

type NodeFactory func() NodeSimple

var nodeRegistry = map[string]NodeFactory{}

func RegisterNode(node NodeSimple) {
	nodeType := reflect.TypeOf(node).Elem()

	nodeKind := nodeType.Name()
	factory := NodeFactory(func() NodeSimple {
		return reflect.New(nodeType).Interface().(NodeSimple)
	})

	if _, hasKind := nodeRegistry[nodeKind]; hasKind {
		panic("interaction: duplicated node kind: " + nodeKind)
	}
	nodeRegistry[nodeKind] = factory
}

func NodeKind(node NodeSimple) string {
	nodeType := reflect.TypeOf(node).Elem()
	return nodeType.Name()
}

func InstantiateNode(kind string) NodeSimple {
	factory, ok := nodeRegistry[kind]
	if !ok {
		panic("interaction: unknown node kind: " + kind)
	}
	return factory()
}
