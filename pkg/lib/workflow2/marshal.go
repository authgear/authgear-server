package workflow2

import (
	"encoding/json"
	"errors"
)

type workflowJSON struct {
	WorkflowID string     `json:"workflow_id,omitempty"`
	InstanceID string     `json:"instance_id,omitempty"`
	Intent     IntentJSON `json:"intent"`
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

func (w *Workflow) MarshalJSON() ([]byte, error) {
	var err error

	intentBytes, err := json.Marshal(w.Intent)
	if err != nil {
		return nil, err
	}

	intentJSON := IntentJSON{
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

	intent, err := InstantiateIntentFromPrivateRegistry(workflowJSON.Intent)
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
