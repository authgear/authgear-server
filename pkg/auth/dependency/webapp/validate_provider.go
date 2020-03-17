package webapp

import (
	"net/url"
)

// ValidateProvider validates url.Values with JSON Schema.
type ValidateProvider interface {
	Validate(schemaID string, form url.Values) error
}
