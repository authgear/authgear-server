package validation

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"

	"github.com/xeipuuv/gojsonschema"
)

type ErrorCause struct {
	Kind    ErrorCauseKind         `json:"kind"`
	Message string                 `json:"message"`
	Pointer string                 `json:"pointer"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (c ErrorCause) String() string {
	if len(c.Details) == 0 {
		return fmt.Sprintf("%s: %s", c.Pointer, c.Kind)
	}
	return fmt.Sprintf("%s: %s %v", c.Pointer, c.Kind, c.Details)
}

type ErrorCauseKind string

const (
	ErrorGeneral      ErrorCauseKind = "General"
	ErrorRequired     ErrorCauseKind = "Required"
	ErrorType         ErrorCauseKind = "Type"
	ErrorConstant     ErrorCauseKind = "Constant"
	ErrorEnum         ErrorCauseKind = "Enum"
	ErrorExtraEntry   ErrorCauseKind = "ExtraEntry"
	ErrorEntryAmount  ErrorCauseKind = "EntryAmount"
	ErrorStringLength ErrorCauseKind = "StringLength"
	ErrorStringFormat ErrorCauseKind = "StringFormat"
	ErrorNumberRange  ErrorCauseKind = "NumberRange"
)

func newCause(err gojsonschema.ResultError) *ErrorCause {
	var kind ErrorCauseKind
	var details map[string]interface{}
	message := err.Description()
	ctx := err.Context()
	d := err.Details()

	switch err.Type() {
	case "number_any_of", "number_one_of", "number_all_of", "number_not":
		// ignore combined schema error
		return nil
	case "condition_then", "condition_else":
		// ignore conditional schema error
		return nil
	case "invalid_property_name":
		// ignore invalid property error: should be handled by toCauses
		return nil

	case "required":
		kind = ErrorRequired
		ctx = gojsonschema.NewJsonContext(d["property"].(string), ctx)

	case "invalid_type":
		kind = ErrorType
		details = map[string]interface{}{
			"expected": d["expected"],
		}

	case "const":
		kind = ErrorConstant
		var c interface{}
		if err := json.Unmarshal([]byte(d["allowed"].(string)), &c); err != nil {
			panic(err)
		}
		details = map[string]interface{}{
			"expected": c,
		}

	case "enum":
		kind = ErrorEnum
		enumJSON := "[" + d["allowed"].(string) + "]"
		var enum []interface{}
		if err := json.Unmarshal([]byte(enumJSON), &enum); err != nil {
			panic(err)
		}
		details = map[string]interface{}{
			"expected": enum,
		}

	case "array_no_additional_items", "additional_property_not_allowed":
		kind = ErrorExtraEntry
		ctx = gojsonschema.NewJsonContext(d["property"].(string), ctx)

	case "array_min_items", "array_min_properties":
		kind = ErrorEntryAmount
		details = map[string]interface{}{"gte": d["min"]}

	case "array_max_items", "array_max_properties":
		kind = ErrorEntryAmount
		details = map[string]interface{}{"lte": d["max"]}

	case "pattern":
		kind = ErrorStringFormat
		details = map[string]interface{}{"pattern": d["pattern"].(*regexp.Regexp).String()}

	case "format":
		kind = ErrorStringFormat
		details = map[string]interface{}{"format": d["format"]}

	case "string_gte":
		kind = ErrorStringLength
		details = map[string]interface{}{"gte": d["min"]}

	case "string_lte":
		kind = ErrorStringLength
		details = map[string]interface{}{"lte": d["max"]}

	case "number_gte":
		kind = ErrorNumberRange
		details = map[string]interface{}{"gte": d["min"]}

	case "number_gt":
		kind = ErrorNumberRange
		details = map[string]interface{}{"gt": d["min"]}

	case "number_lte":
		kind = ErrorNumberRange
		details = map[string]interface{}{"lte": d["max"]}

	case "number_lt":
		kind = ErrorNumberRange
		details = map[string]interface{}{"lt": d["max"]}

	default:
		kind = ErrorGeneral
	}

	return &ErrorCause{
		Kind:    kind,
		Message: message,
		Pointer: ctx.JSONPointer(),
		Details: details,
	}
}

func toCauses(errs []gojsonschema.ResultError) []ErrorCause {
	var causes []ErrorCause

	var propertyCtx *gojsonschema.JsonContext
	var correctPropertyCtx *gojsonschema.JsonContext
	for _, err := range errs {
		if err.Type() == "invalid_property_name" {
			// propertyNames validation error would have pointer to parent object
			// i.e. [(PropertyName, "/x/c"), ("Enum", "x")]
			// special case it to translate it to correct pointer
			// i.e. [(PropertyName, "/x/c"), ("Enum", "x/c")]
			propertyCtx = err.Context()
			field := err.Details()["property"].(string)
			correctPropertyCtx = gojsonschema.NewJsonContext(field, propertyCtx)
		} else if err.Context() == propertyCtx {
			err.SetContext(correctPropertyCtx)
		} else {
			propertyCtx = nil
			correctPropertyCtx = nil
		}

		c := newCause(err)
		if c != nil {
			causes = append(causes, *c)
		}
	}
	sort.Slice(causes, func(i, j int) bool {
		if causes[i].Pointer == causes[j].Pointer {
			return causes[i].Kind < causes[j].Kind
		}
		return causes[i].Pointer < causes[j].Pointer
	})
	return causes
}
