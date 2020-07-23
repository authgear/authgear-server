package newinteraction_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
)

type testNode struct {
	x int
}

func (t testNode) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (t testNode) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return nil, nil
}

var _ newinteraction.Node = &testNode{}

func TestNodeRegistry(t *testing.T) {
	Convey("Node registry", t, func() {
		n0 := &testNode{}
		newinteraction.RegisterNode(n0)

		n1 := &testNode{}
		nodeKind := newinteraction.NodeKind(n1)
		So(nodeKind, ShouldEqual, "testNode")

		n2 := newinteraction.InstantiateNode(nodeKind)
		So(n2, ShouldHaveSameTypeAs, n0)
		So(n2, ShouldNotPointTo, n0)
		So(n2, ShouldNotPointTo, n1)
	})
}
