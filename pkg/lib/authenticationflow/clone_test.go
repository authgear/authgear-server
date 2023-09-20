package authenticationflow

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCloneFlow(t *testing.T) {
	flow := &Flow{
		FlowID:  "wf-0",
		StateID: "wf-0-instance-0",
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
				Type: NodeTypeSubFlow,
				SubFlow: &Flow{
					FlowID:  "wf-1",
					StateID: "wf-1-instance-0",
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

	Convey("CloneFlow", t, func() {
		cloned := CloneFlow(flow)

		cloned.Nodes[2].SubFlow.Nodes = append(
			cloned.Nodes[2].SubFlow.Nodes,
			Node{
				Type: NodeTypeSimple,
				Simple: &testMarshalNode0{
					Node0: "node0-2",
				},
			},
		)

		So(cloned, ShouldNotResemble, flow)
	})
}
