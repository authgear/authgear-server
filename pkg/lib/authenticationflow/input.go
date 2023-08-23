package authenticationflow

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
	CanReactTo(ctx context.Context, deps *Dependencies, flows Flows) (InputSchema, error)
	ReactTo(ctx context.Context, deps *Dependencies, flows Flows, input Input) (*Node, error)
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
	Flows        Flows
	InputReactor InputReactor
	InputSchema  InputSchema
	Boundary     Boundary
}

func FindInputReactor(ctx context.Context, deps *Dependencies, flows Flows) (*FindInputReactorResult, error) {
	var boundary Boundary
	return FindInputReactorForFlow(ctx, deps, flows, boundary)
}

func FindInputReactorForFlow(ctx context.Context, deps *Dependencies, flows Flows, boundary Boundary) (*FindInputReactorResult, error) {
	if len(flows.Nearest.Nodes) > 0 {
		// We check the last node if it can react to input first.
		lastNode := flows.Nearest.Nodes[len(flows.Nearest.Nodes)-1]
		findInputReactorResult, err := FindInputReactorForNode(ctx, deps, flows, boundary, &lastNode)
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
	inputSchema, err := flows.Nearest.Intent.CanReactTo(ctx, deps, flows)
	if err == nil {
		// Update boundary
		if b, ok := flows.Nearest.Intent.(Boundary); ok {
			boundary = b
		}
		return &FindInputReactorResult{
			Flows:        flows,
			InputReactor: flows.Nearest.Intent,
			InputSchema:  inputSchema,
			Boundary:     boundary,
		}, nil
	}

	// err != nil here.
	// Regardless of whether err is ErrEOF, we return err.
	return nil, err
}

func FindInputReactorForNode(ctx context.Context, deps *Dependencies, flows Flows, boundary Boundary, n *Node) (*FindInputReactorResult, error) {
	switch n.Type {
	case NodeTypeSimple:
		reactor, ok := n.Simple.(InputReactor)
		if !ok {
			return nil, ErrEOF
		}

		inputSchema, err := reactor.CanReactTo(ctx, deps, flows)
		if err == nil {
			if b, ok := reactor.(Boundary); ok {
				boundary = b
			}
			return &FindInputReactorResult{
				Flows:        flows,
				InputReactor: reactor,
				InputSchema:  inputSchema,
				Boundary:     boundary,
			}, nil
		}
		return nil, err
	case NodeTypeSubFlow:
		return FindInputReactorForFlow(ctx, deps, flows.Replace(n.SubFlow), boundary)
	default:
		panic(errors.New("unreachable"))
	}
}
