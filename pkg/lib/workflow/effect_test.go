package workflow

import (
	"context"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestApplyRunEffects(t *testing.T) {
	Convey("ApplyRunEffects", t, func() {
		test := func(w *Workflow, expectedEffect string) {
			var buf strings.Builder
			ctx := context.Background()
			ctx = WithEffectWriter(ctx, &buf)
			deps := &Dependencies{}
			err := w.ApplyRunEffects(ctx, deps, NewWorkflows(w))
			So(err, ShouldBeNil)
			So(buf.String(), ShouldEqual, expectedEffect)
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
		}, `run-effect: node0-0
run-effect: node0-1
run-effect: node0-2
`)
	})
}

func TestApplyAllEffects(t *testing.T) {
	Convey("ApplyAllEffects", t, func() {
		test := func(w *Workflow, expectedEffect string) {
			var buf strings.Builder
			ctx := context.Background()
			ctx = WithEffectWriter(ctx, &buf)
			deps := &Dependencies{}
			err := w.ApplyAllEffects(ctx, deps, NewWorkflows(w))
			So(err, ShouldBeNil)
			So(buf.String(), ShouldEqual, expectedEffect)
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
		}, `run-effect: node0-0
run-effect: node0-1
run-effect: node0-2
on-commit-effect: node0-0
on-commit-effect: node0-1
on-commit-effect: intent0-1
on-commit-effect: node0-2
on-commit-effect: intent0-0
`)
	})
}
