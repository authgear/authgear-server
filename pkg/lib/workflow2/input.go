package workflow2

import (
	"context"
	"errors"
	"reflect"
)

type InputReactor interface {
	CanReactTo(ctx context.Context, deps *Dependencies, workflows Workflows) ([]Input, error)
	ReactTo(ctx context.Context, deps *Dependencies, workflows Workflows, input Input) (*Node, error)
}

type Input interface {
	Kinder
	JSONSchemaGetter
}

func AsInput(i Input, iface interface{}) bool {
	if i == nil {
		return false
	}
	val := reflect.ValueOf(iface)
	typ := val.Type()
	targetType := typ.Elem()
	for {
		if reflect.TypeOf(i).AssignableTo(targetType) {
			val.Elem().Set(reflect.ValueOf(i))
			return true
		}
		if x, ok := i.(interface{ Input() Input }); ok {
			i = x.Input()
		} else {
			break
		}
	}
	return false
}

func FindInputReactor(ctx context.Context, deps *Dependencies, workflows Workflows) (Workflows, InputReactor, Boundary, error) {
	var boundary Boundary
	return FindInputReactorForWorkflow(ctx, deps, workflows, boundary)
}

func FindInputReactorForWorkflow(ctx context.Context, deps *Dependencies, workflows Workflows, boundary Boundary) (Workflows, InputReactor, Boundary, error) {
	if len(workflows.Nearest.Nodes) > 0 {
		// We check the last node if it can react to input first.
		lastNode := workflows.Nearest.Nodes[len(workflows.Nearest.Nodes)-1]
		newWorkflows, inputReactor, boundary, err := FindInputReactorForNode(ctx, deps, workflows, boundary, &lastNode)
		if err == nil {
			return newWorkflows, inputReactor, boundary, nil
		}
		// Return non ErrEOF error.
		if err != nil && !errors.Is(err, ErrEOF) {
			return workflows, nil, nil, err
		}
		// err is ErrEOF, fallthrough
	}

	// Otherwise we check if the intent can react to input.
	_, err := workflows.Nearest.Intent.CanReactTo(ctx, deps, workflows)
	if err == nil {
		// Update boundary
		if b, ok := workflows.Nearest.Intent.(Boundary); ok {
			boundary = b
		}
		return workflows, workflows.Nearest.Intent, boundary, nil
	}

	// err != nil here.
	// Regardless of whether err is ErrEOF, we return err.
	return workflows, nil, nil, err
}

func IsEOF(ctx context.Context, deps *Dependencies, workflows Workflows) (bool, error) {
	_, _, _, err := FindInputReactorForWorkflow(ctx, deps, workflows, nil)
	if err != nil {
		if errors.Is(err, ErrEOF) {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

func FindInputReactorForNode(ctx context.Context, deps *Dependencies, workflows Workflows, boundary Boundary, n *Node) (Workflows, InputReactor, Boundary, error) {
	switch n.Type {
	case NodeTypeSimple:
		reactor, ok := n.Simple.(InputReactor)
		if !ok {
			return workflows, nil, nil, ErrEOF
		}

		_, err := reactor.CanReactTo(ctx, deps, workflows)
		if err == nil {
			if b, ok := reactor.(Boundary); ok {
				boundary = b
			}
			return workflows, reactor, boundary, nil
		}
		return workflows, nil, nil, err
	case NodeTypeSubWorkflow:
		return FindInputReactorForWorkflow(ctx, deps, workflows.Replace(n.SubWorkflow), boundary)
	default:
		panic(errors.New("unreachable"))
	}
}
