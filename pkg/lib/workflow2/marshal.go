package workflow2

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

// This file deals with the serialization / deserialization of data structures.

type Kinder interface {
	Kind() string
}

type intentFactory func() Intent

type nodeFactory func() NodeSimple

var intentRegistry = map[string]intentFactory{}

var nodeRegistry = map[string]nodeFactory{}

type workflowJSON struct {
	WorkflowID string     `json:"workflow_id,omitempty"`
	InstanceID string     `json:"instance_id,omitempty"`
	Intent     intentJSON `json:"intent"`
	Nodes      []Node     `json:"nodes,omitempty"`
}

type nodeJSON struct {
	Type        NodeType        `json:"type"`
	Simple      *nodeSimpleJSON `json:"simple,omitempty"`
	SubWorkflow *Workflow       `json:"workflow,omitempty"`
}

type nodeSimpleJSON struct {
	Kind string          `json:"kind"`
	Data json.RawMessage `json:"data"`
}

type intentJSON struct {
	Kind string          `json:"kind"`
	Data json.RawMessage `json:"data"`
}

func (w *Workflow) MarshalJSON() ([]byte, error) {
	var err error

	intentBytes, err := json.Marshal(w.Intent)
	if err != nil {
		return nil, err
	}

	intentJSON := intentJSON{
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

	intent, err := InstantiateIntent(workflowJSON.Intent.Kind)
	if err != nil {
		return
	}

	err = json.Unmarshal(workflowJSON.Intent.Data, intent)
	if err != nil {
		return
	}

	w.WorkflowID = workflowJSON.WorkflowID
	w.InstanceID = workflowJSON.InstanceID
	w.Intent = intent
	w.Nodes = workflowJSON.Nodes
	return nil
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

func RegisterIntent(intent Intent) {
	intentType := reflect.TypeOf(intent).Elem()

	intentKind := intent.Kind()
	factory := intentFactory(func() Intent {
		return reflect.New(intentType).Interface().(Intent)
	})

	if _, hasKind := intentRegistry[intentKind]; hasKind {
		panic(fmt.Errorf("workflow: duplicated intent kind: %v", intentKind))
	}

	intentRegistry[intentKind] = factory
}

func InstantiateIntent(kind string) (Intent, error) {
	factory, ok := intentRegistry[kind]
	if !ok {
		return nil, fmt.Errorf("workflow: unknown intent kind: %v", kind)
	}
	return factory(), nil
}

func RegisterNode(node NodeSimple) {
	nodeType := reflect.TypeOf(node).Elem()

	nodeKind := node.Kind()
	factory := nodeFactory(func() NodeSimple {
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
