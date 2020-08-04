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

		nodeA.EXPECT().DeriveEdges(ctx, any).AnyTimes().Return(
			[]newinteraction.Edge{edgeB, edgeC}, nil,
		)
		nodeB.EXPECT().DeriveEdges(ctx, any).AnyTimes().Return(
			[]newinteraction.Edge{edgeD}, nil,
		)
		nodeC.EXPECT().DeriveEdges(ctx, any).AnyTimes().Return(
			[]newinteraction.Edge{edgeB, edgeE}, nil,
		)
		nodeD.EXPECT().DeriveEdges(ctx, any).AnyTimes().Return(
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

func (n *testGraphGetAMRnode) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *testGraphGetAMRnode) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return nil, nil
}

func (n *testGraphGetAMRnode) UserAuthenticator() (newinteraction.AuthenticationStage, *authenticator.Info) {
	return n.Stage, n.Authenticator
}

func TestGraphGetAMR(t *testing.T) {
	Convey("GraphGetAMR", t, func() {
		var graph *newinteraction.Graph

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
		So(graph.GetAMR(), ShouldResemble, []string{"pwd"})

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
		So(graph.GetAMR(), ShouldResemble, []string{"otp", "sms"})

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
		So(graph.GetAMR(), ShouldResemble, []string{"mfa", "otp", "pwd"})

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
		So(graph.GetAMR(), ShouldResemble, []string{"mfa", "otp", "pwd", "sms"})

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
		So(graph.GetAMR(), ShouldResemble, []string{"mfa", "otp"})
	})
}
