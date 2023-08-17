package workflow2

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"

	"github.com/authgear/authgear-server/pkg/util/validation"
)

// InputReactor, if can react to some input, must return an InputSchema.
// It must react to the Input produced by its InputSchema.
// As a special case, CanReactTo can return a nil InputSchema, which means
// the InputReactor can react to a nil Input.
type InputReactor interface {
	CanReactTo(ctx context.Context, deps *Dependencies, workflows Workflows) (InputSchema, error)
	ReactTo(ctx context.Context, deps *Dependencies, workflows Workflows, input Input) (*Node, error)
}

// InputSchema is something that can produce a JSON Schema.
// In simple case, it can be implemented by the input struct itself.
// If rawMessage does not validate against the JSON schema,
// MakeInput MUST return *validation.AggregateError.
type InputSchema interface {
	SchemaBuilder() validation.SchemaBuilder
	MakeInput(rawMessage json.RawMessage) (Input, error)
}

// Input is a marker to signify some struct is an Input.
type Input interface {
	Input()
}

// InputUnwrapper is for advanced usage.
// This usage is not used at the moment.
type InputUnwrapper interface {
	Unwrap() Input
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
		if x, ok := i.(InputUnwrapper); ok {
			i = x.Unwrap()
		} else {
			break
		}
	}
	return false
}

type FindInputReactorResult struct {
	Workflows    Workflows
	InputReactor InputReactor
	InputSchema  InputSchema
	Boundary     Boundary
}

func FindInputReactor(ctx context.Context, deps *Dependencies, workflows Workflows) (*FindInputReactorResult, error) {
	var boundary Boundary
	return FindInputReactorForWorkflow(ctx, deps, workflows, boundary)
}

func FindInputReactorForWorkflow(ctx context.Context, deps *Dependencies, workflows Workflows, boundary Boundary) (*FindInputReactorResult, error) {
	if len(workflows.Nearest.Nodes) > 0 {
		// We check the last node if it can react to input first.
		lastNode := workflows.Nearest.Nodes[len(workflows.Nearest.Nodes)-1]
		findInputReactorResult, err := FindInputReactorForNode(ctx, deps, workflows, boundary, &lastNode)
		if err == nil {
			return findInputReactorResult, nil
		}
		// Return non ErrEOF error.
		if err != nil && !errors.Is(err, ErrEOF) {
			return nil, err
		}
		// err is ErrEOF, fallthrough
	}

	// Otherwise we check if the intent can react to input.
	inputSchema, err := workflows.Nearest.Intent.CanReactTo(ctx, deps, workflows)
	if err == nil {
		// Update boundary
		if b, ok := workflows.Nearest.Intent.(Boundary); ok {
			boundary = b
		}
		return &FindInputReactorResult{
			Workflows:    workflows,
			InputReactor: workflows.Nearest.Intent,
			InputSchema:  inputSchema,
			Boundary:     boundary,
		}, nil
	}

	// err != nil here.
	// Regardless of whether err is ErrEOF, we return err.
	return nil, err
}

func FindInputReactorForNode(ctx context.Context, deps *Dependencies, workflows Workflows, boundary Boundary, n *Node) (*FindInputReactorResult, error) {
	switch n.Type {
	case NodeTypeSimple:
		reactor, ok := n.Simple.(InputReactor)
		if !ok {
			return nil, ErrEOF
		}

		inputSchema, err := reactor.CanReactTo(ctx, deps, workflows)
		if err == nil {
			if b, ok := reactor.(Boundary); ok {
				boundary = b
			}
			return &FindInputReactorResult{
				Workflows:    workflows,
				InputReactor: reactor,
				InputSchema:  inputSchema,
				Boundary:     boundary,
			}, nil
		}
		return nil, err
	case NodeTypeSubWorkflow:
		return FindInputReactorForWorkflow(ctx, deps, workflows.Replace(n.SubWorkflow), boundary)
	default:
		panic(errors.New("unreachable"))
	}
}
