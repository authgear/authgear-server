package interaction

import (
	"encoding/json"
	"errors"
	"reflect"
	"sort"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

const GraphLifetime = duration.UserInteraction

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

func (g *Graph) FindLastNode(node interface{}) bool {
	idx := g.FindLastNodeAndPosition(node)
	return idx >= 0
}

func (g *Graph) FindLastNodeAndPosition(node interface{}) int {
	val := reflect.ValueOf(node)
	typ := val.Type()
	if typ.Kind() != reflect.Ptr || val.IsNil() {
		panic("interaction: node must be a non-nil pointer")
	}
	if e := typ.Elem(); e.Kind() != reflect.Interface {
		panic("interaction: *node must be interface")
	}
	targetType := typ.Elem()
	for i := len(g.Nodes) - 1; i >= 0; i-- {
		n := g.Nodes[i]
		if reflect.TypeOf(n).AssignableTo(targetType) {
			val.Elem().Set(reflect.ValueOf(n))
			return i
		}
	}
	return -1
}

// FindLastNodeFromList find the last node from a list of node interface
func (g *Graph) FindLastNodeFromList(nodes []interface{}) (node interface{}) {
	maxNodePosition := -1
	nodeIdx := -1
	for i, n := range nodes {
		nodePosition := g.FindLastNodeAndPosition(n)
		if nodePosition > maxNodePosition {
			maxNodePosition = nodePosition
			nodeIdx = i
		}
	}
	if nodeIdx < 0 {
		return nil
	}
	return nodes[nodeIdx]
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
	iden, ok := g.GetUserLastIdentity()
	if !ok {
		panic("interaction: expect user identity presents")
	}
	return iden
}

func (g *Graph) GetUserLastIdentity() (*identity.Info, bool) {
	for i := len(g.Nodes) - 1; i >= 0; i-- {
		if n, ok := g.Nodes[i].(interface{ UserIdentity() *identity.Info }); ok {
			iden := n.UserIdentity()
			if iden != nil {
				return iden, true
			}
		}
	}
	return nil, false
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

func (g *Graph) GetUserAuthenticator(stage authn.AuthenticationStage) (*authenticator.Info, bool) {
	for i := len(g.Nodes) - 1; i >= 0; i-- {
		if n, ok := g.Nodes[i].(interface {
			UserAuthenticator(stage authn.AuthenticationStage) (*authenticator.Info, bool)
		}); ok {
			ai, ok := n.UserAuthenticator(stage)
			if ok {
				return ai, true
			}
		}
	}
	return nil, false
}

func (g *Graph) GetUsedAuthenticationLockoutMethods() []config.AuthenticationLockoutMethod {
	result := []config.AuthenticationLockoutMethod{}
	for i := len(g.Nodes) - 1; i >= 0; i-- {
		if n, ok := g.Nodes[i].(interface {
			UsedAuthenticationLockoutMethod() (config.AuthenticationLockoutMethod, bool)
		}); ok {
			m, ok := n.UsedAuthenticationLockoutMethod()
			if ok {
				result = append(result, m)
			}
		}
	}
	return result
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

func (g *Graph) GetRequireUpdateAuthenticator(stage authn.AuthenticationStage) (*authenticator.Info, bool) {
	for _, node := range g.Nodes {
		if n, ok := node.(interface {
			GetRequireUpdateAuthenticator(stage authn.AuthenticationStage) (*authenticator.Info, bool)
		}); ok {
			info, ok := n.GetRequireUpdateAuthenticator(stage)
			if ok {
				return info, ok
			}
		}
	}
	return nil, false
}

func (g *Graph) GetAMR() []string {
	seen := make(map[string]struct{})
	amr := []string{}

	stages := []authn.AuthenticationStage{
		authn.AuthenticationStagePrimary,
		authn.AuthenticationStageSecondary,
	}

	// Some AMR values are from identity, for example, biometric identity.
	if iden, ok := g.GetUserLastIdentity(); ok {
		for _, value := range iden.AMR() {
			_, ok := seen[value]
			if !ok {
				seen[value] = struct{}{}
				amr = append(amr, value)
			}
		}
	}

	for _, stage := range stages {
		ai, ok := g.GetUserAuthenticator(stage)
		if ok {
			if stage == authn.AuthenticationStageSecondary {
				amr = append(amr, model.AMRMFA)
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

func (g *Graph) FillDetails(err error) error {
	return errorutil.WithDetails(err, errorutil.Details{
		"IntentKind": apierrors.APIErrorDetail.Value(IntentKind(g.Intent)),
	})
}

// Apply applies the effect the the graph nodes into the context.
func (g *Graph) Apply(ctx *Context) error {
	for i, node := range g.Nodes {
		// Prepare the node with sliced graph.
		slicedGraph := *g
		slicedGraph.Nodes = slicedGraph.Nodes[:i+1]
		if err := node.Prepare(ctx, &slicedGraph); err != nil {
			return g.FillDetails(err)
		}

		effs, err := node.GetEffects()
		if err != nil {
			return g.FillDetails(err)
		}
		for _, eff := range effs {
			// Apply the effect with unsliced graph.
			err = eff.apply(ctx, g, i)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Accept run the graph to the deepest node using the input
func (g *Graph) accept(ctx *Context, input interface{}) (*Graph, []Edge, error) {
	graph := g
	for {
		node := graph.CurrentNode()
		edges, err := node.DeriveEdges(graph)
		if err != nil {
			return nil, nil, graph.FillDetails(err)
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
				return graph.clone(), edges, &ErrInputRequired{Inner: err}
			} else if err != nil {
				return nil, nil, graph.FillDetails(err)
			}
			break
		}

		// No edges are followed, input is required
		if nextNode == nil {
			return graph, edges, &ErrInputRequired{}
		}

		// Follow the edge to nextNode
		graph = graph.appendingNode(nextNode)
		err = nextNode.Prepare(ctx, graph)
		if err != nil {
			return nil, nil, err
		}
		effs, err := nextNode.GetEffects()
		if err != nil {
			return nil, nil, err
		}
		for _, eff := range effs {
			err = eff.apply(ctx, graph, len(graph.Nodes)-1)
			if err != nil {
				return nil, nil, err
			}
		}
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
