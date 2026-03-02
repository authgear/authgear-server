package authenticationflow

import "context"

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

func (n *Node) reactToResult() {}

// DelayedOneTimeFunctionResult is the return value of a DelayedOneTimeFunction.
// If UpdatedSession is non-nil, processAcceptResult patches the live session and calls UpdateSession.
type DelayedOneTimeFunctionResult struct {
	UpdatedSession *Session
}

// DelayedOneTimeFunction
//   - executes outside the transaction.
//   - executes just before the flow state is saved to store
type DelayedOneTimeFunction func(ctx context.Context, deps *Dependencies) (DelayedOneTimeFunctionResult, error)

type NodeReactToResult struct {
	Node                   *Node
	DelayedOneTimeFunction DelayedOneTimeFunction // may be nil
	// UpdatedSession, if non-nil, is applied to the live session before delayed functions run.
	UpdatedSession *Session
}

var _ ReactToResult = &NodeReactToResult{}

func (n *NodeReactToResult) reactToResult() {}

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
