package analytic

import (
	"errors"
)

var ErrAnalyticCountNotFound = errors.New("analytic count not found")
var ErrAnalyticRedisIsNotConfigured = errors.New("analytic redis is not configured")
