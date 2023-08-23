package authenticationflow

type NodeType string

const (
	NodeTypeSimple  NodeType = "SIMPLE"
	NodeTypeSubFlow NodeType = "SUB_FLOW"
)

type Node struct {
	Type    NodeType   `json:"type"`
	Simple  NodeSimple `json:"simple,omitempty"`
	SubFlow *Flow      `json:"flow,omitempty"`
}

type NodeSimple interface {
	Kinder
}

func NewNodeSimple(simple NodeSimple) *Node {
	return &Node{
		Type:   NodeTypeSimple,
		Simple: simple,
	}
}

func NewSubFlow(intent Intent) *Node {
	return &Node{
		Type: NodeTypeSubFlow,
		SubFlow: &Flow{
			// FlowID and InstanceID do not matter here.
			Intent: intent,
		},
	}
}
