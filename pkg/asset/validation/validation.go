package validation

import (
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

var (
	Validator *validation.Validator
)

func init() {
	// The actual initialization is in main.go
	Validator = validation.NewValidator("http://v2.skygear.io")
}
