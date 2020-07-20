package newinteraction

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/core/skyerr"
)

type Graph struct {
	// GraphID is the unique ID for a graph.
	// It is a constant value through out a graph.
	// It is used to keep track of which instances belong to a particular graph.
	// When one graph is committed, any other instances sharing the same GraphID become invalid.
	GraphID string `json:"graph_id"`

	// InstanceID is a unique ID for a particular instance of a graph.
	InstanceID string `json:"instance_id"`

	// Intent is the intent (i.e. flow type) of the graph
	Intent Intent `json:"intent"`

	// Nodes are nodes in a specific path from intent of the interaction graph.
	Nodes []Node `json:"nodes"`

	// TODO: any place to store error outside graph?
	Error *skyerr.APIError `json:"error"`
}

func newGraph(intent Intent) *Graph {
	return &Graph{
		GraphID:    newGraphID(),
		InstanceID: "",
		Intent:     intent,
		Nodes:      nil,
	}
}

func (g *Graph) AppendingNode(n Node) *Graph {
	nodes := make([]Node, len(g.Nodes)+1)
	copy(nodes, g.Nodes)
	nodes[len(nodes)-1] = n

	return &Graph{
		GraphID:    g.GraphID,
		InstanceID: "",
		Nodes:      nodes,
	}
}

func (g *Graph) MarshalJSON() ([]byte, error) {
	return nil, nil
}

func (g *Graph) UnmarshalJSON(d []byte) error {
	return nil
}

func (g *Graph) MustGetUserIdentity() *identity.Info {
	for i := len(g.Nodes) - 1; i >= 0; i-- {
		if n, ok := g.Nodes[i].(interface{ UserIdentity() *identity.Info }); ok {
			return n.UserIdentity()
		}
	}
	panic("interaction: expect user identity presents")
}

func (g *Graph) GetAuthenticator(stage AuthenticationStage) (*authenticator.Info, bool) {
	for i := len(g.Nodes) - 1; i >= 0; i-- {
		if n, ok := g.Nodes[i].(interface {
			UserAuthenticator() (AuthenticationStage, *authenticator.Info)
		}); ok {
			s, authenticator := n.UserAuthenticator()
			if s == stage {
				return authenticator, true
			}
		}
	}
	return nil, false
}

// Apply applies the effect the the graph nodes into the context.
func (g *Graph) Apply(ctx *Context) error {
	for _, node := range g.Nodes {
		if err := node.Apply(ctx, g); err != nil {
			return err
		}
	}
	return nil
}
