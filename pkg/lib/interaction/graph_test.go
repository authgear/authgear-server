package interaction

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

type input1 struct{}

func (input1) IsInteractive() bool { return false }

type input2 struct{}

func (input2) IsInteractive() bool { return false }

type input3 struct{}

func (input3) IsInteractive() bool { return false }

type input4 struct{}

func (input4) IsInteractive() bool { return false }

func TestGraph(t *testing.T) {
	Convey("Graph.accept", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		any := gomock.Any()

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
		g := &Graph{AnnotatedNodes: []AnnotatedNode{AnnotatedNode{Node: nodeA}}}

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
			So(graph.AnnotatedNodes, ShouldResemble, []AnnotatedNode{AnnotatedNode{Node: nodeA}, AnnotatedNode{Node: nodeB}})
			So(edges, ShouldResemble, []Edge{edgeD})

			nodeB.EXPECT().GetEffects()
			nodeD.EXPECT().GetEffects()
			graph, edges, err = g.accept(ctx, input2{})
			So(err, ShouldBeNil)
			So(graph.AnnotatedNodes, ShouldResemble, []AnnotatedNode{AnnotatedNode{Node: nodeA}, AnnotatedNode{Node: nodeB}, AnnotatedNode{Node: nodeD}})
			So(edges, ShouldResemble, []Edge{})

			nodeC.EXPECT().GetEffects()
			graph, edges, err = g.accept(ctx, input3{})
			So(errors.As(err, &inputRequired), ShouldBeTrue)
			So(graph.AnnotatedNodes, ShouldResemble, []AnnotatedNode{AnnotatedNode{Node: nodeA}, AnnotatedNode{Node: nodeC}})
			So(edges, ShouldResemble, []Edge{edgeB, edgeE})

			nodeB.EXPECT().GetEffects()
			nodeD.EXPECT().GetEffects()
			graph, edges, err = graph.accept(ctx, input2{})
			So(err, ShouldBeNil)
			So(graph.AnnotatedNodes, ShouldResemble, []AnnotatedNode{AnnotatedNode{Node: nodeA}, AnnotatedNode{Node: nodeC}, AnnotatedNode{Node: nodeB}, AnnotatedNode{Node: nodeD}})
			So(edges, ShouldResemble, []Edge{})
		})

		Convey("should process looping edge", func() {
			var inputRequired *ErrInputRequired

			nodeC.EXPECT().GetEffects()
			graph, edges, err := g.accept(ctx, input3{})
			So(errors.As(err, &inputRequired), ShouldBeTrue)
			So(graph.AnnotatedNodes, ShouldResemble, []AnnotatedNode{AnnotatedNode{Node: nodeA}, AnnotatedNode{Node: nodeC}})
			So(edges, ShouldResemble, []Edge{edgeB, edgeE})

			graph, edges, err = graph.accept(ctx, input4{})
			So(errors.As(err, &inputRequired), ShouldBeTrue)
			So(graph.AnnotatedNodes, ShouldResemble, []AnnotatedNode{AnnotatedNode{Node: nodeA}, AnnotatedNode{Node: nodeC}})
			So(edges, ShouldResemble, []Edge{edgeB, edgeE})

			nodeB.EXPECT().GetEffects()
			nodeD.EXPECT().GetEffects()
			graph, edges, err = graph.accept(ctx, input2{})
			So(err, ShouldBeNil)
			So(graph.AnnotatedNodes, ShouldResemble, []AnnotatedNode{AnnotatedNode{Node: nodeA}, AnnotatedNode{Node: nodeC}, AnnotatedNode{Node: nodeB}, AnnotatedNode{Node: nodeD}})
			So(edges, ShouldResemble, []Edge{})
		})
	})
}

type testGraphGetAMRnode struct {
	Stage         authn.AuthenticationStage
	Identity      *identity.Info
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
			AnnotatedNodes: []AnnotatedNode{
				AnnotatedNode{
					Node: &testGraphGetAMRnode{
						Stage: authn.AuthenticationStagePrimary,
						Authenticator: &authenticator.Info{
							Type: authn.AuthenticatorTypePassword,
						},
					},
				},
			},
		}
		amr = graph.GetAMR()
		So(amr, ShouldResemble, []string{"pwd"})

		// oob
		graph = &Graph{
			AnnotatedNodes: []AnnotatedNode{
				AnnotatedNode{
					Node: &testGraphGetAMRnode{
						Stage: authn.AuthenticationStagePrimary,
						Authenticator: &authenticator.Info{
							Type:   authn.AuthenticatorTypeOOBSMS,
							Claims: map[string]interface{}{},
						},
					},
				},
			},
		}
		amr = graph.GetAMR()
		So(amr, ShouldResemble, []string{"otp", "sms"})

		// password + email oob
		graph = &Graph{
			AnnotatedNodes: []AnnotatedNode{
				AnnotatedNode{
					Node: &testGraphGetAMRnode{
						Stage: authn.AuthenticationStagePrimary,
						Authenticator: &authenticator.Info{
							Type: authn.AuthenticatorTypePassword,
						},
					},
				},
				AnnotatedNode{
					Node: &testGraphGetAMRnode{
						Stage: authn.AuthenticationStageSecondary,
						Authenticator: &authenticator.Info{
							Type:   authn.AuthenticatorTypeOOBEmail,
							Claims: map[string]interface{}{},
						},
					},
				},
			},
		}
		amr = graph.GetAMR()
		So(amr, ShouldResemble, []string{"mfa", "otp", "pwd"})

		// password + SMS oob
		graph = &Graph{
			AnnotatedNodes: []AnnotatedNode{
				AnnotatedNode{
					Node: &testGraphGetAMRnode{
						Stage: authn.AuthenticationStagePrimary,
						Authenticator: &authenticator.Info{
							Type: authn.AuthenticatorTypePassword,
						},
					},
				},
				AnnotatedNode{
					Node: &testGraphGetAMRnode{
						Stage: authn.AuthenticationStageSecondary,
						Authenticator: &authenticator.Info{
							Type:   authn.AuthenticatorTypeOOBSMS,
							Claims: map[string]interface{}{},
						},
					},
				},
			},
		}
		amr = graph.GetAMR()
		So(amr, ShouldResemble, []string{"mfa", "otp", "pwd", "sms"})

		// oauth + totp
		graph = &Graph{
			AnnotatedNodes: []AnnotatedNode{
				AnnotatedNode{
					Node: &testGraphGetAMRnode{
						Stage: authn.AuthenticationStageSecondary,
						Authenticator: &authenticator.Info{
							Type: authn.AuthenticatorTypeTOTP,
						},
					},
				},
			},
		}
		amr = graph.GetAMR()
		So(amr, ShouldResemble, []string{"mfa", "otp"})

		// biometric
		graph = &Graph{
			AnnotatedNodes: []AnnotatedNode{
				AnnotatedNode{
					Node: &testGraphGetAMRnode{
						Identity: &identity.Info{
							Type: authn.IdentityTypeBiometric,
						},
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
			AnnotatedNodes: []AnnotatedNode{
				AnnotatedNode{
					Node: nodeA,
				},
				AnnotatedNode{
					Node: nodeB,
				},
				AnnotatedNode{
					Node: nodeC,
				},
			},
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

type testNodeJSONEncoding struct {
	Str string
}

func (n *testNodeJSONEncoding) Prepare(ctx *Context, graph *Graph) error {
	return nil
}
func (n *testNodeJSONEncoding) GetEffects() ([]Effect, error) {
	return nil, nil
}
func (n *testNodeJSONEncoding) DeriveEdges(graph *Graph) ([]Edge, error) {
	return nil, nil
}

type testIntentJSONEncoding struct {
	Str string
}

func (i *testIntentJSONEncoding) InstantiateRootNode(ctx *Context, graph *Graph) (Node, error) {
	return nil, nil
}

func (i *testIntentJSONEncoding) DeriveEdgesForNode(graph *Graph, node Node) ([]Edge, error) {
	return nil, nil
}

func TestGraphJSONEncoding(t *testing.T) {
	RegisterNode(&testNodeJSONEncoding{})
	RegisterIntent(&testIntentJSONEncoding{})

	Convey("Graph JSON encoding", t, func() {

		Convey("graph with annotated nodes", func() {
			g := &Graph{
				GraphID:    "a",
				InstanceID: "b",
				Intent: &testIntentJSONEncoding{
					Str: "intent",
				},
				AnnotatedNodes: []AnnotatedNode{
					AnnotatedNode{
						Interactive: true,
						Node: &testNodeJSONEncoding{
							Str: "node1",
						},
					},
				},
			}

			b, err := json.Marshal(g)
			So(err, ShouldBeNil)

			var gg Graph
			err = json.Unmarshal(b, &gg)
			So(err, ShouldBeNil)

			So(&gg, ShouldResemble, g)
		})

		Convey("graph with legacy nodes", func() {
			g := &Graph{
				GraphID:    "a",
				InstanceID: "b",
				Intent: &testIntentJSONEncoding{
					Str: "intent",
				},
				AnnotatedNodes: []AnnotatedNode{
					AnnotatedNode{
						Interactive: false,
						Node: &testNodeJSONEncoding{
							Str: "node1",
						},
					},
				},
			}

			j := `
			{
				"graph_id": "a",
				"instance_id": "b",
				"intent": {
					"Kind": "testIntentJSONEncoding",
					"Data": {
						"Str": "intent"
					}
				},
				"nodes": [
					{
						"Kind": "testNodeJSONEncoding",
						"Data": {
							"Str": "node1"
						}
					}
				]
			}
			`

			var gg Graph
			err := json.Unmarshal([]byte(j), &gg)
			So(err, ShouldBeNil)

			So(&gg, ShouldResemble, g)
		})
	})
}
