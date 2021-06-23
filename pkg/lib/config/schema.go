package config

import (
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var Schema = validation.NewMultipartSchema("AppConfig")

var SecretConfigSchema = validation.NewMultipartSchema("SecretConfig")

var FeatureConfigSchema = validation.NewMultipartSchema("FeatureConfig")

func init() {
	Schema.Instantiate()
	SecretConfigSchema.Instantiate()
	FeatureConfigSchema.Instantiate()
}

func DumpSchema() (string, error) {
	return Schema.DumpSchemaString(true)
}

func DumpSecretConfigSchema() (string, error) {
	return SecretConfigSchema.DumpSchemaString(true)
}
