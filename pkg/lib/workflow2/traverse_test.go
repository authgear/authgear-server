package workflow2

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTraverse(t *testing.T) {
	Convey("Traverse", t, func() {
		test := func(w *Workflow, expectedHistory []string) {
			var actualHistory []string
			err := TraverseWorkflow(WorkflowTraverser{
				Intent: func(intent Intent, w *Workflow) error {
					if i, ok := intent.(*testMarshalIntent0); ok {
						actualHistory = append(actualHistory, i.Intent0)
					}
					return nil
				},
				NodeSimple: func(nodeSimple NodeSimple, w *Workflow) error {
					if n, ok := nodeSimple.(*testMarshalNode0); ok {
						actualHistory = append(actualHistory, n.Node0)
					}
					return nil
				},
			}, w)
			So(err, ShouldBeNil)
			So(actualHistory, ShouldResemble, expectedHistory)
		}

		test(&Workflow{
			WorkflowID: "wf-0",
			InstanceID: "wf-0-instance-0",
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
					Type: NodeTypeSubWorkflow,
					SubWorkflow: &Workflow{
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
}
