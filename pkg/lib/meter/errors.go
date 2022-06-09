package meter

import (
	"errors"
)

var ErrMeterRedisIsNotConfigured = errors.New("meter redis is not configured")
