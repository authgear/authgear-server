package interaction_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

type testNode struct {
	_ int
}

func (t testNode) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (t testNode) Apply(perform func(eff interaction.Effect) error, graph *interaction.Graph) error {
	return nil
}

func (t testNode) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return nil, nil
}

var _ interaction.Node = &testNode{}

func TestNodeRegistry(t *testing.T) {
	Convey("Node registry", t, func() {
		n0 := &testNode{}
		interaction.RegisterNode(n0)

		n1 := &testNode{}
		nodeKind := interaction.NodeKind(n1)
		So(nodeKind, ShouldEqual, "testNode")

		n2 := interaction.InstantiateNode(nodeKind)
		So(n2, ShouldHaveSameTypeAs, n0)
		So(n2, ShouldNotPointTo, n0)
		So(n2, ShouldNotPointTo, n1)
	})
}
