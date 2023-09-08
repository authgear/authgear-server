package authenticationflow

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

var ErrInvalidJSONPointer = errors.New("invalid json pointer")
var ErrNoEntries = errors.New("no entries")

type TraverseEntry struct {
	FlowObject  config.AuthenticationFlowObject
	JSONPointer jsonpointer.T
	ID          string
	FieldName   string
	Index       int
}

func FlowObjectGetID(o config.AuthenticationFlowObject) string {
	if root, ok := o.(config.AuthenticationFlowObjectFlowRoot); ok {
		return root.GetID()
	}
	if step, ok := o.(config.AuthenticationFlowObjectFlowStep); ok {
		return step.GetID()
	}
	return ""
}

func FlowObjectGetSteps(o config.AuthenticationFlowObject) ([]config.AuthenticationFlowObject, bool) {
	if root, ok := o.(config.AuthenticationFlowObjectFlowRoot); ok {
		return root.GetSteps(), true
	}
	if branch, ok := o.(config.AuthenticationFlowObjectFlowBranch); ok {
		return branch.GetSteps(), true
	}
	return nil, false
}

func FlowObjectGetOneOf(o config.AuthenticationFlowObject) ([]config.AuthenticationFlowObject, bool) {
	if step, ok := o.(config.AuthenticationFlowObjectFlowStep); ok {
		return step.GetOneOf(), true
	}
	return nil, false
}

func Traverse(o config.AuthenticationFlowObject, pointer jsonpointer.T) ([]TraverseEntry, error) {
	entries := []TraverseEntry{
		{
			FlowObject: o,
			ID:         FlowObjectGetID(o),
		},
	}
	var traversedPointer jsonpointer.T
	for {
		switch {
		case len(pointer) == 0:
			return entries, nil
		case len(pointer)%2 == 1:
			return nil, ErrInvalidJSONPointer
		default:
			fieldName := pointer[0]

			indexStr := pointer[1]
			index, err := strconv.Atoi(indexStr)
			if err != nil {
				return nil, fmt.Errorf("%v is not an index", indexStr)
			}

			traversedPointer = traversedPointer.AddReferenceToken(fieldName)
			traversedPointer = traversedPointer.AddReferenceToken(indexStr)
			pointer = pointer[2:]

			var objects []config.AuthenticationFlowObject
			var ok bool
			switch fieldName {
			case "steps":
				objects, ok = FlowObjectGetSteps(o)
			case "one_of":
				objects, ok = FlowObjectGetOneOf(o)
			default:
				break
			}
			if !ok {
				return nil, ErrInvalidJSONPointer
			}

			if index >= len(objects) {
				return nil, fmt.Errorf("index out of bound: %v >= %v", index, len(objects))
			}

			o = objects[index]
			entries = append(entries, TraverseEntry{
				FlowObject:  o,
				JSONPointer: traversedPointer,
				ID:          FlowObjectGetID(o),
				FieldName:   fieldName,
				Index:       index,
			})
		}
	}
}

func GetCurrentObject(entries []TraverseEntry) (config.AuthenticationFlowObject, error) {
	if len(entries) <= 0 {
		return nil, ErrNoEntries
	}
	return entries[len(entries)-1].FlowObject, nil
}

func JSONPointerForStep(p jsonpointer.T, index int) jsonpointer.T {
	return p.AddReferenceToken("steps").AddReferenceToken(strconv.Itoa(index))
}

func JSONPointerForOneOf(p jsonpointer.T, index int) jsonpointer.T {
	return p.AddReferenceToken("one_of").AddReferenceToken(strconv.Itoa(index))
}

func JSONPointerToParent(p jsonpointer.T) jsonpointer.T {
	if len(p) < 2 {
		panic(ErrInvalidJSONPointer)
	}
	return p[:len(p)-2]
}

func FlowObject(flowRootObject config.AuthenticationFlowObject, pointer jsonpointer.T) (config.AuthenticationFlowObject, error) {
	entries, err := Traverse(flowRootObject, pointer)
	if err != nil {
		return nil, err
	}

	current, err := GetCurrentObject(entries)
	if err != nil {
		return nil, err
	}

	return current, nil
}
