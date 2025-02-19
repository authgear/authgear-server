package authenticationflow

import (
	"context"
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

var _ Intent = &testMarshalIntent0{}
var _ EffectGetter = &testMarshalIntent0{}
var _ DataOutputer = &testMarshalIntent0{}

func (*testMarshalIntent0) Kind() string {
	return "testMarshalIntent0"
}

func (i *testMarshalIntent0) GetEffects(ctx context.Context, deps *Dependencies, flows Flows) ([]Effect, error) {
	return []Effect{
		OnCommitEffect(func(ctx context.Context, deps *Dependencies) error {
			if w, ok := GetEffectWriter(ctx); ok {
				fmt.Fprintf(w, "on-commit-effect: %v\n", i.Intent0)
			}
			return nil
		}),
	}, nil
}

func (*testMarshalIntent0) CanReactTo(ctx context.Context, deps *Dependencies, flows Flows) (InputSchema, error) {
	return nil, ErrEOF
}

func (testMarshalIntent0) ReactTo(ctx context.Context, deps *Dependencies, flows Flows, input Input) (ReactToResult, error) {
	panic("unreachable")
}

func (i *testMarshalIntent0) OutputData(ctx context.Context, deps *Dependencies, flows Flows) (Data, error) {
	return mapData{
		"intent0": i.Intent0,
	}, nil
}

type testMarshalIntent1 struct {
	Intent1 string
}

var _ Intent = &testMarshalIntent1{}
var _ EffectGetter = &testMarshalIntent1{}
var _ DataOutputer = &testMarshalIntent1{}

func (*testMarshalIntent1) Kind() string {
	return "testMarshalIntent1"
}

func (i *testMarshalIntent1) GetEffects(ctx context.Context, deps *Dependencies, flows Flows) ([]Effect, error) {
	return []Effect{
		OnCommitEffect(func(ctx context.Context, deps *Dependencies) error {
			if w, ok := GetEffectWriter(ctx); ok {
				fmt.Fprintf(w, "on-commit-effect: %v\n", i.Intent1)
			}
			return nil
		}),
	}, nil
}

func (*testMarshalIntent1) CanReactTo(ctx context.Context, deps *Dependencies, flows Flows) (InputSchema, error) {
	return nil, ErrEOF
}

func (testMarshalIntent1) ReactTo(ctx context.Context, deps *Dependencies, flows Flows, input Input) (ReactToResult, error) {
	panic("unreachable")
}

func (i *testMarshalIntent1) OutputData(ctx context.Context, deps *Dependencies, flows Flows) (Data, error) {
	return mapData{
		"intent1": i.Intent1,
	}, nil
}

type testMarshalNode0 struct {
	Node0 string
}

var _ NodeSimple = &testMarshalNode0{}
var _ EffectGetter = &testMarshalNode0{}
var _ DataOutputer = &testMarshalNode0{}

func (*testMarshalNode0) Kind() string {
	return "testMarshalNode0"
}

func (n *testMarshalNode0) GetEffects(ctx context.Context, deps *Dependencies, flows Flows) ([]Effect, error) {
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

func (i *testMarshalNode0) CanReactTo(ctx context.Context, deps *Dependencies, flows Flows) (InputSchema, error) {
	return nil, ErrEOF
}

func (i *testMarshalNode0) ReactTo(ctx context.Context, deps *Dependencies, flows Flows, input Input) (ReactToResult, error) {
	return nil, ErrIncompatibleInput
}

func (i *testMarshalNode0) OutputData(ctx context.Context, deps *Dependencies, flows Flows) (Data, error) {
	return mapData{
		"node0": i.Node0,
	}, nil
}

type testMarshalNode1 struct {
	Node1 string
}

var _ NodeSimple = &testMarshalNode1{}
var _ EffectGetter = &testMarshalNode1{}
var _ DataOutputer = &testMarshalNode1{}

func (*testMarshalNode1) Kind() string {
	return "testMarshalNode1"
}

func (n *testMarshalNode1) GetEffects(ctx context.Context, deps *Dependencies, flows Flows) ([]Effect, error) {
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

func (i *testMarshalNode1) CanReactTo(ctx context.Context, deps *Dependencies, flows Flows) (InputSchema, error) {
	return nil, ErrEOF
}

func (i *testMarshalNode1) ReactTo(ctx context.Context, deps *Dependencies, flows Flows, input Input) (ReactToResult, error) {
	return nil, ErrIncompatibleInput
}

func (i *testMarshalNode1) OutputData(ctx context.Context, deps *Dependencies, flows Flows) (Data, error) {
	return mapData{
		"node1": i.Node1,
	}, nil
}
