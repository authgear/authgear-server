package webapp

import (
	"net/url"
)

// ValidateProvider validates url.Values with JSON Schema.
type ValidateProvider interface {
	// Prevalidate fills in default values.
	Prevalidate(form url.Values)
	// validate validate from against the schema.
	Validate(schemaID string, form url.Values) error
}
