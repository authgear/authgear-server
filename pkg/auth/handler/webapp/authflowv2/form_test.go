package authflowv2

import (
	"encoding/json"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
)

func formData(fields ...declarative.FillFormDataField) declarative.FillFormDataForm {
	return declarative.FillFormDataForm{Fields: fields}
}

func f(key string, typ declarative.FillFormFieldType, required bool) declarative.FillFormDataField {
	return declarative.FillFormDataField{Key: key, Type: typ, Required: required}
}

// roundTrip makes an error the way it reaches a page, which is through JSON in Redis.
func roundTrip(err *apierrors.APIError) *apierrors.APIError {
	b, e := json.Marshal(err)
	So(e, ShouldBeNil)
	var out apierrors.APIError
	So(json.Unmarshal(b, &out), ShouldBeNil)
	return &out
}

func TestBuildFormFieldsInput(t *testing.T) {
	Convey("BuildFormFieldsInput", t, func() {
		Convey("types", func() {
			form := url.Values{
				"/given_name":       {"John"},
				"/x_age":            {"42"},
				"/x_score":          {"1.5"},
				"/x_terms":          {"true", "false"},
				"/address/locality": {"London"},
				"/address/country":  {"GB"},
			}
			in := BuildFormFieldsInput(form, formData(
				f("/given_name", declarative.FillFormFieldTypeString, false),
				f("/x_age", declarative.FillFormFieldTypeInteger, false),
				f("/x_score", declarative.FillFormFieldTypeNumber, false),
				f("/x_terms", declarative.FillFormFieldTypeBoolean, false),
				f("/address", declarative.FillFormFieldTypeAddress, false),
			))
			So(in.Keys, ShouldResemble, []string{"/given_name", "/x_age", "/x_score", "/x_terms", "/address"})
			So(in.Fields, ShouldResemble, []map[string]any{
				{"key": "/given_name", "value": "John"},
				{"key": "/x_age", "value": int64(42)},
				{"key": "/x_score", "value": 1.5},
				{"key": "/x_terms", "value": true},
				{"key": "/address", "value": map[string]any{"locality": "London", "country": "GB"}},
			})
		})

		Convey("empty fields are skipped, and Fields is never nil", func() {
			in := BuildFormFieldsInput(url.Values{"/address/locality": {""}}, formData(
				f("/given_name", declarative.FillFormFieldTypeString, false),
				f("/x_plan", declarative.FillFormFieldTypeEnum, false),
				f("/address", declarative.FillFormFieldTypeAddress, false),
			))
			So(in.Fields, ShouldNotBeNil)
			So(in.Fields, ShouldBeEmpty)
			So(in.Keys, ShouldBeEmpty)
			b, _ := json.Marshal(map[string]any{"fields": in.Fields})
			So(string(b), ShouldEqual, `{"fields":[]}`)
		})

		Convey("an unchecked boolean is false, and the first value wins", func() {
			data := formData(f("/x_terms", declarative.FillFormFieldTypeBoolean, false))
			So(BuildFormFieldsInput(url.Values{"/x_terms": {"false"}}, data).Fields[0]["value"], ShouldEqual, false)
			So(BuildFormFieldsInput(url.Values{}, data).Fields[0]["value"], ShouldEqual, false)
			So(BuildFormFieldsInput(url.Values{"/x_terms": {"true", "false"}}, data).Fields[0]["value"], ShouldEqual, true)
		})

		Convey("a number that does not parse is sent as the raw string", func() {
			in := BuildFormFieldsInput(url.Values{"/x_age": {"abc"}}, formData(f("/x_age", declarative.FillFormFieldTypeInteger, false)))
			So(in.Fields[0]["value"], ShouldEqual, "abc")
		})

		Convey("NaN and Inf are sent as the raw string", func() {
			data := formData(f("/x_score", declarative.FillFormFieldTypeNumber, false))
			for _, raw := range []string{"NaN", "Inf", "-infinity", "1e400"} {
				in := BuildFormFieldsInput(url.Values{"/x_score": {raw}}, data)
				So(in.Fields[0]["value"], ShouldEqual, raw)
			}
			So(BuildFormFieldsInput(url.Values{"/x_score": {"1.5"}}, data).Fields[0]["value"], ShouldEqual, 1.5)
		})

		Convey("a nil form", func() {
			in := BuildFormFieldsInput(nil, formData(f("/given_name", declarative.FillFormFieldTypeString, false)))
			So(in.Fields, ShouldBeEmpty)
		})
	})
}

func TestFormErrorLocations(t *testing.T) {
	Convey("FormErrorLocations", t, func() {
		locations := FormErrorLocations(formData(), []string{"/given_name", "/address"})
		So(locations["/given_name"], ShouldResemble, []string{"/fields/0/value"})
		So(locations["/address"], ShouldResemble, []string{
			"/fields/1/value",
			"/fields/1/value/formatted",
			"/fields/1/value/street_address",
			"/fields/1/value/locality",
			"/fields/1/value/region",
			"/fields/1/value/postal_code",
			"/fields/1/value/country",
		})
	})
}

func hookError(reasons ...map[string]any) *apierrors.APIError {
	rs := []any{}
	for _, r := range reasons {
		rs = append(rs, r)
	}
	return roundTrip(&apierrors.APIError{
		Kind: apierrors.Kind{Name: "Forbidden", Reason: "HookDisallowed"},
		Info_ReadOnly: map[string]any{
			"reasons": []any{map[string]any{"title": "t", "reason": "r", "reasons": rs}},
		},
	})
}

func TestHookFieldMessages(t *testing.T) {
	Convey("HookFieldMessages", t, func() {
		Convey("matched and unmatched keys", func() {
			err := hookError(
				map[string]any{"type": "invalid_form_field", "key": "/a", "message": "bad a"},
				map[string]any{"type": "invalid_form_field", "key": "/zzz", "message": "bad zzz"},
				map[string]any{"type": "other", "key": "/a", "message": "ignored"},
			)
			byKey, unmatched := HookFieldMessages(err, []string{"/a", "/b"})
			So(byKey, ShouldResemble, map[string]string{"/a": "bad a"})
			So(unmatched, ShouldResemble, []string{"bad zzz"})
		})

		Convey("a second message of a field is unmatched", func() {
			err := hookError(
				map[string]any{"type": "invalid_form_field", "key": "/a", "message": "one"},
				map[string]any{"type": "invalid_form_field", "key": "/a", "message": "two"},
			)
			byKey, unmatched := HookFieldMessages(err, []string{"/a"})
			So(byKey["/a"], ShouldEqual, "one")
			So(unmatched, ShouldResemble, []string{"two"})
		})

		Convey("not a hook error", func() {
			byKey, unmatched := HookFieldMessages(roundTrip(&apierrors.APIError{Kind: apierrors.Kind{Name: "Invalid", Reason: "ValidationFailed"}}), []string{"/a"})
			So(byKey, ShouldBeEmpty)
			So(unmatched, ShouldBeEmpty)
			byKey, _ = HookFieldMessages(nil, nil)
			So(byKey, ShouldBeEmpty)
		})
	})
}

func validationError(causes ...map[string]any) *apierrors.APIError {
	cs := []any{}
	for _, c := range causes {
		cs = append(cs, c)
	}
	return roundTrip(&apierrors.APIError{
		Kind:          apierrors.Kind{Name: "Invalid", Reason: "ValidationFailed"},
		Info_ReadOnly: map[string]any{"causes": cs},
	})
}

func TestNewFormViewModel(t *testing.T) {
	Convey("NewFormViewModel", t, func() {
		form := formData(
			f("/given_name", declarative.FillFormFieldTypeString, true),
			f("/family_name", declarative.FillFormFieldTypeString, true),
			f("/nickname", declarative.FillFormFieldTypeString, false),
			f("/x_terms", declarative.FillFormFieldTypeBoolean, false),
		)
		form.Fields[2].Value = "Johnny"

		Convey("without an error", func() {
			vm, err := NewFormViewModel(form, &viewmodels.BaseViewModel{})
			So(err, ShouldBeNil)
			So(vm.Fields, ShouldHaveLength, 4)
			So(vm.Fields[2].Value, ShouldEqual, "Johnny")
			So(vm.UnknownError, ShouldBeNil)
			So(vm.Timezones, ShouldNotBeEmpty)
		})

		Convey("restored values win over current values", func() {
			base := &viewmodels.BaseViewModel{}
			base.SetFormJSON(url.Values{"/nickname": {"Changed"}, "/x_terms": {"true", "false"}})
			vm, err := NewFormViewModel(form, base)
			So(err, ShouldBeNil)
			So(vm.Fields[2].Value, ShouldEqual, "Changed")
			So(vm.Fields[3].Checked, ShouldBeTrue)
		})

		Convey("a cause is attributed to the field by its position among the submitted fields", func() {
			// Only the 2nd and 3rd form fields were submitted, so /fields/0 is /family_name.
			base := &viewmodels.BaseViewModel{
				RawError: validationError(map[string]any{"location": "/fields/0/value", "kind": "minLength", "details": map[string]any{"expected": 2, "actual": 1}}),
			}
			base.RestoredForm = url.Values{"/family_name": {"D"}, "/nickname": {"x"}}
			vm, err := NewFormViewModel(form, base)
			So(err, ShouldBeNil)
			So(vm.Fields[1].Error, ShouldNotBeNil)
			So(vm.Fields[0].Error, ShouldBeNil)
			So(vm.Fields[2].Error, ShouldBeNil)
			So(vm.UnknownError, ShouldBeNil)
		})

		Convey("a cause at /fields marks the missing required fields", func() {
			base := &viewmodels.BaseViewModel{
				RawError: validationError(map[string]any{"location": "/fields", "kind": "contains"}),
			}
			base.RestoredForm = url.Values{"/given_name": {"John"}}
			vm, err := NewFormViewModel(form, base)
			So(err, ShouldBeNil)
			So(vm.Fields[0].MissingRequired, ShouldBeFalse)
			So(vm.Fields[1].MissingRequired, ShouldBeTrue)
			So(vm.Fields[2].MissingRequired, ShouldBeFalse)
			// It is not shown as a general error.
			So(vm.UnknownError, ShouldBeNil)
		})

		Convey("causes at /fields/<i>/key mark the missing required fields and no submitted field", func() {
			// With /given_name submitted and /family_name missing, the validator reports a cause per submitted field.
			base := &viewmodels.BaseViewModel{
				RawError: validationError(map[string]any{"location": "/fields/0/key", "kind": "const", "details": map[string]any{"expected": "/family_name", "actual": "/given_name"}},
					map[string]any{"location": "/fields/0/value", "kind": "minLength", "details": map[string]any{"expected": 1, "actual": 0}}),
			}
			base.RestoredForm = url.Values{"/given_name": {"John"}}
			vm, err := NewFormViewModel(form, base)
			So(err, ShouldBeNil)
			So(vm.Fields[0].Error, ShouldBeNil)
			So(vm.Fields[0].MissingRequired, ShouldBeFalse)
			So(vm.Fields[1].MissingRequired, ShouldBeTrue)
			So(vm.UnknownError, ShouldBeNil)
		})

		Convey("an unchecked required checkbox is flagged by the /fields/<i>/key causes, with no error on the filled field", func() {
			form := formData(
				f("/given_name", declarative.FillFormFieldTypeString, true),
				f("/x_terms", declarative.FillFormFieldTypeBoolean, true),
			)
			base := &viewmodels.BaseViewModel{
				RawError: validationError(map[string]any{"location": "/fields/0/key", "kind": "const", "details": map[string]any{"expected": "/x_terms", "actual": "/given_name"}},
					map[string]any{"location": "/fields/0/value", "kind": "const", "details": map[string]any{"expected": true, "actual": "John"}},
					map[string]any{"location": "/fields/1/value", "kind": "const", "details": map[string]any{"expected": true, "actual": false}}),
			}
			base.RestoredForm = url.Values{"/given_name": {"John"}, "/x_terms": {"false"}}
			vm, err := NewFormViewModel(form, base)
			So(err, ShouldBeNil)
			So(vm.Fields[0].Error, ShouldBeNil)
			So(vm.Fields[0].MissingRequired, ShouldBeFalse)
			So(vm.Fields[1].MissingRequired, ShouldBeTrue)
		})

		Convey("an unchecked required checkbox is a missing required field", func() {
			form := formData(
				f("/x_terms", declarative.FillFormFieldTypeBoolean, true),
				f("/x_privacy", declarative.FillFormFieldTypeBoolean, true),
			)
			base := &viewmodels.BaseViewModel{
				RawError: validationError(map[string]any{"location": "/fields/0/value", "kind": "const", "details": map[string]any{"expected": true, "actual": false}}),
			}
			base.RestoredForm = url.Values{"/x_terms": {"false"}, "/x_privacy": {"true", "false"}}
			vm, err := NewFormViewModel(form, base)
			So(err, ShouldBeNil)
			So(vm.Fields[0].MissingRequired, ShouldBeTrue)
			So(vm.Fields[1].MissingRequired, ShouldBeFalse)
		})

		Convey("hook messages", func() {
			base := &viewmodels.BaseViewModel{
				RawError: hookError(
					map[string]any{"type": "invalid_form_field", "key": "/nickname", "message": "bad nickname"},
					map[string]any{"type": "invalid_form_field", "key": "/other", "message": "bad other"},
				),
			}
			vm, err := NewFormViewModel(form, base)
			So(err, ShouldBeNil)
			So(vm.Fields[2].HookMessage, ShouldEqual, "bad nickname")
			So(vm.HookUnmatchedMessages, ShouldResemble, []string{"bad other"})
			// The hook title and reason are in the general error.
			So(vm.UnknownError, ShouldNotBeNil)
		})
	})
}

func TestIsFormError(t *testing.T) {
	Convey("isFormError", t, func() {
		So(isFormError(nil), ShouldBeFalse)
		So(isFormError(validationError(map[string]any{"location": "/fields/0/value"}, map[string]any{"location": "/fields"})), ShouldBeTrue)
		So(isFormError(validationError(map[string]any{"location": "/fields/0/value"}, map[string]any{"location": "/login_id"})), ShouldBeFalse)
		So(isFormError(validationError(map[string]any{"location": "/fieldsX"})), ShouldBeFalse)
		formHook := roundTrip(&apierrors.APIError{
			Kind:          apierrors.Kind{Name: "Forbidden", Reason: "HookDisallowed"},
			Info_ReadOnly: map[string]any{"event_type": "authentication.form.post_submitted"},
		})
		So(isFormError(formHook), ShouldBeTrue)
		otherHook := roundTrip(&apierrors.APIError{
			Kind:          apierrors.Kind{Name: "Forbidden", Reason: "HookDisallowed"},
			Info_ReadOnly: map[string]any{"event_type": "authentication.post_identified"},
		})
		So(isFormError(otherHook), ShouldBeFalse)
	})
}
