package workflow

import (
	"context"
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	RegisterPrivateIntent(&intentNoValidateDuringUnmarshal{})
}

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

var intentNoValidateDuringUnmarshalSchema = validation.NewSimpleSchema(`{
	"type": "object",
	"additionalProperties": false
}`)

type intentNoValidateDuringUnmarshal struct {
	SomethingInternal string `json:"something_internal"`
}

func (*intentNoValidateDuringUnmarshal) Kind() string {
	return "intentNoValidateDuringUnmarshal"
}

func (*intentNoValidateDuringUnmarshal) JSONSchema() *validation.SimpleSchema {
	return intentNoValidateDuringUnmarshalSchema
}

func (i *intentNoValidateDuringUnmarshal) GetEffects(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Effect, error) {
	return nil, nil
}

func (*intentNoValidateDuringUnmarshal) CanReactTo(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Input, error) {
	return nil, ErrEOF
}

func (intentNoValidateDuringUnmarshal) ReactTo(ctx context.Context, deps *Dependencies, workflow *Workflow, input Input) (*Node, error) {
	return nil, ErrIncompatibleInput
}

func (i *intentNoValidateDuringUnmarshal) OutputData(ctx context.Context, deps *Dependencies, workflow *Workflow) (interface{}, error) {
	return nil, nil
}

func TestPublicPrivateIntentMarshal(t *testing.T) {
	Convey("Instantiate intent from private registry, add something internal, and then marshal and unmarshal", t, func() {
		intentJSON := IntentJSON{
			Kind: "intentNoValidateDuringUnmarshal",
			Data: json.RawMessage("{}"),
		}
		intent, err := InstantiateIntentFromPrivateRegistry(intentJSON)
		i, _ := intent.(*intentNoValidateDuringUnmarshal)
		i.SomethingInternal = "secret"

		So(err, ShouldBeNil)

		workflow := &Workflow{
			Intent: intent,
		}

		workflowBytes, err := json.Marshal(workflow)
		So(err, ShouldBeNil)

		var w Workflow
		err = json.Unmarshal(workflowBytes, &w)
		So(err, ShouldBeNil)
		So(&w, ShouldResemble, &Workflow{
			Intent: &intentNoValidateDuringUnmarshal{
				SomethingInternal: "secret",
			},
		})
	})
}
