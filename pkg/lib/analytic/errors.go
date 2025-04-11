package analytic

import (
	"errors"
)

var ErrAnalyticCountNotFound = errors.New("analytic count not found")
var ErrMissingPosthogCredential = errors.New("missing posthog credential")
