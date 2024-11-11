package interaction

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
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

		nodeA.EXPECT().Prepare(gomock.Any(), ctx, any).AnyTimes().Return(nil)
		nodeA.EXPECT().DeriveEdges(gomock.Any(), any).AnyTimes().Return(
			[]Edge{edgeB, edgeC}, nil,
		)
		nodeB.EXPECT().Prepare(gomock.Any(), ctx, any).AnyTimes().Return(nil)
		nodeB.EXPECT().DeriveEdges(gomock.Any(), any).AnyTimes().Return(
			[]Edge{edgeD}, nil,
		)
		nodeC.EXPECT().Prepare(gomock.Any(), ctx, any).AnyTimes().Return(nil)
		nodeC.EXPECT().DeriveEdges(gomock.Any(), any).AnyTimes().Return(
			[]Edge{edgeB, edgeE}, nil,
		)
		nodeD.EXPECT().Prepare(gomock.Any(), ctx, any).AnyTimes().Return(nil)
		nodeD.EXPECT().DeriveEdges(gomock.Any(), any).AnyTimes().Return(
			[]Edge{}, nil,
		)

		edgeB.EXPECT().Instantiate(gomock.Any(), ctx, any, any).AnyTimes().DoAndReturn(
			func(goCtx context.Context, ctx *Context, g *Graph, input interface{}) (Node, error) {
				if _, ok := input.(input1); ok {
					return nodeB, nil
				}
				if _, ok := input.(input2); ok {
					return nodeB, nil
				}
				return nil, ErrIncompatibleInput
			})
		edgeC.EXPECT().Instantiate(gomock.Any(), ctx, any, any).AnyTimes().DoAndReturn(
			func(goCtx context.Context, ctx *Context, g *Graph, input interface{}) (Node, error) {
				if _, ok := input.(input3); ok {
					return nodeC, nil
				}
				return nil, ErrIncompatibleInput
			})
		edgeD.EXPECT().Instantiate(gomock.Any(), ctx, any, any).AnyTimes().DoAndReturn(
			func(goCtx context.Context, ctx *Context, g *Graph, input interface{}) (Node, error) {
				if _, ok := input.(input2); ok {
					return nodeD, nil
				}
				return nil, ErrIncompatibleInput
			})
		edgeE.EXPECT().Instantiate(gomock.Any(), ctx, any, any).AnyTimes().DoAndReturn(
			func(goCtx context.Context, ctx *Context, g *Graph, input interface{}) (Node, error) {
				if _, ok := input.(input4); ok {
					return nil, ErrSameNode
				}
				return nil, ErrIncompatibleInput
			})

		Convey("should go to deepest node", func() {
			var inputRequired *ErrInputRequired

			goCtx := context.Background()
			nodeB.EXPECT().GetEffects(gomock.Any())
			graph, edges, err := g.accept(goCtx, ctx, input1{})
			So(errors.As(err, &inputRequired), ShouldBeTrue)
			So(graph.Nodes, ShouldResemble, []Node{nodeA, nodeB})
			So(edges, ShouldResemble, []Edge{edgeD})

			nodeB.EXPECT().GetEffects(gomock.Any())
			nodeD.EXPECT().GetEffects(gomock.Any())
			graph, edges, err = g.accept(goCtx, ctx, input2{})
			So(err, ShouldBeNil)
			So(graph.Nodes, ShouldResemble, []Node{nodeA, nodeB, nodeD})
			So(edges, ShouldResemble, []Edge{})

			nodeC.EXPECT().GetEffects(gomock.Any())
			graph, edges, err = g.accept(goCtx, ctx, input3{})
			So(errors.As(err, &inputRequired), ShouldBeTrue)
			So(graph.Nodes, ShouldResemble, []Node{nodeA, nodeC})
			So(edges, ShouldResemble, []Edge{edgeB, edgeE})

			nodeB.EXPECT().GetEffects(gomock.Any())
			nodeD.EXPECT().GetEffects(gomock.Any())
			graph, edges, err = graph.accept(goCtx, ctx, input2{})
			So(err, ShouldBeNil)
			So(graph.Nodes, ShouldResemble, []Node{nodeA, nodeC, nodeB, nodeD})
			So(edges, ShouldResemble, []Edge{})
		})

		Convey("should process looping edge", func() {
			var inputRequired *ErrInputRequired

			goCtx := context.Background()
			nodeC.EXPECT().GetEffects(gomock.Any())
			graph, edges, err := g.accept(goCtx, ctx, input3{})
			So(errors.As(err, &inputRequired), ShouldBeTrue)
			So(graph.Nodes, ShouldResemble, []Node{nodeA, nodeC})
			So(edges, ShouldResemble, []Edge{edgeB, edgeE})

			graph, edges, err = graph.accept(goCtx, ctx, input4{})
			So(errors.As(err, &inputRequired), ShouldBeTrue)
			So(graph.Nodes, ShouldResemble, []Node{nodeA, nodeC})
			So(edges, ShouldResemble, []Edge{edgeB, edgeE})

			nodeB.EXPECT().GetEffects(gomock.Any())
			nodeD.EXPECT().GetEffects(gomock.Any())
			graph, edges, err = graph.accept(goCtx, ctx, input2{})
			So(err, ShouldBeNil)
			So(graph.Nodes, ShouldResemble, []Node{nodeA, nodeC, nodeB, nodeD})
			So(edges, ShouldResemble, []Edge{})
		})
	})
}

type testGraphGetAMRnode struct {
	Stage         authn.AuthenticationStage
	Identity      *identity.Info
	Authenticator *authenticator.Info
}

func (n *testGraphGetAMRnode) Prepare(goCtx context.Context, ctx *Context, graph *Graph) error {
	return nil
}

func (n *testGraphGetAMRnode) GetEffects(goCtx context.Context) ([]Effect, error) {
	return nil, nil
}

func (n *testGraphGetAMRnode) DeriveEdges(goCtx context.Context, graph *Graph) ([]Edge, error) {
	return nil, nil
}

func (n *testGraphGetAMRnode) UserAuthenticator(stage authn.AuthenticationStage) (*authenticator.Info, bool) {
	if n.Stage == stage && n.Authenticator != nil {
		return n.Authenticator, true
	}
	return nil, false
}

func (n *testGraphGetAMRnode) UserIdentity() *identity.Info {
	return n.Identity
}

func TestGraphGetAMR(t *testing.T) {
	Convey("GraphGetAMR", t, func() {
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
						Type: model.AuthenticatorTypePassword,
					},
				},
			},
		}
		amr = graph.GetAMR()
		So(amr, ShouldResemble, []string{"pwd"})

		// oob
		graph = &Graph{
			Nodes: []Node{
				&testGraphGetAMRnode{
					Stage: authn.AuthenticationStagePrimary,
					Authenticator: &authenticator.Info{
						Type: model.AuthenticatorTypeOOBSMS,
					},
				},
			},
		}
		amr = graph.GetAMR()
		So(amr, ShouldResemble, []string{"otp", "sms"})

		// password + email oob
		graph = &Graph{
			Nodes: []Node{
				&testGraphGetAMRnode{
					Stage: authn.AuthenticationStagePrimary,
					Authenticator: &authenticator.Info{
						Type: model.AuthenticatorTypePassword,
					},
				},
				&testGraphGetAMRnode{
					Stage: authn.AuthenticationStageSecondary,
					Authenticator: &authenticator.Info{
						Type: model.AuthenticatorTypeOOBEmail,
					},
				},
			},
		}
		amr = graph.GetAMR()
		So(amr, ShouldResemble, []string{"mfa", "otp", "pwd"})

		// password + SMS oob
		graph = &Graph{
			Nodes: []Node{
				&testGraphGetAMRnode{
					Stage: authn.AuthenticationStagePrimary,
					Authenticator: &authenticator.Info{
						Type: model.AuthenticatorTypePassword,
					},
				},
				&testGraphGetAMRnode{
					Stage: authn.AuthenticationStageSecondary,
					Authenticator: &authenticator.Info{
						Type: model.AuthenticatorTypeOOBSMS,
					},
				},
			},
		}
		amr = graph.GetAMR()
		So(amr, ShouldResemble, []string{"mfa", "otp", "pwd", "sms"})

		// oauth + totp
		graph = &Graph{
			Nodes: []Node{
				&testGraphGetAMRnode{
					Stage: authn.AuthenticationStageSecondary,
					Authenticator: &authenticator.Info{
						Type: model.AuthenticatorTypeTOTP,
					},
				},
			},
		}
		amr = graph.GetAMR()
		So(amr, ShouldResemble, []string{"mfa", "otp"})

		// biometric
		graph = &Graph{
			Nodes: []Node{
				&testGraphGetAMRnode{
					Identity: &identity.Info{
						Type: model.IdentityTypeBiometric,
					},
				},
			},
		}
		amr = graph.GetAMR()
		So(amr, ShouldResemble, []string{"x_biometric"})
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
		So(graph.FindLastNodeAndPosition(&a), ShouldEqual, 0)
		So(a, ShouldEqual, nodeA)

		var b interface{ B() }
		So(graph.FindLastNode(&b), ShouldBeTrue)
		So(graph.FindLastNodeAndPosition(&b), ShouldEqual, 2)
		So(b, ShouldEqual, nodeC)

		var c interface{ C() }
		So(graph.FindLastNode(&c), ShouldBeTrue)
		So(graph.FindLastNodeAndPosition(&c), ShouldEqual, 2)
		So(c, ShouldEqual, nodeC)

		var d interface{ D() }
		So(graph.FindLastNode(&d), ShouldBeFalse)
		So(graph.FindLastNodeAndPosition(&d), ShouldEqual, -1)

		var e interface{ E() }
		So(graph.FindLastNodeFromList([]interface{}{&a, &b, &d}), ShouldEqual, &b)
		So(graph.FindLastNodeFromList([]interface{}{&a, &b, &c}), ShouldEqual, &b)
		So(graph.FindLastNodeFromList([]interface{}{&a, &b, &c, &d}), ShouldEqual, &b)
		So(graph.FindLastNodeFromList([]interface{}{&d, &a}), ShouldEqual, &a)
		So(graph.FindLastNodeFromList([]interface{}{&d, &e}), ShouldEqual, nil)

	})
}
