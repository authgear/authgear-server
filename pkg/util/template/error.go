package template

import (
	"errors"
)

var ErrLimitReached = errors.New("template: rendered template is too large")
var ErrNotFound = errors.New("requested template not found")
var ErrUpdateDisallowed = errors.New("template: update disallowed")
var ErrMissingFeatureFlagInCtx = errors.New("template: feature flag config missing in ctx")
