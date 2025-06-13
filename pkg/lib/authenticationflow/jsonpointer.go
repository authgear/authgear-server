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

const (
	JsonPointerTokenOneOf string = "one_of"
	JsonPointerTokenSteps string = "steps"
)

type TraverseEntry struct {
	FlowObject  config.AuthenticationFlowObject
	JSONPointer jsonpointer.T
	Name        string
	FieldName   string
	Index       int
}

func FlowObjectGetName(o config.AuthenticationFlowObject) string {
	if root, ok := o.(config.AuthenticationFlowObjectFlowRoot); ok {
		return root.GetName()
	}
	if step, ok := o.(config.AuthenticationFlowObjectFlowStep); ok {
		return step.GetName()
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
			Name:       FlowObjectGetName(o),
		},
	}
	var traversedPointer jsonpointer.T
	for {
		switch {
		case len(pointer) == 0:
			return entries, nil
		case len(pointer)%2 == 1:
			// Programming error, panic so we have stack trace
			panic(ErrInvalidJSONPointer)
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
				// Programming error, panic so we have stack trace
				panic(ErrInvalidJSONPointer)
			}

			if index >= len(objects) {
				return nil, fmt.Errorf("index out of bound: %v >= %v", index, len(objects))
			}

			o = objects[index]
			entries = append(entries, TraverseEntry{
				FlowObject:  o,
				JSONPointer: traversedPointer,
				Name:        FlowObjectGetName(o),
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
	return p.AddReferenceToken(JsonPointerTokenSteps).AddReferenceToken(strconv.Itoa(index))
}

func JSONPointerForOneOf(p jsonpointer.T, index int) jsonpointer.T {
	return p.AddReferenceToken(JsonPointerTokenOneOf).AddReferenceToken(strconv.Itoa(index))
}

func JSONPointerToParent(p jsonpointer.T) jsonpointer.T {
	if len(p) < 2 {
		panic(ErrInvalidJSONPointer)
	}
	return p[:len(p)-2]
}

func JSONPointerSubtract(p1 jsonpointer.T, p2 jsonpointer.T) jsonpointer.T {
	result := jsonpointer.T{}
	isDifferent := false
	for idx, el := range p1 {
		// Compare the element at same position, until found a different
		if (len(p2)-1) < idx || el != p2[idx] {
			isDifferent = true
		}
		if isDifferent {
			result = append(result, el)
		}
	}
	return result
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

func GetFlowAction(flowRootObject config.AuthenticationFlowObject, pointer jsonpointer.T) *FlowAction {
	flowObject, err := FlowObject(flowRootObject, pointer)
	if err != nil {
		panic(err)
	}

	switch o := flowObject.(type) {
	case config.AuthenticationFlowObjectFlowRoot:
		return nil
	case config.AuthenticationFlowObjectFlowStep:
		return &FlowAction{
			Type: FlowActionTypeFromStepType(o.GetType()),
		}
	case config.AuthenticationFlowObjectFlowBranch:
		branchInfo := o.GetBranchInfo()
		step := &FlowAction{
			Identification: branchInfo.Identification,
			Authentication: branchInfo.Authentication,
		}

		stepPointer := JSONPointerToParent(pointer)
		stepFlowObjectBeforeCast, err := FlowObject(flowRootObject, stepPointer)
		if err != nil {
			panic(err)
		}

		stepFlowObject := stepFlowObjectBeforeCast.(config.AuthenticationFlowObjectFlowStep)
		step.Type = FlowActionTypeFromStepType(stepFlowObject.GetType())

		return step
	default:
		panic(fmt.Errorf("unexpected flow object: %T", flowObject))
	}
}
