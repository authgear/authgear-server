package workflow

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMarshalJSON(t *testing.T) {
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

	jsonString := `
{
    "instance_id": "wf-0-instance-0",
    "intent": {
        "data": {
            "Intent0": "intent0"
        },
        "kind": "testMarshalIntent0"
    },
    "nodes": [
        {
            "simple": {
                "data": {
                    "Node0": "node0-0"
                },
                "kind": "testMarshalNode0"
            },
            "type": "SIMPLE"
        },
        {
            "simple": {
                "data": {
                    "Node1": "node1-0"
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
                        "Intent1": "intent1"
                    },
                    "kind": "testMarshalIntent1"
                },
                "nodes": [
                    {
                        "simple": {
                            "data": {
                                "Node0": "node0-1"
                            },
                            "kind": "testMarshalNode0"
                        },
                        "type": "SIMPLE"
                    },
                    {
                        "simple": {
                            "data": {
                                "Node1": "node1-1"
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
	`

	Convey("MarshalJSON", t, func() {
		bytes, err := json.MarshalIndent(workflow, "", "  ")
		So(err, ShouldBeNil)
		So(string(bytes), ShouldEqualJSON, jsonString)
	})

	Convey("UnmarshalJSON", t, func() {
		var w Workflow
		err := json.Unmarshal([]byte(jsonString), &w)
		So(err, ShouldBeNil)
		So(&w, ShouldResemble, workflow)
	})
}
