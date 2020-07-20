package newinteraction

import (
	"reflect"

	"github.com/authgear/authgear-server/pkg/core/errors"
)

var ErrIncompatibleInput = errors.New("incompatible input type for this node")

type Node interface {
	Apply(ctx *Context, graph *Graph) error
	DeriveEdges(ctx *Context, graph *Graph) ([]Edge, error)
}

type Edge interface {
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
