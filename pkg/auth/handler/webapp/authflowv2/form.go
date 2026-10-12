package authflowv2

import (
	htmltemplate "html/template"
	"maps"
	"math"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/territoryutil"
	"github.com/authgear/authgear-server/pkg/util/tzutil"
)

var addressSubKeys = []string{"formatted", "street_address", "locality", "region", "postal_code", "country"}

// FormFieldsInput is the `fields` input built from the posted HTML form,
// plus the key of each item so that error locations /fields/{i}/value map back to fields.
type FormFieldsInput struct {
	Fields []map[string]any
	Keys   []string
}

// BuildFormFieldsInput reads the fields of data from form.
// HTML input names are the field keys, which are JSON pointers.
// An address is read from the inputs named after its sub-attributes, such as /address/locality.
func BuildFormFieldsInput(form url.Values, data declarative.FillFormDataForm) FormFieldsInput {
	// Fields must not be nil, which marshals to null and fails the schema of an all-optional form.
	out := FormFieldsInput{Fields: []map[string]any{}}
	for _, f := range data.Fields {
		value, ok := readFormFieldValue(form, f)
		if !ok {
			continue
		}
		out.Fields = append(out.Fields, map[string]any{"key": f.Key, "value": value})
		out.Keys = append(out.Keys, f.Key)
	}
	return out
}

func readFormFieldValue(form url.Values, f declarative.FillFormDataField) (value any, ok bool) {
	switch f.Type {
	case declarative.FillFormFieldTypeBoolean:
		// A checkbox comes before the hidden false, so the first value wins.
		return form.Get(f.Key) == "true", true
	case declarative.FillFormFieldTypeAddress:
		address := map[string]any{}
		for _, sub := range addressSubKeys {
			if v := form.Get(f.Key + "/" + sub); v != "" {
				address[sub] = v
			}
		}
		if len(address) == 0 {
			return nil, false
		}
		return address, true
	}

	raw := form.Get(f.Key)
	if raw == "" {
		return nil, false
	}
	switch f.Type {
	case declarative.FillFormFieldTypeInteger:
		if i, err := strconv.ParseInt(raw, 10, 64); err == nil {
			return i, true
		}
	case declarative.FillFormFieldTypeNumber:
		// ParseFloat accepts NaN and Inf, which JSON cannot carry.
		if n, err := strconv.ParseFloat(raw, 64); err == nil && !math.IsNaN(n) && !math.IsInf(n, 0) {
			return n, true
		}
	}
	// On a parse error the raw string is sent, so the server returns the type error.
	return raw, true
}

// FormErrorLocations maps each submitted field key to the cause locations of that field.
func FormErrorLocations(data declarative.FillFormDataForm, keys []string) map[string][]string {
	out := map[string][]string{}
	for i, key := range keys {
		base := "/fields/" + strconv.Itoa(i) + "/value"
		locations := []string{base}
		if key == "/address" {
			for _, sub := range addressSubKeys {
				locations = append(locations, base+"/"+sub)
			}
		}
		out[key] = locations
	}
	return out
}

// HookFieldMessages extracts invalid_form_field messages from a HookDisallowed error.
// unmatched holds the messages whose key is not in keys.
// info.reasons arrives as []any of map[string]any, because the error went through JSON.
func HookFieldMessages(apiErr *apierrors.APIError, keys []string) (byKey map[string]string, unmatched []string) {
	byKey = map[string]string{}
	if apiErr == nil || apiErr.Reason != "HookDisallowed" {
		return
	}
	items, _ := apiErr.Info_ReadOnly["reasons"].([]any)
	for _, item := range items {
		itemMap, _ := item.(map[string]any)
		reasons, _ := itemMap["reasons"].([]any)
		for _, r := range reasons {
			reason, _ := r.(map[string]any)
			if reason["type"] != "invalid_form_field" {
				continue
			}
			key, _ := reason["key"].(string)
			message, _ := reason["message"].(string)
			if message == "" {
				continue
			}
			if contains(keys, key) {
				if _, exists := byKey[key]; !exists {
					byKey[key] = message
					continue
				}
			}
			unmatched = append(unmatched, message)
		}
	}
	return
}

func contains(l []string, s string) bool {
	return slices.Contains(l, s)
}

type FormFieldViewModel struct {
	Key   string
	Type  string
	Label string
	// LabelHTML is for boolean only. The Authflow API has sanitized it.
	LabelHTML htmltemplate.HTML
	Required  bool
	// Value is the restored form value if present, else the current value of the field.
	Value any
	// Checked is for boolean.
	Checked bool
	// Address is for address. It is never nil there.
	Address     map[string]string
	Options     []declarative.FillFormDataFieldOption
	Suggestions []string
	Constraints *declarative.FillFormDataFieldConstraints

	// Error holds the causes of this field only, as JSON. It is nil when none.
	Error any
	// MissingRequired is true when the field is required, was not submitted, and the input failed.
	MissingRequired bool
	HookMessage     string
}

type FormViewModel struct {
	Name   string
	Fields []FormFieldViewModel
	// UnknownError is the error not attributed to a field, as JSON.
	UnknownError          any
	HookUnmatchedMessages []string
	// Timezones and Alpha2 are read by the input partials from the root context.
	Timezones []tzutil.Timezone
	Alpha2    []string
}

func NewFormViewModel(form declarative.FillFormDataForm, base *viewmodels.BaseViewModel) (FormViewModel, error) {
	input := BuildFormFieldsInput(base.RestoredForm, form)
	restored := base.RestoredForm

	vm := FormViewModel{
		Name:   form.Name,
		Fields: make([]FormFieldViewModel, 0, len(form.Fields)),
		Alpha2: territoryutil.Alpha2,
	}
	timezones, err := tzutil.List(time.Unix(int64(base.PageLoadedAt), 0).UTC())
	if err != nil {
		return FormViewModel{}, err
	}
	vm.Timezones = timezones

	rawErr := normalizeMissingRequiredCauses(base.RawError)

	var fieldErrors map[string]any
	if rawErr != nil {
		locations := FormErrorLocations(form, input.Keys)
		resolveInput := template.ResolveErrorRawInput{}
		for key, locs := range locations {
			resolveInput[key] = map[string]any{"by_location": locs}
		}
		fieldErrors = template.ResolveError(rawErr, resolveInput)
		vm.UnknownError = withoutFieldsCause(fieldErrors["unknown"])
	}

	allKeys := make([]string, 0, len(form.Fields))
	for _, f := range form.Fields {
		allKeys = append(allKeys, f.Key)
	}
	hookByKey, hookUnmatched := HookFieldMessages(rawErr, allKeys)
	vm.HookUnmatchedMessages = hookUnmatched
	hasMissingRequiredCause := hasFieldsCause(rawErr)

	for _, f := range form.Fields {
		fvm := FormFieldViewModel{
			Key:         f.Key,
			Type:        string(f.Type),
			Label:       f.Label,
			Required:    f.Required,
			Value:       f.Value,
			Options:     f.Options,
			Suggestions: f.Suggestions,
			Constraints: f.Constraints,
			Error:       fieldErrors[f.Key],
			HookMessage: hookByKey[f.Key],
		}
		fvm.MissingRequired = hasMissingRequiredCause && f.Required && !contains(input.Keys, f.Key)
		if restored != nil {
			switch f.Type {
			case declarative.FillFormFieldTypeBoolean:
				fvm.Value = restored.Get(f.Key) == "true"
			case declarative.FillFormFieldTypeAddress:
				// Handled below.
			default:
				if restored.Has(f.Key) {
					fvm.Value = restored.Get(f.Key)
				}
			}
		}

		switch f.Type {
		case declarative.FillFormFieldTypeBoolean:
			fvm.Checked, _ = fvm.Value.(bool)
			// An unchecked checkbox submits false, which is missing as a value, or as a required field.
			if f.Required && !fvm.Checked && (hasMissingRequiredCause || fvm.Error != nil) {
				fvm.MissingRequired = true
			}
			fvm.LabelHTML = htmltemplate.HTML(f.Label) // nolint:gosec
		case declarative.FillFormFieldTypeAddress:
			fvm.Address = map[string]string{}
			current, _ := f.Value.(map[string]any)
			for _, sub := range addressSubKeys {
				name := f.Key + "/" + sub
				if restored != nil && restored.Has(name) {
					fvm.Address[sub] = restored.Get(name)
				} else if s, ok := current[sub].(string); ok {
					fvm.Address[sub] = s
				}
			}
		}

		vm.Fields = append(vm.Fields, fvm)
	}

	return vm, nil
}

// normalizeMissingRequiredCauses folds the causes of a missing required field into one cause at /fields.
// The validator reports it at /fields when no field is submitted. When other fields are submitted,
// it reports the key and the value of each of them, as it checks each against the required field.
// A submitted field with a key cause has none of its own value errors reported, because of that.
func normalizeMissingRequiredCauses(apiErr *apierrors.APIError) *apierrors.APIError {
	if apiErr == nil || apiErr.Reason != "ValidationFailed" {
		return apiErr
	}
	causes, _ := apiErr.Info_ReadOnly["causes"].([]any)

	checkedAgainstRequired := map[string]struct{}{}
	for _, c := range causes {
		cause, _ := c.(map[string]any)
		location, _ := cause["location"].(string)
		if index, rest, ok := splitFieldLocation(location); ok && rest == "key" {
			checkedAgainstRequired[index] = struct{}{}
		}
	}
	if len(checkedAgainstRequired) == 0 {
		return apiErr
	}

	var kept []any
	for _, c := range causes {
		cause, _ := c.(map[string]any)
		location, _ := cause["location"].(string)
		if index, _, ok := splitFieldLocation(location); ok {
			if _, noise := checkedAgainstRequired[index]; noise {
				continue
			}
		}
		kept = append(kept, c)
	}
	kept = append(kept, map[string]any{"location": "/fields", "kind": "contains"})
	info := apierrors.Details{}
	maps.Copy(info, apiErr.Info_ReadOnly)
	info["causes"] = kept
	return apiErr.CloneWithInfo(info)
}

// splitFieldLocation splits /fields/<index>/<rest> into the index and rest.
func splitFieldLocation(location string) (index string, rest string, ok bool) {
	after, found := strings.CutPrefix(location, "/fields/")
	if !found {
		return "", "", false
	}
	index, rest, _ = strings.Cut(after, "/")
	if _, err := strconv.Atoi(index); err != nil {
		return "", "", false
	}
	return index, rest, true
}

// hasFieldsCause reports whether the error has a ValidationFailed cause at /fields,
// which is where a missing required field is reported.
func hasFieldsCause(apiErr *apierrors.APIError) bool {
	if apiErr == nil || apiErr.Reason != "ValidationFailed" {
		return false
	}
	causes, _ := apiErr.Info_ReadOnly["causes"].([]any)
	for _, c := range causes {
		cause, _ := c.(map[string]any)
		if cause["location"] == "/fields" {
			return true
		}
	}
	return false
}

// withoutFieldsCause removes the causes at /fields from a ValidationFailed error,
// because they are shown under the required fields.
// It returns nil when no cause is left.
func withoutFieldsCause(unknown any) any {
	errMap, ok := unknown.(map[string]any)
	if !ok || errMap["reason"] != "ValidationFailed" {
		return unknown
	}
	info, _ := errMap["info"].(map[string]any)
	causes, _ := info["causes"].([]any)
	var rest []any
	for _, c := range causes {
		cause, _ := c.(map[string]any)
		if cause["location"] == "/fields" {
			continue
		}
		rest = append(rest, c)
	}
	if len(rest) == 0 {
		return nil
	}
	info["causes"] = rest
	return errMap
}

// IdentificationOptionForm returns the form of the option of the identification, or nil.
func IdentificationOptionForm(screen *webapp.AuthflowScreenWithFlowResponse, identification string) *declarative.FillFormDataForm {
	data, ok := screen.StateTokenFlowResponse.Action.Data.(declarative.IdentificationData)
	if !ok {
		return nil
	}
	for _, option := range data.Options {
		if string(option.Identification) == identification && option.Form != nil {
			return option.Form
		}
	}
	return nil
}

// isFormError reports whether an error of the identify page belongs to its form, not to the login ID.
func isFormError(apiErr *apierrors.APIError) bool {
	if apiErr == nil {
		return false
	}
	switch apiErr.Reason {
	case "HookDisallowed":
		return apiErr.Info_ReadOnly["event_type"] == "authentication.form.post_submitted"
	case "ValidationFailed":
		causes, _ := apiErr.Info_ReadOnly["causes"].([]any)
		if len(causes) == 0 {
			return false
		}
		for _, c := range causes {
			cause, _ := c.(map[string]any)
			location, _ := cause["location"].(string)
			if location != "/fields" && !strings.HasPrefix(location, "/fields/") {
				return false
			}
		}
		return true
	}
	return false
}

// IdentifyFormViewModels builds the forms of the options of the identify step, keyed by identification.
func IdentifyFormViewModels(screen *webapp.AuthflowScreenWithFlowResponse, base *viewmodels.BaseViewModel) (map[string]FormViewModel, error) {
	out := map[string]FormViewModel{}
	data, ok := screen.StateTokenFlowResponse.Action.Data.(declarative.IdentificationData)
	if !ok {
		return out, nil
	}
	for _, option := range data.Options {
		if option.Form == nil {
			continue
		}
		vm, err := NewFormViewModel(*option.Form, base)
		if err != nil {
			return nil, err
		}
		out[string(option.Identification)] = vm
	}
	return out, nil
}
