package config

var _ = SecretConfigSchema.Add("TelemetryAuditLogStreamTLSMaterials", `
{
	"type": "array",
	"items": { "$ref": "#/$defs/TelemetryAuditLogStreamTLSMaterialsItem" }
}
`)

type TelemetryAuditLogStreamTLSMaterials []TelemetryAuditLogStreamTLSMaterialsItem

var _ SecretItemData = &TelemetryAuditLogStreamTLSMaterials{}

// Resolve returns the material declared for streamName.
func (m *TelemetryAuditLogStreamTLSMaterials) Resolve(streamName string) (*TelemetryAuditLogStreamTLSMaterialsItem, bool) {
	if m == nil {
		return nil, false
	}
	for idx := range *m {
		item := (*m)[idx]
		if item.StreamName == streamName {
			return &item, true
		}
	}
	return nil, false
}

func (m *TelemetryAuditLogStreamTLSMaterials) SensitiveStrings() []string {
	return nil
}

var _ = SecretConfigSchema.Add("TelemetryAuditLogStreamTLSMaterialsItem", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"stream_name": { "type": "string", "pattern": "^[a-zA-Z0-9_-]{1,63}$" },
		"client_certificate": { "$ref": "#/$defs/TelemetryAuditLogStreamClientCertificate" },
		"certificate_authority": { "$ref": "#/$defs/X509Certificate" }
	},
	"required": ["stream_name"],
	"anyOf": [
		{ "required": ["client_certificate"] },
		{ "required": ["certificate_authority"] }
	]
}
`)

type TelemetryAuditLogStreamTLSMaterialsItem struct {
	StreamName string `json:"stream_name,omitempty"`
	// ClientCertificate and CertificateAuthority are nullable so
	// config.SetFieldDefaults leaves an omitted one nil, rather than
	// materializing an empty struct that is indistinguishable from a
	// genuinely configured one to a nil check.
	ClientCertificate    *TelemetryAuditLogStreamClientCertificate `json:"client_certificate,omitempty" nullable:"true"`
	CertificateAuthority *X509Certificate                          `json:"certificate_authority,omitempty" nullable:"true"`
}

var _ = SecretConfigSchema.Add("TelemetryAuditLogStreamClientCertificate", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"certificate": { "$ref": "#/$defs/X509Certificate" },
		"key": { "$ref": "#/$defs/JWK" }
	},
	"required": ["certificate", "key"]
}
`)

type TelemetryAuditLogStreamClientCertificate struct {
	Certificate *X509Certificate `json:"certificate,omitempty"`
	Key         *JWK             `json:"key,omitempty"`
}

var _ = SecretConfigSchema.Add("TelemetryAuditLogStreamDatadogCredentials", `
{
	"type": "array",
	"items": { "$ref": "#/$defs/TelemetryAuditLogStreamDatadogCredentialsItem" }
}
`)

type TelemetryAuditLogStreamDatadogCredentials []TelemetryAuditLogStreamDatadogCredentialsItem

var _ SecretItemData = &TelemetryAuditLogStreamDatadogCredentials{}

// Resolve returns the credentials declared for streamName.
func (c *TelemetryAuditLogStreamDatadogCredentials) Resolve(streamName string) (*TelemetryAuditLogStreamDatadogCredentialsItem, bool) {
	if c == nil {
		return nil, false
	}
	for idx := range *c {
		item := (*c)[idx]
		if item.StreamName == streamName {
			return &item, true
		}
	}
	return nil, false
}

// SensitiveStrings returns every api_key, so that a key cannot appear in a
// log line. This differs from TelemetryAuditLogStreamTLSMaterials, which
// returns nil: a certificate is not a bearer credential, an API key is.
func (c *TelemetryAuditLogStreamDatadogCredentials) SensitiveStrings() []string {
	if c == nil {
		return nil
	}
	var out []string
	for _, item := range *c {
		if item.APIKey != "" {
			out = append(out, item.APIKey)
		}
	}
	return out
}

var _ = SecretConfigSchema.Add("TelemetryAuditLogStreamDatadogCredentialsItem", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"stream_name": { "type": "string", "pattern": "^[a-zA-Z0-9_-]{1,63}$" },
		"api_key": { "type": "string", "minLength": 1 }
	},
	"required": ["stream_name", "api_key"]
}
`)

type TelemetryAuditLogStreamDatadogCredentialsItem struct {
	StreamName string `json:"stream_name,omitempty"`
	APIKey     string `json:"api_key,omitempty"`
}
