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

var _ ReactToResult = &Node{}

func (n *Node) reactToResult() {
	panic("unimplemented")
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
			// FlowID and StateToken do not matter here.
			Intent: intent,
		},
	}
}
