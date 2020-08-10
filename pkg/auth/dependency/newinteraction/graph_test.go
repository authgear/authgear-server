package newinteraction_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func TestGraph(t *testing.T) {
	Convey("Graph.Accept", t, func() {
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

		ctx := &newinteraction.Context{}
		g := &newinteraction.Graph{Nodes: []newinteraction.Node{nodeA}}

		nodeA.EXPECT().Prepare(ctx, any).AnyTimes().Return(nil)
		nodeA.EXPECT().DeriveEdges(any).AnyTimes().Return(
			[]newinteraction.Edge{edgeB, edgeC}, nil,
		)
		nodeB.EXPECT().Prepare(ctx, any).AnyTimes().Return(nil)
		nodeB.EXPECT().DeriveEdges(any).AnyTimes().Return(
			[]newinteraction.Edge{edgeD}, nil,
		)
		nodeC.EXPECT().Prepare(ctx, any).AnyTimes().Return(nil)
		nodeC.EXPECT().DeriveEdges(any).AnyTimes().Return(
			[]newinteraction.Edge{edgeB, edgeE}, nil,
		)
		nodeD.EXPECT().Prepare(ctx, any).AnyTimes().Return(nil)
		nodeD.EXPECT().DeriveEdges(any).AnyTimes().Return(
			[]newinteraction.Edge{}, nil,
		)

		edgeB.EXPECT().Instantiate(ctx, any, any).AnyTimes().DoAndReturn(
			func(ctx *newinteraction.Context, g *newinteraction.Graph, input interface{}) (newinteraction.Node, error) {
				if _, ok := input.(input1); ok {
					return nodeB, nil
				}
				if _, ok := input.(input2); ok {
					return nodeB, nil
				}
				return nil, newinteraction.ErrIncompatibleInput
			})
		edgeC.EXPECT().Instantiate(ctx, any, any).AnyTimes().DoAndReturn(
			func(ctx *newinteraction.Context, g *newinteraction.Graph, input interface{}) (newinteraction.Node, error) {
				if _, ok := input.(input3); ok {
					return nodeC, nil
				}
				return nil, newinteraction.ErrIncompatibleInput
			})
		edgeD.EXPECT().Instantiate(ctx, any, any).AnyTimes().DoAndReturn(
			func(ctx *newinteraction.Context, g *newinteraction.Graph, input interface{}) (newinteraction.Node, error) {
				if _, ok := input.(input2); ok {
					return nodeD, nil
				}
				return nil, newinteraction.ErrIncompatibleInput
			})
		edgeE.EXPECT().Instantiate(ctx, any, any).AnyTimes().DoAndReturn(
			func(ctx *newinteraction.Context, g *newinteraction.Graph, input interface{}) (newinteraction.Node, error) {
				if _, ok := input.(input4); ok {
					return nil, newinteraction.ErrSameNode
				}
				return nil, newinteraction.ErrIncompatibleInput
			})

		Convey("should go to deepest node", func() {
			nodeB.EXPECT().Apply(any, any)
			graph, edges, err := g.Accept(ctx, input1{})
			So(err, ShouldBeError, newinteraction.ErrInputRequired)
			So(graph.Nodes, ShouldResemble, []newinteraction.Node{nodeA, nodeB})
			So(edges, ShouldResemble, []newinteraction.Edge{edgeD})

			nodeB.EXPECT().Apply(any, any)
			nodeD.EXPECT().Apply(any, any)
			graph, edges, err = g.Accept(ctx, input2{})
			So(err, ShouldBeNil)
			So(graph.Nodes, ShouldResemble, []newinteraction.Node{nodeA, nodeB, nodeD})
			So(edges, ShouldResemble, []newinteraction.Edge{})

			nodeC.EXPECT().Apply(any, any)
			graph, edges, err = g.Accept(ctx, input3{})
			So(err, ShouldBeError, newinteraction.ErrInputRequired)
			So(graph.Nodes, ShouldResemble, []newinteraction.Node{nodeA, nodeC})
			So(edges, ShouldResemble, []newinteraction.Edge{edgeB, edgeE})

			nodeB.EXPECT().Apply(any, any)
			nodeD.EXPECT().Apply(any, any)
			graph, edges, err = graph.Accept(ctx, input2{})
			So(err, ShouldBeNil)
			So(graph.Nodes, ShouldResemble, []newinteraction.Node{nodeA, nodeC, nodeB, nodeD})
			So(edges, ShouldResemble, []newinteraction.Edge{})
		})

		Convey("should process looping edge", func() {
			nodeC.EXPECT().Apply(any, any)
			graph, edges, err := g.Accept(ctx, input3{})
			So(err, ShouldBeError, newinteraction.ErrInputRequired)
			So(graph.Nodes, ShouldResemble, []newinteraction.Node{nodeA, nodeC})
			So(edges, ShouldResemble, []newinteraction.Edge{edgeB, edgeE})

			graph, edges, err = graph.Accept(ctx, input4{})
			So(err, ShouldBeError, newinteraction.ErrInputRequired)
			So(graph.Nodes, ShouldResemble, []newinteraction.Node{nodeA, nodeC})
			So(edges, ShouldResemble, []newinteraction.Edge{edgeB, edgeE})

			nodeB.EXPECT().Apply(any, any)
			nodeD.EXPECT().Apply(any, any)
			graph, edges, err = graph.Accept(ctx, input2{})
			So(err, ShouldBeNil)
			So(graph.Nodes, ShouldResemble, []newinteraction.Node{nodeA, nodeC, nodeB, nodeD})
			So(edges, ShouldResemble, []newinteraction.Edge{})
		})
	})
}

type testGraphGetAMRnode struct {
	Stage         newinteraction.AuthenticationStage
	Authenticator *authenticator.Info
}

func (n *testGraphGetAMRnode) Prepare(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	return nil
}

func (n *testGraphGetAMRnode) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *testGraphGetAMRnode) DeriveEdges(graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return nil, nil
}

func (n *testGraphGetAMRnode) UserAuthenticator(stage newinteraction.AuthenticationStage) (*authenticator.Info, bool) {
	if n.Stage == stage && n.Authenticator != nil {
		return n.Authenticator, true
	}
	return nil, false
}

func TestGraphGetAMRACR(t *testing.T) {
	Convey("GraphGetAMRACR", t, func() {
		var graph *newinteraction.Graph
		var amr []string

		graph = &newinteraction.Graph{}
		So(graph.GetAMR(), ShouldBeEmpty)

		// password
		graph = &newinteraction.Graph{
			Nodes: []newinteraction.Node{
				&testGraphGetAMRnode{
					Stage: newinteraction.AuthenticationStagePrimary,
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
		graph = &newinteraction.Graph{
			Nodes: []newinteraction.Node{
				&testGraphGetAMRnode{
					Stage: newinteraction.AuthenticationStagePrimary,
					Authenticator: &authenticator.Info{
						Type: authn.AuthenticatorTypeOOB,
						Props: map[string]interface{}{
							authenticator.AuthenticatorPropOOBOTPChannelType: string(authn.AuthenticatorOOBChannelSMS),
						},
					},
				},
			},
		}
		amr = graph.GetAMR()
		So(amr, ShouldResemble, []string{"otp", "sms"})
		So(graph.GetACR(amr), ShouldEqual, "")

		// password + email oob
		graph = &newinteraction.Graph{
			Nodes: []newinteraction.Node{
				&testGraphGetAMRnode{
					Stage: newinteraction.AuthenticationStagePrimary,
					Authenticator: &authenticator.Info{
						Type: authn.AuthenticatorTypePassword,
					},
				},
				&testGraphGetAMRnode{
					Stage: newinteraction.AuthenticationStageSecondary,
					Authenticator: &authenticator.Info{
						Type: authn.AuthenticatorTypeOOB,
						Props: map[string]interface{}{
							authenticator.AuthenticatorPropOOBOTPChannelType: string(authn.AuthenticatorOOBChannelEmail),
						},
					},
				},
			},
		}
		amr = graph.GetAMR()
		So(amr, ShouldResemble, []string{"mfa", "otp", "pwd"})
		So(graph.GetACR(amr), ShouldEqual, authn.ACRMFA)

		// password + SMS oob
		graph = &newinteraction.Graph{
			Nodes: []newinteraction.Node{
				&testGraphGetAMRnode{
					Stage: newinteraction.AuthenticationStagePrimary,
					Authenticator: &authenticator.Info{
						Type: authn.AuthenticatorTypePassword,
					},
				},
				&testGraphGetAMRnode{
					Stage: newinteraction.AuthenticationStageSecondary,
					Authenticator: &authenticator.Info{
						Type: authn.AuthenticatorTypeOOB,
						Props: map[string]interface{}{
							authenticator.AuthenticatorPropOOBOTPChannelType: string(authn.AuthenticatorOOBChannelSMS),
						},
					},
				},
			},
		}
		amr = graph.GetAMR()
		So(amr, ShouldResemble, []string{"mfa", "otp", "pwd", "sms"})
		So(graph.GetACR(amr), ShouldEqual, authn.ACRMFA)

		// oauth + totp
		graph = &newinteraction.Graph{
			Nodes: []newinteraction.Node{
				&testGraphGetAMRnode{
					Stage: newinteraction.AuthenticationStageSecondary,
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

type findNodeA struct{ newinteraction.Node }

func (*findNodeA) A() {}

type findNodeB struct{ newinteraction.Node }

func (*findNodeB) B() {}

type findNodeC struct{ newinteraction.Node }

func (*findNodeC) B() {}
func (*findNodeC) C() {}

func TestGraphFindLastNode(t *testing.T) {
	Convey("Graph.FindLastNode", t, func() {
		nodeA := &findNodeA{}
		nodeB := &findNodeB{}
		nodeC := &findNodeC{}
		graph := &newinteraction.Graph{
			Nodes: []newinteraction.Node{nodeA, nodeB, nodeC},
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
