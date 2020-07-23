package newinteraction

import (
	"reflect"

	"github.com/authgear/authgear-server/pkg/core/errors"
)

//go:generate mockgen -source=node.go -destination=node_mock_test.go -package newinteraction_test

var ErrIncompatibleInput = errors.New("incompatible input type for this node")
var ErrSameNode = errors.New("the edge points to the same current node")

type Node interface {
	// Apply the effects of this node to context.
	// This may be ran multiple times, due to replaying the graph.
	// So no external visible side effect is allowed.
	Apply(perform func(eff Effect) error, graph *Graph) error
	DeriveEdges(ctx *Context, graph *Graph) ([]Edge, error)
}

type Edge interface {
	// Instantiate instantiates the node pointed by the edge.
	// It is ran once only for the pointed node, so side effects visible
	// outside the transaction (e.g. sending messages) is allowed.
	// It may return ErrSameNode if the edge loops back to self.
	// This is used to model side-effect only actions, such as sending
	// OTP message.
	Instantiate(ctx *Context, graph *Graph, input interface{}) (Node, error)
}

type NodeFactory func() Node

var nodeRegistry = map[string]NodeFactory{}

func RegisterNode(node Node) {
	nodeType := reflect.TypeOf(node).Elem()

	nodeKind := nodeType.Name()
	factory := NodeFactory(func() Node {
		return reflect.New(nodeType).Interface().(Node)
	})
	nodeRegistry[nodeKind] = factory
}

func NodeKind(node Node) string {
	nodeType := reflect.TypeOf(node).Elem()
	return nodeType.Name()
}

func InstantiateNode(kind string) Node {
	factory, ok := nodeRegistry[kind]
	if !ok {
		panic("interaction: unknown node kind: " + kind)
	}
	return factory()
}
