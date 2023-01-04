package workflow

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestOutput(t *testing.T) {
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

	Convey("ToOutput", t, func() {
		ctx := &Context{}
		output, err := workflow.ToOutput(ctx)
		So(err, ShouldBeNil)
		bytes, err := json.MarshalIndent(output, "", "  ")
		So(err, ShouldBeNil)
		So(string(bytes), ShouldEqualJSON, `
{
    "instance_id": "wf-0-instance-0",
    "intent": {
        "data": {
            "intent0": "intent0"
        },
        "kind": "testMarshalIntent0"
    },
    "nodes": [
        {
            "simple": {
                "data": {
                    "node0": "node0-0"
                },
                "kind": "testMarshalNode0"
            },
            "type": "SIMPLE"
        },
        {
            "simple": {
                "data": {
                    "node1": "node1-0"
                },
                "kind": "testMarshalNode1"
            },
            "type": "SIMPLE"
        },
        {
            "type": "SUB_WORKFLOW",
            "workflow": {
                "instance_id": "wf-1-instance-0",
                "intent": {
                    "data": {
                        "intent1": "intent1"
                    },
                    "kind": "testMarshalIntent1"
                },
                "nodes": [
                    {
                        "simple": {
                            "data": {
                                "node0": "node0-1"
                            },
                            "kind": "testMarshalNode0"
                        },
                        "type": "SIMPLE"
                    },
                    {
                        "simple": {
                            "data": {
                                "node1": "node1-1"
                            },
                            "kind": "testMarshalNode1"
                        },
                        "type": "SIMPLE"
                    }
                ],
                "workflow_id": "wf-1"
            }
        }
    ],
    "workflow_id": "wf-0"
}
		`)
	})
}
