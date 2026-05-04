package template

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/util/copyutil"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

type ResolveErrorRawInput map[string]any // actually map[string][]string, but go template dict always map[string]interface{}
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

func toErrorJSON(apiErr *apierrors.APIError) any {
	if apiErr == nil {
		return nil
	}
	b, err := json.Marshal(struct {
		Error *apierrors.APIError `json:"error"`
	}{apiErr})
	if err != nil {
		return nil
	}
	var eJSON map[string]any
	err = json.Unmarshal(b, &eJSON)
	if err != nil {
		return nil
	}
	return eJSON["error"]
}

func ResolveError(apiErr *apierrors.APIError, input ResolveErrorRawInput) map[string]any {
	if apiErr == nil {
		return nil
	}

	resolveErrInput, err := newResolveErrorInput(input)
	if err != nil {
		// This panic is legitimate, because it only happens if the developer uses `resolveError` incorrectly
		panic(fmt.Errorf("invalid resolveError input: %w", err))
	}

	errMap := resolveError(apiErr, resolveErrInput)

	newErrMap := make(map[string]any)
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
	apiErrByReason := getErrorByReason(apiErr, getErrInput.ByReason)
	if apiErrByReason != nil {
		return apiErrByReason
	}
	switch apiErr.Reason {
	case "ValidationFailed":
		apiErrByLocation := getValidationErrorByLocation(apiErr, getErrInput.ByLocation)
		if apiErrByLocation != nil {
			return apiErrByLocation
		}
	}

	return nil
}

func getErrorByReason(apiErr *apierrors.APIError, reasons []string) *apierrors.APIError {
	if slices.Contains(reasons, apiErr.Reason) {
		return apiErr
	}
	return nil
}

// getValidationErrorByLocation returns an error with trimmed causes
//
// If all causes are trimmed, it returns nil
// nolint: gocognit
func getValidationErrorByLocation(apiErr *apierrors.APIError, locations []string) *apierrors.APIError {
	out := apiErr.Clone()
	causes, ok := out.Info_ReadOnly["causes"].([]any)
	if !ok {
		return nil
	}
	typedCauses := slice.Map(causes, func(cause any) map[string]any {
		_cause, ok := cause.(map[string]any)
		if !ok {
			return nil
		}
		return _cause
	})

	typedCausesWithoutNil := slice.Filter(typedCauses, func(cause map[string]any) bool {
		return cause != nil
	})

	trimmedCauses := trimValidationErrorCauses(typedCausesWithoutNil, locations)
	if len(trimmedCauses) == 0 {
		return nil
	}
	out.Info_ReadOnly["causes"] = trimmedCauses
	return out
}

// trimValidationErrorCauses removes causes that
//
//   - have mismatching location; AND
//   - have details with ALL missing and expected fields that does not match required fields
//     -- i.e. partially-matched cause will not be removed
func trimValidationErrorCauses(causes []map[string]any, locations []string) []map[string]any {
	trimmedCauses := slice.Map(causes, func(cause map[string]any) map[string]any {
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

	trimmedCausesWithoutNil := slice.Filter(trimmedCauses, func(cause map[string]any) bool {
		return cause != nil
	})

	return trimmedCausesWithoutNil
}

// trimCauseWithLocation returns matching location cause, otherwise nil
func trimCauseWithLocation(cause map[string]any, locations []string) map[string]any {
	for _, loc := range locations {
		if cause["location"] == loc {
			return cause
		}
	}
	return nil
}

// trimCauseWithoutLocation removes causes with missing and expected fields that does not match required fields
func trimCauseWithoutLocation(cause map[string]any, requiredFields []string) map[string]any {
	kind, ok := cause["kind"].(string)
	if !ok {
		return nil
	}
	if kind != "required" {
		return nil
	}
	newCause := make(map[string]any)
	for k, v := range cause {
		switch k {
		case "details":
			newDetails := trimCauseDetails(v.(map[string]any), requiredFields)
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
func trimCauseDetails(details map[string]any, requiredFields []string) map[string]any {
	newDetails := make(map[string]any)
	for k, v := range details {
		switch k {
		case "missing":
			missings := details["missing"].([]any)
			newDetails["missing"] = slice.Filter(missings, func(missingItem any) bool {
				return slice.ContainsString(requiredFields, missingItem.(string))
			})
		case "expected":
			expecteds := details["expected"].([]any)
			newDetails["expected"] = slice.Filter(expecteds, func(expectedItem any) bool {
				return slice.ContainsString(requiredFields, expectedItem.(string))
			})
		default:
			newDetails[k] = v
		}
	}
	return newDetails
}

// isRequiredDetailsEmpty returns true if missing and expected fields are empty
func isRequiredDetailsEmpty(details map[string]any) bool {
	return len(details["missing"].([]any)) == 0 || len(details["expected"].([]any)) == 0
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
