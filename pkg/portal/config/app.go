package config

import (
	"errors"
	"strings"
)

type AppConfig struct {
	HostTemplate string              `envconfig:"HOST_TEMPLATE"`
	IDPattern    string              `envconfig:"ID_PATTERN" default:"^[a-z0-9][a-z0-9-]{2,30}[a-z0-9]$"`
	Secret       AppSecretConfig     `envconfig:"SECRET"`
	Kubernetes   AppKubernetesConfig `envconfig:"KUBERNETES"`
}

type AppTLSCertType string

const (
	AppTLSCertNone        AppTLSCertType = "none"
	AppTLSCertStatic      AppTLSCertType = "static"
	AppTLSCertCertManager AppTLSCertType = "cert-manager"
)

type AppTLSCertSource struct {
	Type AppTLSCertType

	// for static type
	SecretName string

	// for cert-manager type
	IssuerKind string
	IssuerName string
}

func ParseTLSCertSource(desc string) (*AppTLSCertSource, error) {
	parts := strings.Split(desc, ":")
	switch AppTLSCertType(parts[0]) {
	case AppTLSCertNone:
		if len(parts) != 1 {
			return nil, errors.New("'none' certificate type expects no arguments")
		}
		return &AppTLSCertSource{
			Type: AppTLSCertNone,
		}, nil

	case AppTLSCertStatic:
		if len(parts) != 2 {
			return nil, errors.New("'static' certificate type expects 1 argument")
		}
		return &AppTLSCertSource{
			Type:       AppTLSCertStatic,
			SecretName: parts[1],
		}, nil

	case AppTLSCertCertManager:
		if len(parts) != 3 {
			return nil, errors.New("'static' certificate type expects 2 arguments")
		}
		return &AppTLSCertSource{
			Type:       AppTLSCertCertManager,
			IssuerKind: parts[1],
			IssuerName: parts[2],
		}, nil

	default:
		return nil, errors.New("unknown certificate type")
	}
}

type AppKubernetesConfig struct {
	NewResourcePrefix    string `envconfig:"NEW_RESOURCE_PREFIX" default:"app-"`
	IngressTemplateFile  string `envconfig:"INGRESS_TEMPLATE_FILE"`
	DefaultDomainTLSCert string `envconfig:"DEFAULT_DOMAIN_TLS_CERT" default:"none"`
	CustomDomainTLSCert  string `envconfig:"CUSTOM_DOMAIN_TLS_CERT" default:"none"`
}

type AppSecretConfig struct {
	DatabaseURL      string `envconfig:"DATABASE_URL"`
	DatabaseSchema   string `envconfig:"DATABASE_SCHEMA"`
	RedisURL         string `envconfig:"REDIS_URL"`
	SMTPHost         string `envconfig:"SMTP_HOST"`
	SMTPPort         int    `envconfig:"SMTP_PORT"`
	SMTPMode         string `envconfig:"SMTP_MODE"`
	SMTPUsername     string `envconfig:"SMTP_USERNAME"`
	SMTPPassword     string `envconfig:"SMTP_PASSWORD"`
	TwilioAccountSID string `envconfig:"TWILIO_ACCOUNT_SID"`
	TwilioAuthToken  string `envconfig:"TWILIO_AUTH_TOKEN"`
	NexmoAPIKey      string `envconfig:"NEXMO_API_KEY"`
	NexmoAPISecret   string `envconfig:"NEXMO_API_SECRET"`
}
