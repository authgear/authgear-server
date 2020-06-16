package config

import (
	"github.com/skygeario/skygear-server/pkg/validation"
)

var Schema = validation.NewMultipartSchema("AppConfig")

func init() {
	Schema.Instantiate()
}
