package interaction_test

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
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

		ctx := &interaction.Context{}
		g := &interaction.Graph{Nodes: []interaction.Node{nodeA}}

		nodeA.EXPECT().Prepare(ctx, any).AnyTimes().Return(nil)
		nodeA.EXPECT().DeriveEdges(any).AnyTimes().Return(
			[]interaction.Edge{edgeB, edgeC}, nil,
		)
		nodeB.EXPECT().Prepare(ctx, any).AnyTimes().Return(nil)
		nodeB.EXPECT().DeriveEdges(any).AnyTimes().Return(
			[]interaction.Edge{edgeD}, nil,
		)
		nodeC.EXPECT().Prepare(ctx, any).AnyTimes().Return(nil)
		nodeC.EXPECT().DeriveEdges(any).AnyTimes().Return(
			[]interaction.Edge{edgeB, edgeE}, nil,
		)
		nodeD.EXPECT().Prepare(ctx, any).AnyTimes().Return(nil)
		nodeD.EXPECT().DeriveEdges(any).AnyTimes().Return(
			[]interaction.Edge{}, nil,
		)

		edgeB.EXPECT().Instantiate(ctx, any, any).AnyTimes().DoAndReturn(
			func(ctx *interaction.Context, g *interaction.Graph, input interface{}) (interaction.Node, error) {
				if _, ok := input.(input1); ok {
					return nodeB, nil
				}
				if _, ok := input.(input2); ok {
					return nodeB, nil
				}
				return nil, interaction.ErrIncompatibleInput
			})
		edgeC.EXPECT().Instantiate(ctx, any, any).AnyTimes().DoAndReturn(
			func(ctx *interaction.Context, g *interaction.Graph, input interface{}) (interaction.Node, error) {
				if _, ok := input.(input3); ok {
					return nodeC, nil
				}
				return nil, interaction.ErrIncompatibleInput
			})
		edgeD.EXPECT().Instantiate(ctx, any, any).AnyTimes().DoAndReturn(
			func(ctx *interaction.Context, g *interaction.Graph, input interface{}) (interaction.Node, error) {
				if _, ok := input.(input2); ok {
					return nodeD, nil
				}
				return nil, interaction.ErrIncompatibleInput
			})
		edgeE.EXPECT().Instantiate(ctx, any, any).AnyTimes().DoAndReturn(
			func(ctx *interaction.Context, g *interaction.Graph, input interface{}) (interaction.Node, error) {
				if _, ok := input.(input4); ok {
					return nil, interaction.ErrSameNode
				}
				return nil, interaction.ErrIncompatibleInput
			})

		Convey("should go to deepest node", func() {
			var inputRequired *interaction.ErrInputRequired

			nodeB.EXPECT().GetEffects()
			graph, edges, err := g.Accept(ctx, input1{})
			So(errors.As(err, &inputRequired), ShouldBeTrue)
			So(graph.Nodes, ShouldResemble, []interaction.Node{nodeA, nodeB})
			So(edges, ShouldResemble, []interaction.Edge{edgeD})

			nodeB.EXPECT().GetEffects()
			nodeD.EXPECT().GetEffects()
			graph, edges, err = g.Accept(ctx, input2{})
			So(err, ShouldBeNil)
			So(graph.Nodes, ShouldResemble, []interaction.Node{nodeA, nodeB, nodeD})
			So(edges, ShouldResemble, []interaction.Edge{})

			nodeC.EXPECT().GetEffects()
			graph, edges, err = g.Accept(ctx, input3{})
			So(errors.As(err, &inputRequired), ShouldBeTrue)
			So(graph.Nodes, ShouldResemble, []interaction.Node{nodeA, nodeC})
			So(edges, ShouldResemble, []interaction.Edge{edgeB, edgeE})

			nodeB.EXPECT().GetEffects()
			nodeD.EXPECT().GetEffects()
			graph, edges, err = graph.Accept(ctx, input2{})
			So(err, ShouldBeNil)
			So(graph.Nodes, ShouldResemble, []interaction.Node{nodeA, nodeC, nodeB, nodeD})
			So(edges, ShouldResemble, []interaction.Edge{})
		})

		Convey("should process looping edge", func() {
			var inputRequired *interaction.ErrInputRequired

			nodeC.EXPECT().GetEffects()
			graph, edges, err := g.Accept(ctx, input3{})
			So(errors.As(err, &inputRequired), ShouldBeTrue)
			So(graph.Nodes, ShouldResemble, []interaction.Node{nodeA, nodeC})
			So(edges, ShouldResemble, []interaction.Edge{edgeB, edgeE})

			graph, edges, err = graph.Accept(ctx, input4{})
			So(errors.As(err, &inputRequired), ShouldBeTrue)
			So(graph.Nodes, ShouldResemble, []interaction.Node{nodeA, nodeC})
			So(edges, ShouldResemble, []interaction.Edge{edgeB, edgeE})

			nodeB.EXPECT().GetEffects()
			nodeD.EXPECT().GetEffects()
			graph, edges, err = graph.Accept(ctx, input2{})
			So(err, ShouldBeNil)
			So(graph.Nodes, ShouldResemble, []interaction.Node{nodeA, nodeC, nodeB, nodeD})
			So(edges, ShouldResemble, []interaction.Edge{})
		})
	})
}

type testGraphGetAMRnode struct {
	Stage         interaction.AuthenticationStage
	Authenticator *authenticator.Info
}

func (n *testGraphGetAMRnode) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *testGraphGetAMRnode) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *testGraphGetAMRnode) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return nil, nil
}

func (n *testGraphGetAMRnode) UserAuthenticator(stage interaction.AuthenticationStage) (*authenticator.Info, bool) {
	if n.Stage == stage && n.Authenticator != nil {
		return n.Authenticator, true
	}
	return nil, false
}

func TestGraphGetAMRACR(t *testing.T) {
	Convey("GraphGetAMRACR", t, func() {
		var graph *interaction.Graph
		var amr []string

		graph = &interaction.Graph{}
		So(graph.GetAMR(), ShouldBeEmpty)

		// password
		graph = &interaction.Graph{
			Nodes: []interaction.Node{
				&testGraphGetAMRnode{
					Stage: interaction.AuthenticationStagePrimary,
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
		graph = &interaction.Graph{
			Nodes: []interaction.Node{
				&testGraphGetAMRnode{
					Stage: interaction.AuthenticationStagePrimary,
					Authenticator: &authenticator.Info{
						Type: authn.AuthenticatorTypeOOB,
						Claims: map[string]interface{}{
							authenticator.AuthenticatorClaimOOBOTPChannelType: string(authn.AuthenticatorOOBChannelSMS),
						},
					},
				},
			},
		}
		amr = graph.GetAMR()
		So(amr, ShouldResemble, []string{"otp", "sms"})
		So(graph.GetACR(amr), ShouldEqual, "")

		// password + email oob
		graph = &interaction.Graph{
			Nodes: []interaction.Node{
				&testGraphGetAMRnode{
					Stage: interaction.AuthenticationStagePrimary,
					Authenticator: &authenticator.Info{
						Type: authn.AuthenticatorTypePassword,
					},
				},
				&testGraphGetAMRnode{
					Stage: interaction.AuthenticationStageSecondary,
					Authenticator: &authenticator.Info{
						Type: authn.AuthenticatorTypeOOB,
						Claims: map[string]interface{}{
							authenticator.AuthenticatorClaimOOBOTPChannelType: string(authn.AuthenticatorOOBChannelEmail),
						},
					},
				},
			},
		}
		amr = graph.GetAMR()
		So(amr, ShouldResemble, []string{"mfa", "otp", "pwd"})
		So(graph.GetACR(amr), ShouldEqual, authn.ACRMFA)

		// password + SMS oob
		graph = &interaction.Graph{
			Nodes: []interaction.Node{
				&testGraphGetAMRnode{
					Stage: interaction.AuthenticationStagePrimary,
					Authenticator: &authenticator.Info{
						Type: authn.AuthenticatorTypePassword,
					},
				},
				&testGraphGetAMRnode{
					Stage: interaction.AuthenticationStageSecondary,
					Authenticator: &authenticator.Info{
						Type: authn.AuthenticatorTypeOOB,
						Claims: map[string]interface{}{
							authenticator.AuthenticatorClaimOOBOTPChannelType: string(authn.AuthenticatorOOBChannelSMS),
						},
					},
				},
			},
		}
		amr = graph.GetAMR()
		So(amr, ShouldResemble, []string{"mfa", "otp", "pwd", "sms"})
		So(graph.GetACR(amr), ShouldEqual, authn.ACRMFA)

		// oauth + totp
		graph = &interaction.Graph{
			Nodes: []interaction.Node{
				&testGraphGetAMRnode{
					Stage: interaction.AuthenticationStageSecondary,
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

type findNodeA struct{ interaction.Node }

func (*findNodeA) A() {}

type findNodeB struct{ interaction.Node }

func (*findNodeB) B() {}

type findNodeC struct{ interaction.Node }

func (*findNodeC) B() {}
func (*findNodeC) C() {}

func TestGraphFindLastNode(t *testing.T) {
	Convey("Graph.FindLastNode", t, func() {
		nodeA := &findNodeA{}
		nodeB := &findNodeB{}
		nodeC := &findNodeC{}
		graph := &interaction.Graph{
			Nodes: []interaction.Node{nodeA, nodeB, nodeC},
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
