package authenticationflow

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTraverse(t *testing.T) {
	Convey("TraverseFlow", t, func() {
		test := func(w *Flow, expectedHistory []string) {
			var actualHistory []string
			err := TraverseFlow(Traverser{
				Intent: func(intent Intent, w *Flow) error {
					if i, ok := intent.(*testMarshalIntent0); ok {
						actualHistory = append(actualHistory, i.Intent0)
					}
					return nil
				},
				NodeSimple: func(nodeSimple NodeSimple, w *Flow) error {
					if n, ok := nodeSimple.(*testMarshalNode0); ok {
						actualHistory = append(actualHistory, n.Node0)
					}
					return nil
				},
			}, w)
			So(err, ShouldBeNil)
			So(actualHistory, ShouldResemble, expectedHistory)
		}

		test(&Flow{
			FlowID:     "wf-0",
			StateToken: "wf-0-state-0",
			Intent: &testMarshalIntent0{
				Intent0: "intent0-0",
			},
			Nodes: []Node{
				Node{
					Type: NodeTypeSimple,
					Simple: &testMarshalNode0{
						Node0: "node0-0",
					},
				},
				Node{
					Type: NodeTypeSubFlow,
					SubFlow: &Flow{
						Intent: &testMarshalIntent0{
							Intent0: "intent0-1",
						},
						Nodes: []Node{
							Node{
								Type: NodeTypeSimple,
								Simple: &testMarshalNode0{
									Node0: "node0-1",
								},
							},
						},
					},
				},
				Node{
					Type: NodeTypeSimple,
					Simple: &testMarshalNode0{
						Node0: "node0-2",
					},
				},
			},
		}, []string{
			"node0-0",
			"node0-1",
			"intent0-1",
			"node0-2",
			"intent0-0",
		})
	})

	Convey("TraverseFlowIntentFirst", t, func() {
		test := func(w *Flow, expectedHistory []string) {
			var actualHistory []string
			err := TraverseFlowIntentFirst(Traverser{
				Intent: func(intent Intent, w *Flow) error {
					if i, ok := intent.(*testMarshalIntent0); ok {
						actualHistory = append(actualHistory, i.Intent0)
					}
					return nil
				},
				NodeSimple: func(nodeSimple NodeSimple, w *Flow) error {
					if n, ok := nodeSimple.(*testMarshalNode0); ok {
						actualHistory = append(actualHistory, n.Node0)
					}
					return nil
				},
			}, w)
			So(err, ShouldBeNil)
			So(actualHistory, ShouldResemble, expectedHistory)
		}

		test(&Flow{
			FlowID:     "wf-0",
			StateToken: "wf-0-state-0",
			Intent: &testMarshalIntent0{
				Intent0: "intent0-0",
			},
			Nodes: []Node{
				Node{
					Type: NodeTypeSimple,
					Simple: &testMarshalNode0{
						Node0: "node0-0",
					},
				},
				Node{
					Type: NodeTypeSubFlow,
					SubFlow: &Flow{
						Intent: &testMarshalIntent0{
							Intent0: "intent0-1",
						},
						Nodes: []Node{
							Node{
								Type: NodeTypeSimple,
								Simple: &testMarshalNode0{
									Node0: "node0-1",
								},
							},
						},
					},
				},
				Node{
					Type: NodeTypeSimple,
					Simple: &testMarshalNode0{
						Node0: "node0-2",
					},
				},
			},
		}, []string{
			"intent0-0",
			"node0-0",
			"intent0-1",
			"node0-1",
			"node0-2",
		})
	})

	Convey("TraverseIntentFromEndToRoot", t, func() {
		test := func(w *Flow, expectedHistory []string) {
			var actualHistory []string
			err := TraverseIntentFromEndToRoot(func(intent Intent) error {
				if i, ok := intent.(*testMarshalIntent0); ok {
					actualHistory = append(actualHistory, i.Intent0)
				}
				return nil
			}, w)
			So(err, ShouldBeNil)
			So(actualHistory, ShouldResemble, expectedHistory)
		}

		test(&Flow{
			FlowID:     "wf-0",
			StateToken: "wf-0-state-0",
			Intent: &testMarshalIntent0{
				Intent0: "intent0-0",
			},
			Nodes: []Node{
				Node{
					Type: NodeTypeSimple,
					Simple: &testMarshalNode0{
						Node0: "node0-0",
					},
				},
				Node{
					Type: NodeTypeSubFlow,
					SubFlow: &Flow{
						Intent: &testMarshalIntent0{
							Intent0: "intent0-1",
						},
						Nodes: []Node{
							Node{
								Type: NodeTypeSimple,
								Simple: &testMarshalNode0{
									Node0: "node0-1",
								},
							},
							Node{
								Type: NodeTypeSimple,
								Simple: &testMarshalNode0{
									Node0: "node0-2",
								},
							},
						},
					},
				},
			},
		}, []string{
			"intent0-1",
			"intent0-0",
		})

		test(&Flow{
			FlowID:     "wf-0",
			StateToken: "wf-0-state-0",
			Intent: &testMarshalIntent0{
				Intent0: "intent0-0",
			},
			Nodes: []Node{
				Node{
					Type: NodeTypeSimple,
					Simple: &testMarshalNode0{
						Node0: "node0-0",
					},
				},
				Node{
					Type: NodeTypeSubFlow,
					SubFlow: &Flow{
						Intent: &testMarshalIntent0{
							Intent0: "intent0-1",
						},
						Nodes: []Node{
							Node{
								Type: NodeTypeSimple,
								Simple: &testMarshalNode0{
									Node0: "node0-1",
								},
							},
							Node{
								Type: NodeTypeSimple,
								Simple: &testMarshalNode0{
									Node0: "node0-2",
								},
							},
						},
					},
				},
				Node{
					Type: NodeTypeSimple,
					Simple: &testMarshalNode0{
						Node0: "node0-3",
					},
				},
			},
		}, []string{
			"intent0-0",
		})

		test(&Flow{
			FlowID:     "wf-0",
			StateToken: "wf-0-state-0",
			Intent: &testMarshalIntent0{
				Intent0: "intent0-0",
			},
			Nodes: []Node{
				Node{
					Type: NodeTypeSimple,
					Simple: &testMarshalNode0{
						Node0: "node0-0",
					},
				},
				Node{
					Type: NodeTypeSubFlow,
					SubFlow: &Flow{
						Intent: &testMarshalIntent0{
							Intent0: "intent0-1",
						},
						Nodes: []Node{
							Node{
								Type: NodeTypeSimple,
								Simple: &testMarshalNode0{
									Node0: "node0-1",
								},
							},
							Node{
								Type: NodeTypeSubFlow,
								SubFlow: &Flow{
									Intent: &testMarshalIntent0{
										Intent0: "intent0-2",
									},
									Nodes: []Node{
										Node{
											Type: NodeTypeSimple,
											Simple: &testMarshalNode0{
												Node0: "node0-3",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}, []string{
			"intent0-2",
			"intent0-1",
			"intent0-0",
		})
	})
}
