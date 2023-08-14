package workflowconfig

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

var ErrInvalidJSONPointer = errors.New("invalid json pointer")
var ErrNoEntries = errors.New("no entries")
var ErrEOF = errors.New("eof")

type TraverseEntry struct {
	WorkflowObject config.WorkflowObject
	JSONPointer    jsonpointer.T
	ID             string
	FieldName      string
	Index          int
}

func getID(o config.WorkflowObject) string {
	id, _ := o.GetID()
	return id
}

func Traverse(o config.WorkflowObject, pointer jsonpointer.T) ([]TraverseEntry, error) {
	entries := []TraverseEntry{
		{
			WorkflowObject: o,
			ID:             getID(o),
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

			var objects []config.WorkflowObject
			var ok bool
			switch fieldName {
			case "steps":
				objects, ok = o.GetSteps()
			case "one_of":
				objects, ok = o.GetOneOf()
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
				WorkflowObject: o,
				JSONPointer:    traversedPointer,
				ID:             getID(o),
				FieldName:      fieldName,
				Index:          index,
			})
		}
	}
}

func GetCurrentObject(entries []TraverseEntry) (config.WorkflowObject, error) {
	if len(entries) <= 0 {
		return nil, ErrNoEntries
	}
	return entries[len(entries)-1].WorkflowObject, nil
}

func JSONPointerForStep(p jsonpointer.T, index int) jsonpointer.T {
	return p.AddReferenceToken("steps").AddReferenceToken(strconv.Itoa(index))
}

func JSONPointerForOneOf(p jsonpointer.T, index int) jsonpointer.T {
	return p.AddReferenceToken("one_of").AddReferenceToken(strconv.Itoa(index))
}
