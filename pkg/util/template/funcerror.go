package template

import (
	"encoding/json"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/util/copyutil"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

type ResolveErrorRawInput map[string]interface{} // actually map[string][]string, but go template dict always map[string]interface{}
type ResolveErrorInput map[string]*GetErrorInput

func newResolveErrorInput(input ResolveErrorRawInput) (ResolveErrorInput, error) {
	// convert map to json
	bytes, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to convert map to json: %w", err)
	}

	// convert json to struct
	var resolveErrInput ResolveErrorInput
	err = json.Unmarshal(bytes, &resolveErrInput)
	if err != nil {
		return nil, fmt.Errorf("invalid json input: %w", err)
	}

	return resolveErrInput, nil
}

func toErrorJSON(apiErr *apierrors.APIError) interface{} {
	if apiErr == nil {
		return nil
	}
	b, err := json.Marshal(struct {
		Error *apierrors.APIError `json:"error"`
	}{apiErr})
	if err != nil {
		return nil
	}
	var eJSON map[string]interface{}
	err = json.Unmarshal(b, &eJSON)
	if err != nil {
		return nil
	}
	return eJSON["error"]
}

func ResolveError(apiErr *apierrors.APIError, input ResolveErrorRawInput) map[string]interface{} {
	if apiErr == nil {
		return nil
	}

	resolveErrInput, err := newResolveErrorInput(input)
	if err != nil {
		// This panic is legitimate, because it only happens if the developer uses `resolveError` incorrectly
		panic(fmt.Errorf("invalid resolveError input: %w", err))
	}

	errMap := resolveError(apiErr, resolveErrInput)

	newErrMap := make(map[string]interface{})
	for errK, errV := range errMap {
		newErrMap[errK] = toErrorJSON(errV)
	}
	return newErrMap
}

func resolveError(apiErr *apierrors.APIError, input ResolveErrorInput) map[string]*apierrors.APIError {
	if apiErr == nil {
		return nil
	}

	output := make(map[string]*apierrors.APIError)

	for errName, getErrInput := range input {
		resultErr := getError(apiErr, getErrInput)
		output[errName] = resultErr
	}

	outputWithUnknownErr := addUnknownError(apiErr, output)

	return outputWithUnknownErr
}

type GetErrorInput struct {
	ByReason   []string `json:"by_reason,omitempty"`
	ByLocation []string `json:"by_location,omitempty"` // this field actually check both location and required field
}

func getError(apiErr *apierrors.APIError, getErrInput *GetErrorInput) *apierrors.APIError {
	if apiErr == nil {
		return nil
	}
	switch apiErr.Reason {
	case "ValidationFailed":
		apiErrByLocation := getValidationErrorByLocation(apiErr, getErrInput.ByLocation)
		if apiErrByLocation != nil {
			return apiErrByLocation
		}
	default:
		apiErrByReason := getErrorByReason(apiErr, getErrInput.ByReason)
		if apiErrByReason != nil {
			return apiErrByReason
		}
	}

	return nil
}

func getErrorByReason(apiErr *apierrors.APIError, reasons []string) *apierrors.APIError {
	for _, reason := range reasons {
		if apiErr.Reason == reason {
			return apiErr
		}
	}
	return nil
}

// getValidationErrorByLocation returns an error with trimmed causes
//
// If all causes are trimmed, it returns nil
// nolint: gocognit
func getValidationErrorByLocation(apiErr *apierrors.APIError, locations []string) *apierrors.APIError {
	out := apiErr.Clone()
	causes, ok := out.Info["causes"].([]interface{})
	if !ok {
		return nil
	}
	typedCauses := slice.Map(causes, func(cause interface{}) map[string]interface{} {
		_cause, ok := cause.(map[string]interface{})
		if !ok {
			return nil
		}
		return _cause
	})

	typedCausesWithoutNil := slice.Filter(typedCauses, func(cause map[string]interface{}) bool {
		return cause != nil
	})

	trimmedCauses := trimValidationErrorCauses(typedCausesWithoutNil, locations)
	if len(trimmedCauses) == 0 {
		return nil
	}
	out.Info["causes"] = trimmedCauses
	return out
}

// trimValidationErrorCauses removes causes that
//
//   - have mismatching location; AND
//   - have details with ALL missing and expected fields that does not match required fields
//     -- i.e. partially-matched cause will not be removed
func trimValidationErrorCauses(causes []map[string]interface{}, locations []string) []map[string]interface{} {
	trimmedCauses := slice.Map(causes, func(cause map[string]interface{}) map[string]interface{} {
		causeLoc, ok := cause["location"].(string)
		if !ok {
			return nil
		}
		switch causeLoc {
		case "":
			return trimCauseWithoutLocation(cause, locations)
		default:
			return trimCauseWithLocation(cause, locations)
		}
	})

	trimmedCausesWithoutNil := slice.Filter(trimmedCauses, func(cause map[string]interface{}) bool {
		return cause != nil
	})

	return trimmedCausesWithoutNil
}

// trimCauseWithLocation returns matching location cause, otherwise nil
func trimCauseWithLocation(cause map[string]interface{}, locations []string) map[string]interface{} {
	for _, loc := range locations {
		if cause["location"] == loc {
			return cause
		}
	}
	return nil
}

// trimCauseWithoutLocation removes causes with missing and expected fields that does not match required fields
func trimCauseWithoutLocation(cause map[string]interface{}, requiredFields []string) map[string]interface{} {
	kind, ok := cause["kind"].(string)
	if !ok {
		return nil
	}
	if kind != "required" {
		return nil
	}
	newCause := make(map[string]interface{})
	for k, v := range cause {
		switch k {
		case "details":
			newDetails := trimCauseDetails(v.(map[string]interface{}), requiredFields)
			if isRequiredDetailsEmpty(newDetails) {
				return nil
			}

			newCause["details"] = newDetails
		default:
			newCause[k] = v
		}

	}
	return newCause
}

// trimCauseDetails removes missing and expected fields that does not match required fields
func trimCauseDetails(details map[string]interface{}, requiredFields []string) map[string]interface{} {
	newDetails := make(map[string]interface{})
	for k, v := range details {
		switch k {
		case "missing":
			missings := details["missing"].([]interface{})
			newDetails["missing"] = slice.Filter(missings, func(missingItem interface{}) bool {
				return slice.ContainsString(requiredFields, missingItem.(string))
			})
		case "expected":
			expecteds := details["expected"].([]interface{})
			newDetails["expected"] = slice.Filter(expecteds, func(expectedItem interface{}) bool {
				return slice.ContainsString(requiredFields, expectedItem.(string))
			})
		default:
			newDetails[k] = v
		}
	}
	return newDetails
}

// isRequiredDetailsEmpty returns true if missing and expected fields are empty
func isRequiredDetailsEmpty(details map[string]interface{}) bool {
	return len(details["missing"].([]interface{})) == 0 || len(details["expected"].([]interface{})) == 0
}

func addUnknownError(apiErr *apierrors.APIError, currOutput map[string]*apierrors.APIError) map[string]*apierrors.APIError {
	outputClone, err := copyutil.Clone(currOutput)
	if err != nil {
		return currOutput
	}

	output := outputClone.(map[string]*apierrors.APIError)

	hasNonNilErr := false
	for _, err := range output {
		if err != nil {
			hasNonNilErr = true
			break
		}
	}

	output["unknown"] = nil
	if !hasNonNilErr {
		output["unknown"] = apiErr
	}
	return output
}
