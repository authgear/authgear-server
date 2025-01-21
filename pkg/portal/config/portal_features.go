package config

type PortalFeaturesConfig struct {
	ShowCustomSMSGateway bool `envconfig:"SHOW_CUSTOM_SMS_GATEWAY" default:"false"`
}
