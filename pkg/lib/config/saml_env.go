package config

type SAMLEnvironmentConfig struct {
	IdPEntityIDTemplate string `envconfig:"IDP_ENTITY_ID_TEMPLATE" default:"urn:{{.app_id}}.localhost"`
}
