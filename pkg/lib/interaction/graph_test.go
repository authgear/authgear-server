package interaction

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
)

func TestGraph(t *testing.T) {
	Convey("Graph.accept", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		any := gomock.Any()

		type input1 struct{}
		type input2 struct{}
		type input3 struct{}
		type input4 struct{}

		// A --> B --> D
		//  |    ^
		//  |    |
		//  \--> C --\
		//       ^---/
		nodeA := NewMockNode(ctrl)
		nodeB := NewMockNode(ctrl)
		nodeC := NewMockNode(ctrl)
		nodeD := NewMockNode(ctrl)
		edgeB := NewMockEdge(ctrl)
		edgeC := NewMockEdge(ctrl)
		edgeD := NewMockEdge(ctrl)
		edgeE := NewMockEdge(ctrl)

		ctx := &Context{}
		g := &Graph{Nodes: []Node{nodeA}}

		nodeA.EXPECT().Prepare(ctx, any).AnyTimes().Return(nil)
		nodeA.EXPECT().DeriveEdges(any).AnyTimes().Return(
			[]Edge{edgeB, edgeC}, nil,
		)
		nodeB.EXPECT().Prepare(ctx, any).AnyTimes().Return(nil)
		nodeB.EXPECT().DeriveEdges(any).AnyTimes().Return(
			[]Edge{edgeD}, nil,
		)
		nodeC.EXPECT().Prepare(ctx, any).AnyTimes().Return(nil)
		nodeC.EXPECT().DeriveEdges(any).AnyTimes().Return(
			[]Edge{edgeB, edgeE}, nil,
		)
		nodeD.EXPECT().Prepare(ctx, any).AnyTimes().Return(nil)
		nodeD.EXPECT().DeriveEdges(any).AnyTimes().Return(
			[]Edge{}, nil,
		)

		edgeB.EXPECT().Instantiate(ctx, any, any).AnyTimes().DoAndReturn(
			func(ctx *Context, g *Graph, input interface{}) (Node, error) {
				if _, ok := input.(input1); ok {
					return nodeB, nil
				}
				if _, ok := input.(input2); ok {
					return nodeB, nil
				}
				return nil, ErrIncompatibleInput
			})
		edgeC.EXPECT().Instantiate(ctx, any, any).AnyTimes().DoAndReturn(
			func(ctx *Context, g *Graph, input interface{}) (Node, error) {
				if _, ok := input.(input3); ok {
					return nodeC, nil
				}
				return nil, ErrIncompatibleInput
			})
		edgeD.EXPECT().Instantiate(ctx, any, any).AnyTimes().DoAndReturn(
			func(ctx *Context, g *Graph, input interface{}) (Node, error) {
				if _, ok := input.(input2); ok {
					return nodeD, nil
				}
				return nil, ErrIncompatibleInput
			})
		edgeE.EXPECT().Instantiate(ctx, any, any).AnyTimes().DoAndReturn(
			func(ctx *Context, g *Graph, input interface{}) (Node, error) {
				if _, ok := input.(input4); ok {
					return nil, ErrSameNode
				}
				return nil, ErrIncompatibleInput
			})

		Convey("should go to deepest node", func() {
			var inputRequired *ErrInputRequired

			nodeB.EXPECT().GetEffects()
			graph, edges, err := g.accept(ctx, input1{})
			So(errors.As(err, &inputRequired), ShouldBeTrue)
			So(graph.Nodes, ShouldResemble, []Node{nodeA, nodeB})
			So(edges, ShouldResemble, []Edge{edgeD})

			nodeB.EXPECT().GetEffects()
			nodeD.EXPECT().GetEffects()
			graph, edges, err = g.accept(ctx, input2{})
			So(err, ShouldBeNil)
			So(graph.Nodes, ShouldResemble, []Node{nodeA, nodeB, nodeD})
			So(edges, ShouldResemble, []Edge{})

			nodeC.EXPECT().GetEffects()
			graph, edges, err = g.accept(ctx, input3{})
			So(errors.As(err, &inputRequired), ShouldBeTrue)
			So(graph.Nodes, ShouldResemble, []Node{nodeA, nodeC})
			So(edges, ShouldResemble, []Edge{edgeB, edgeE})

			nodeB.EXPECT().GetEffects()
			nodeD.EXPECT().GetEffects()
			graph, edges, err = graph.accept(ctx, input2{})
			So(err, ShouldBeNil)
			So(graph.Nodes, ShouldResemble, []Node{nodeA, nodeC, nodeB, nodeD})
			So(edges, ShouldResemble, []Edge{})
		})

		Convey("should process looping edge", func() {
			var inputRequired *ErrInputRequired

			nodeC.EXPECT().GetEffects()
			graph, edges, err := g.accept(ctx, input3{})
			So(errors.As(err, &inputRequired), ShouldBeTrue)
			So(graph.Nodes, ShouldResemble, []Node{nodeA, nodeC})
			So(edges, ShouldResemble, []Edge{edgeB, edgeE})

			graph, edges, err = graph.accept(ctx, input4{})
			So(errors.As(err, &inputRequired), ShouldBeTrue)
			So(graph.Nodes, ShouldResemble, []Node{nodeA, nodeC})
			So(edges, ShouldResemble, []Edge{edgeB, edgeE})

			nodeB.EXPECT().GetEffects()
			nodeD.EXPECT().GetEffects()
			graph, edges, err = graph.accept(ctx, input2{})
			So(err, ShouldBeNil)
			So(graph.Nodes, ShouldResemble, []Node{nodeA, nodeC, nodeB, nodeD})
			So(edges, ShouldResemble, []Edge{})
		})
	})
}

type testGraphGetAMRnode struct {
	Stage         authn.AuthenticationStage
	Authenticator *authenticator.Info
}

func (n *testGraphGetAMRnode) Prepare(ctx *Context, graph *Graph) error {
	return nil
}

func (n *testGraphGetAMRnode) GetEffects() ([]Effect, error) {
	return nil, nil
}

func (n *testGraphGetAMRnode) DeriveEdges(graph *Graph) ([]Edge, error) {
	return nil, nil
}

func (n *testGraphGetAMRnode) UserAuthenticator(stage authn.AuthenticationStage) (*authenticator.Info, bool) {
	if n.Stage == stage && n.Authenticator != nil {
		return n.Authenticator, true
	}
	return nil, false
}

func TestGraphGetAMRACR(t *testing.T) {
	Convey("GraphGetAMRACR", t, func() {
		var graph *Graph
		var amr []string

		graph = &Graph{}
		So(graph.GetAMR(), ShouldBeEmpty)

		// password
		graph = &Graph{
			Nodes: []Node{
				&testGraphGetAMRnode{
					Stage: authn.AuthenticationStagePrimary,
					Authenticator: &authenticator.Info{
						Type: authn.AuthenticatorTypePassword,
					},
				},
			},
		}
		amr = graph.GetAMR()
		So(amr, ShouldResemble, []string{"pwd"})
		So(graph.GetACR(amr), ShouldEqual, "")

		// oob
		graph = &Graph{
			Nodes: []Node{
				&testGraphGetAMRnode{
					Stage: authn.AuthenticationStagePrimary,
					Authenticator: &authenticator.Info{
						Type:   authn.AuthenticatorTypeOOBSMS,
						Claims: map[string]interface{}{},
					},
				},
			},
		}
		amr = graph.GetAMR()
		So(amr, ShouldResemble, []string{"otp", "sms"})
		So(graph.GetACR(amr), ShouldEqual, "")

		// password + email oob
		graph = &Graph{
			Nodes: []Node{
				&testGraphGetAMRnode{
					Stage: authn.AuthenticationStagePrimary,
					Authenticator: &authenticator.Info{
						Type: authn.AuthenticatorTypePassword,
					},
				},
				&testGraphGetAMRnode{
					Stage: authn.AuthenticationStageSecondary,
					Authenticator: &authenticator.Info{
						Type:   authn.AuthenticatorTypeOOBEmail,
						Claims: map[string]interface{}{},
					},
				},
			},
		}
		amr = graph.GetAMR()
		So(amr, ShouldResemble, []string{"mfa", "otp", "pwd"})
		So(graph.GetACR(amr), ShouldEqual, authn.ACRMFA)

		// password + SMS oob
		graph = &Graph{
			Nodes: []Node{
				&testGraphGetAMRnode{
					Stage: authn.AuthenticationStagePrimary,
					Authenticator: &authenticator.Info{
						Type: authn.AuthenticatorTypePassword,
					},
				},
				&testGraphGetAMRnode{
					Stage: authn.AuthenticationStageSecondary,
					Authenticator: &authenticator.Info{
						Type:   authn.AuthenticatorTypeOOBSMS,
						Claims: map[string]interface{}{},
					},
				},
			},
		}
		amr = graph.GetAMR()
		So(amr, ShouldResemble, []string{"mfa", "otp", "pwd", "sms"})
		So(graph.GetACR(amr), ShouldEqual, authn.ACRMFA)

		// oauth + totp
		graph = &Graph{
			Nodes: []Node{
				&testGraphGetAMRnode{
					Stage: authn.AuthenticationStageSecondary,
					Authenticator: &authenticator.Info{
						Type: authn.AuthenticatorTypeTOTP,
					},
				},
			},
		}
		amr = graph.GetAMR()
		So(amr, ShouldResemble, []string{"mfa", "otp"})
		So(graph.GetACR(amr), ShouldEqual, authn.ACRMFA)
	})
}

type findNodeA struct{ Node }

func (*findNodeA) A() {}

type findNodeB struct{ Node }

func (*findNodeB) B() {}

type findNodeC struct{ Node }

func (*findNodeC) B() {}
func (*findNodeC) C() {}

func TestGraphFindLastNode(t *testing.T) {
	Convey("Graph.FindLastNode", t, func() {
		nodeA := &findNodeA{}
		nodeB := &findNodeB{}
		nodeC := &findNodeC{}
		graph := &Graph{
			Nodes: []Node{nodeA, nodeB, nodeC},
		}

		var a interface{ A() }
		So(graph.FindLastNode(&a), ShouldBeTrue)
		So(a, ShouldEqual, nodeA)

		var b interface{ B() }
		So(graph.FindLastNode(&b), ShouldBeTrue)
		So(b, ShouldEqual, nodeC)

		var c interface{ C() }
		So(graph.FindLastNode(&c), ShouldBeTrue)
		So(c, ShouldEqual, nodeC)

		var d interface{ D() }
		So(graph.FindLastNode(&d), ShouldBeFalse)
	})
}
