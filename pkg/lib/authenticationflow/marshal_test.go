package authenticationflow

import (
	"context"
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func init() {
	RegisterIntent(&intentNoValidateDuringUnmarshal{})
}

func TestMarshalJSON(t *testing.T) {
	flow := &Flow{
		FlowID:     "wf-0",
		StateToken: "wf-0-state-0",
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
					FlowID:     "wf-1",
					StateToken: "wf-1-state-0",
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
    "state_token": "wf-0-state-0",
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
            "type": "SUB_FLOW",
            "flow": {
                "state_token": "wf-1-state-0",
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
                "flow_id": "wf-1"
            }
        }
    ],
    "flow_id": "wf-0"
}
	`

	Convey("MarshalJSON", t, func() {
		bytes, err := json.MarshalIndent(flow, "", "  ")
		So(err, ShouldBeNil)
		So(string(bytes), ShouldEqualJSON, jsonString)
	})

	Convey("UnmarshalJSON", t, func() {
		var w Flow
		err := json.Unmarshal([]byte(jsonString), &w)
		So(err, ShouldBeNil)
		So(&w, ShouldResemble, flow)
	})
}

type intentNoValidateDuringUnmarshal struct {
	SomethingInternal string `json:"something_internal"`
}

var _ Intent = &intentNoValidateDuringUnmarshal{}

func (*intentNoValidateDuringUnmarshal) Kind() string {
	return "intentNoValidateDuringUnmarshal"
}

func (*intentNoValidateDuringUnmarshal) CanReactTo(ctx context.Context, deps *Dependencies, flows Flows) (InputSchema, error) {
	return nil, ErrEOF
}

func (intentNoValidateDuringUnmarshal) ReactTo(ctx context.Context, deps *Dependencies, flows Flows, input Input) (ReactToResult, error) {
	return nil, ErrIncompatibleInput
}

func TestPublicPrivateIntentMarshal(t *testing.T) {
	Convey("Instantiate intent from private registry, add something internal, and then marshal and unmarshal", t, func() {
		intentJSON := intentJSON{
			Kind: "intentNoValidateDuringUnmarshal",
			Data: json.RawMessage("{}"),
		}
		intent, err := InstantiateIntent(intentJSON.Kind)
		So(err, ShouldBeNil)
		err = json.Unmarshal(intentJSON.Data, intent)
		So(err, ShouldBeNil)

		i, _ := intent.(*intentNoValidateDuringUnmarshal)
		i.SomethingInternal = "secret"

		flow := &Flow{
			Intent: intent,
		}

		flowBytes, err := json.Marshal(flow)
		So(err, ShouldBeNil)

		var w Flow
		err = json.Unmarshal(flowBytes, &w)
		So(err, ShouldBeNil)
		So(&w, ShouldResemble, &Flow{
			Intent: &intentNoValidateDuringUnmarshal{
				SomethingInternal: "secret",
			},
		})
	})
}
