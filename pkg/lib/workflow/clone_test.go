package workflow

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestClone(t *testing.T) {
	workflow := &Workflow{
		WorkflowID: "wf-0",
		InstanceID: "wf-0-instance-0",
		Intent: &testMarshalIntent0{
			Intent0: "intent0",
		},
		Nodes: []Node{
			Node{
				Type: NodeTypeSimple,
				Simple: &testMarshalNode0{
					Node0: "node0-0",
				},
			},
			Node{
				Type: NodeTypeSimple,
				Simple: &testMarshalNode1{
					Node1: "node1-0",
				},
			},
			Node{
				Type: NodeTypeSubWorkflow,
				SubWorkflow: &Workflow{
					WorkflowID: "wf-1",
					InstanceID: "wf-1-instance-0",
					Intent: &testMarshalIntent1{
						Intent1: "intent1",
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
							Simple: &testMarshalNode1{
								Node1: "node1-1",
							},
						},
					},
				},
			},
		},
	}

	Convey("Clone", t, func() {
		cloned := workflow.Clone()

		cloned.Nodes[2].SubWorkflow.Nodes = append(
			cloned.Nodes[2].SubWorkflow.Nodes,
			Node{
				Type: NodeTypeSimple,
				Simple: &testMarshalNode0{
					Node0: "node0-2",
				},
			},
		)

		So(cloned, ShouldNotResemble, workflow)
	})
}
