package workflow

import (
	"context"
)

func init() {
	RegisterIntent(&testMarshalIntent0{})
	RegisterIntent(&testMarshalIntent1{})
	RegisterNode(&testMarshalNode0{})
	RegisterNode(&testMarshalNode1{})
}

type testMarshalIntent0 struct {
	Intent0 string
}

func (*testMarshalIntent0) GetEffects(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Effect, error) {
	return nil, nil
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

func (*testMarshalIntent1) GetEffects(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Effect, error) {
	return nil, nil
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

func (*testMarshalNode0) GetEffects(ctx context.Context, deps *Dependencies) ([]Effect, error) {
	return nil, nil
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

func (*testMarshalNode1) GetEffects(ctx context.Context, deps *Dependencies) ([]Effect, error) {
	return nil, nil
}

func (*testMarshalNode1) DeriveEdges(ctx context.Context, deps *Dependencies) ([]Edge, error) {
	return nil, nil
}

func (i *testMarshalNode1) OutputData(ctx context.Context, deps *Dependencies) (interface{}, error) {
	return map[string]interface{}{
		"node1": i.Node1,
	}, nil
}
