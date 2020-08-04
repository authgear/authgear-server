package newinteraction

import (
	"encoding/json"
	"errors"
	"sort"
	"time"

	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/core/authn"
	"github.com/authgear/authgear-server/pkg/core/utils"
)

const GraphLifetime = 5 * time.Minute

var ErrInputRequired = errors.New("new input is required")

type Graph struct {
	// GraphID is the unique ID for a graph.
	// It is a constant value through out a graph.
	// It is used to keep track of which instances belong to a particular graph.
	// When one graph is committed, any other instances sharing the same GraphID become invalid.
	GraphID string

	// InstanceID is a unique ID for a particular instance of a graph.
	InstanceID string

	// Intent is the intent (i.e. flow type) of the graph
	Intent Intent

	// Nodes are nodes in a specific path from intent of the interaction graph.
	Nodes []Node
}

func newGraph(intent Intent) *Graph {
	return &Graph{
		GraphID:    "",
		InstanceID: "",
		Intent:     intent,
		Nodes:      nil,
	}
}

func (g *Graph) CurrentNode() Node {
	return g.Nodes[len(g.Nodes)-1]
}

func (g *Graph) clone() *Graph {
	nodes := make([]Node, len(g.Nodes))
	copy(nodes, g.Nodes)

	return &Graph{
		GraphID:    g.GraphID,
		InstanceID: "",
		Intent:     g.Intent,
		Nodes:      nodes,
	}
}

func (g *Graph) appendingNode(n Node) *Graph {
	graph := g.clone()
	graph.Nodes = append(graph.Nodes, n)
	return graph
}

func (g *Graph) MarshalJSON() ([]byte, error) {
	var err error

	intent := ifaceJSON{Kind: IntentKind(g.Intent)}
	if intent.Data, err = json.Marshal(g.Intent); err != nil {
		return nil, err
	}

	nodes := make([]ifaceJSON, len(g.Nodes))
	for i, node := range g.Nodes {
		nodes[i].Kind = NodeKind(node)
		if nodes[i].Data, err = json.Marshal(node); err != nil {
			return nil, err
		}
	}

	graph := &graphJSON{
		GraphID:    g.GraphID,
		InstanceID: g.InstanceID,
		Intent:     intent,
		Nodes:      nodes,
	}
	return json.Marshal(graph)
}

func (g *Graph) UnmarshalJSON(d []byte) error {
	graph := &graphJSON{}
	if err := json.Unmarshal(d, graph); err != nil {
		return err
	}

	intent := InstantiateIntent(graph.Intent.Kind)
	if err := json.Unmarshal(graph.Intent.Data, intent); err != nil {
		return err
	}

	nodes := make([]Node, len(graph.Nodes))
	for i, node := range graph.Nodes {
		nodes[i] = InstantiateNode(node.Kind)
		if err := json.Unmarshal(node.Data, nodes[i]); err != nil {
			return err
		}
	}

	g.GraphID = graph.GraphID
	g.InstanceID = graph.InstanceID
	g.Intent = intent
	g.Nodes = nodes
	return nil
}

func (g *Graph) MustGetUserID() string {
	for i := len(g.Nodes) - 1; i >= 0; i-- {
		if n, ok := g.Nodes[i].(interface{ UserID() string }); ok {
			return n.UserID()
		}
	}
	panic("interaction: expect user ID presents")
}

func (g *Graph) GetNewUserID() (string, bool) {
	for i := len(g.Nodes) - 1; i >= 0; i-- {
		if n, ok := g.Nodes[i].(interface{ NewUserID() string }); ok {
			return n.NewUserID(), true
		}
	}
	return "", false
}

func (g *Graph) MustGetUserLastIdentity() *identity.Info {
	for i := len(g.Nodes) - 1; i >= 0; i-- {
		if n, ok := g.Nodes[i].(interface{ UserIdentity() *identity.Info }); ok {
			return n.UserIdentity()
		}
	}
	panic("interaction: expect user identity presents")
}

func (g *Graph) MustGetUpdateIdentityID() string {
	for i := len(g.Nodes) - 1; i >= 0; i-- {
		if n, ok := g.Nodes[i].(interface{ UpdateIdentityID() string }); ok {
			return n.UpdateIdentityID()
		}
	}
	panic("interaction: expect update identity ID presents")
}

func (g *Graph) GetUserNewIdentities() []*identity.Info {
	var identities []*identity.Info
	for _, node := range g.Nodes {
		if n, ok := node.(interface{ UserNewIdentity() *identity.Info }); ok {
			identities = append(identities, n.UserNewIdentity())
		}
	}
	return identities
}

func (g *Graph) GetUserAuthenticator(stage AuthenticationStage) (*authenticator.Info, bool) {
	for i := len(g.Nodes) - 1; i >= 0; i-- {
		if n, ok := g.Nodes[i].(interface {
			UserAuthenticator(stage AuthenticationStage) (*authenticator.Info, bool)
		}); ok {
			ai, ok := n.UserAuthenticator(stage)
			if ok {
				return ai, true
			}
		}
	}
	return nil, false
}

func (g *Graph) GetUserNewAuthenticators() []*authenticator.Info {
	var authenticators []*authenticator.Info
	for _, node := range g.Nodes {
		if n, ok := node.(interface{ UserNewAuthenticators() []*authenticator.Info }); ok {
			authenticators = append(authenticators, n.UserNewAuthenticators()...)
		}
	}
	return authenticators
}

func (g *Graph) GetAMR() []string {
	seen := make(map[string]struct{})
	amr := []string{}

	stages := []AuthenticationStage{
		AuthenticationStagePrimary,
		AuthenticationStageSecondary,
	}

	for _, stage := range stages {
		ai, ok := g.GetUserAuthenticator(stage)
		if ok {
			if stage == AuthenticationStageSecondary {
				amr = append(amr, authn.AMRMFA)
			}

			for _, value := range ai.AMR() {
				_, ok := seen[value]
				if !ok {
					seen[value] = struct{}{}
					amr = append(amr, value)
				}
			}
		}
	}

	sort.Strings(amr)

	return amr
}

func (g *Graph) GetACR(amrValues []string) string {
	if utils.StringSliceContains(amrValues, authn.AMRMFA) {
		return authn.ACRMFA
	}

	return ""
}

// Apply applies the effect the the graph nodes into the context.
func (g *Graph) Apply(ctx *Context) error {
	for _, node := range g.Nodes {
		if err := node.Apply(ctx.perform, g); err != nil {
			return err
		}
	}
	return nil
}

// Accept run the graph to the deepest node using the input
func (g *Graph) Accept(ctx *Context, input interface{}) (*Graph, []Edge, error) {
	graph := g
	hasTransitioned := false
	for {
		node := graph.CurrentNode()
		edges, err := node.DeriveEdges(ctx, graph)
		if err != nil {
			return nil, nil, err
		}

		if len(edges) == 0 {
			// No more edges, reached the end of the graph
			return graph, edges, nil
		}

		var nextNode Node
		for _, edge := range edges {
			nextNode, err = edge.Instantiate(ctx, graph, input)
			if errors.Is(err, ErrIncompatibleInput) {
				// Continue to check next edges
				continue
			} else if errors.Is(err, ErrSameNode) {
				// The next node is the same current node,
				// so no need to update the graph.
				// Continuing would keep traversing the same edge,
				// so stop and request new input.
				return graph.clone(), edges, ErrInputRequired
			} else if err != nil {
				return nil, nil, err
			}
			break
		}

		// No edges are followed, input is required
		if nextNode == nil {
			// Sanity check: we must have performed at least a transition, otherwise the graph
			// is stuck in same state
			if !hasTransitioned {
				panic("interaction: no transition is performed")
			}
			return graph, edges, ErrInputRequired
		}

		// Follow the edge to nextNode
		graph = graph.appendingNode(nextNode)
		err = nextNode.Apply(ctx.perform, graph)
		if err != nil {
			return nil, nil, err
		}

		hasTransitioned = true
	}
}

type ifaceJSON struct {
	Kind string          `json:"kind"`
	Data json.RawMessage `json:"data"`
}

type graphJSON struct {
	GraphID    string      `json:"graph_id"`
	InstanceID string      `json:"instance_id"`
	Intent     ifaceJSON   `json:"intent"`
	Nodes      []ifaceJSON `json:"nodes"`
}
