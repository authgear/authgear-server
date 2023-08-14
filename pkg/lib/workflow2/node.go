package workflow2

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

type NodeSimple interface {
	Kinder
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
