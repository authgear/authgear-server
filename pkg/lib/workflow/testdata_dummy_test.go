package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

func init() {
	RegisterIntent(&testMarshalIntent0{})
	RegisterIntent(&testMarshalIntent1{})
	RegisterNode(&testMarshalNode0{})
	RegisterNode(&testMarshalNode1{})
}

func WithEffectWriter(ctx context.Context, w io.Writer) context.Context {
	return context.WithValue(ctx, "writer", w)
}

func GetEffectWriter(ctx context.Context) (io.Writer, bool) {
	w, ok := ctx.Value("writer").(io.Writer)
	return w, ok
}

type testMarshalIntent0 struct {
	Intent0 string
}

func (*testMarshalIntent0) Kind() string {
	return "testMarshalIntent0"
}

func (i *testMarshalIntent0) Instantiate(data json.RawMessage) error {
	return json.Unmarshal(data, i)
}

func (i *testMarshalIntent0) GetEffects(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Effect, error) {
	return []Effect{
		OnCommitEffect(func(ctx context.Context, deps *Dependencies) error {
			if w, ok := GetEffectWriter(ctx); ok {
				fmt.Fprintf(w, "on-commit-effect: %v\n", i.Intent0)
			}
			return nil
		}),
	}, nil
}

func (*testMarshalIntent0) DeriveEdges(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Edge, error) {
	return nil, nil
}

func (i *testMarshalIntent0) OutputData(ctx context.Context, deps *Dependencies, workflow *Workflow) (interface{}, error) {
	return map[string]interface{}{
		"intent0": i.Intent0,
	}, nil
}

type testMarshalIntent1 struct {
	Intent1 string
}

func (*testMarshalIntent1) Kind() string {
	return "testMarshalIntent1"
}

func (i *testMarshalIntent1) Instantiate(data json.RawMessage) error {
	return json.Unmarshal(data, i)
}

func (i *testMarshalIntent1) GetEffects(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Effect, error) {
	return []Effect{
		OnCommitEffect(func(ctx context.Context, deps *Dependencies) error {
			if w, ok := GetEffectWriter(ctx); ok {
				fmt.Fprintf(w, "on-commit-effect: %v\n", i.Intent1)
			}
			return nil
		}),
	}, nil
}

func (*testMarshalIntent1) DeriveEdges(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Edge, error) {
	return nil, nil
}

func (i *testMarshalIntent1) OutputData(ctx context.Context, deps *Dependencies, workflow *Workflow) (interface{}, error) {
	return map[string]interface{}{
		"intent1": i.Intent1,
	}, nil
}

type testMarshalNode0 struct {
	Node0 string
}

func (*testMarshalNode0) Kind() string {
	return "testMarshalNode0"
}

func (n *testMarshalNode0) GetEffects(ctx context.Context, deps *Dependencies) ([]Effect, error) {
	return []Effect{
		RunEffect(func(ctx context.Context, deps *Dependencies) error {
			if w, ok := GetEffectWriter(ctx); ok {
				fmt.Fprintf(w, "run-effect: %v\n", n.Node0)
			}
			return nil
		}),
		OnCommitEffect(func(ctx context.Context, deps *Dependencies) error {
			if w, ok := GetEffectWriter(ctx); ok {
				fmt.Fprintf(w, "on-commit-effect: %v\n", n.Node0)
			}
			return nil
		}),
	}, nil
}

func (*testMarshalNode0) DeriveEdges(ctx context.Context, deps *Dependencies) ([]Edge, error) {
	return nil, nil
}

func (i *testMarshalNode0) OutputData(ctx context.Context, deps *Dependencies) (interface{}, error) {
	return map[string]interface{}{
		"node0": i.Node0,
	}, nil
}

type testMarshalNode1 struct {
	Node1 string
}

func (*testMarshalNode1) Kind() string {
	return "testMarshalNode1"
}

func (n *testMarshalNode1) GetEffects(ctx context.Context, deps *Dependencies) ([]Effect, error) {
	return []Effect{
		RunEffect(func(ctx context.Context, deps *Dependencies) error {
			if w, ok := GetEffectWriter(ctx); ok {
				fmt.Fprintf(w, "run-effect: %v\n", n.Node1)
			}
			return nil
		}),
		OnCommitEffect(func(ctx context.Context, deps *Dependencies) error {
			if w, ok := GetEffectWriter(ctx); ok {
				fmt.Fprintf(w, "on-commit-effect: %v\n", n.Node1)
			}
			return nil
		}),
	}, nil
}

func (*testMarshalNode1) DeriveEdges(ctx context.Context, deps *Dependencies) ([]Edge, error) {
	return nil, nil
}

func (i *testMarshalNode1) OutputData(ctx context.Context, deps *Dependencies) (interface{}, error) {
	return map[string]interface{}{
		"node1": i.Node1,
	}, nil
}
