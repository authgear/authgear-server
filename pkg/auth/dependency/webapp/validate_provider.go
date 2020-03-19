package webapp

import (
	"net/url"
)

// ValidateProvider validates url.Values with JSON Schema.
type ValidateProvider interface {
	// PrepareValues removes empty values and populate defaults.
	PrepareValues(form url.Values)
	// validate validate from against the schema.
	Validate(schemaID string, form url.Values) error
}
