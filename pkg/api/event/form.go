package event

// FormFieldRef identifies a field of a form. It has exactly one non-nil member.
type FormFieldRef struct {
	UserProfile *FormFieldRefUserProfile `json:"user_profile,omitempty"`
}

type FormFieldRefUserProfile struct {
	Pointer string `json:"pointer"`
}

// Key returns the Authflow API key of the field, i.e. the pointer for user_profile.
func (r FormFieldRef) Key() string {
	if r.UserProfile != nil {
		return r.UserProfile.Pointer
	}
	return ""
}

type SubmittedForm struct {
	Name   string               `json:"name,omitempty"`
	Fields []SubmittedFormField `json:"fields"`
}

type SubmittedFormField struct {
	Field FormFieldRef `json:"field"`
	// Value is null when the field is empty.
	Value any `json:"value"`
}

type HookResponseReasonType string

const HookResponseReasonTypeInvalidFormField HookResponseReasonType = "invalid_form_field"

type HookResponseReason struct {
	Type    HookResponseReasonType `json:"type"`
	Field   *FormFieldRef          `json:"field,omitempty"`
	Message string                 `json:"message,omitempty"`
}
