package config

import "fmt"

var _ = Schema.Add("TelemetryConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"audit_logs": { "$ref": "#/$defs/TelemetryAuditLogsConfig" }
	}
}
`)

type TelemetryConfig struct {
	AuditLogs *TelemetryAuditLogsConfig `json:"audit_logs,omitempty"`
}

var _ = Schema.Add("TelemetryAuditLogsConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"streams": {
			"type": "array",
			"items": { "$ref": "#/$defs/TelemetryAuditLogStreamConfig" }
		}
	}
}
`)

type TelemetryAuditLogsConfig struct {
	Streams []*TelemetryAuditLogStreamConfig `json:"streams,omitempty"`
}

type TelemetryAuditLogStreamType string

const (
	TelemetryAuditLogStreamTypeSyslog  TelemetryAuditLogStreamType = "syslog"
	TelemetryAuditLogStreamTypeDatadog TelemetryAuditLogStreamType = "datadog"
)

type TelemetryAuditLogStreamTransport string

const (
	TelemetryAuditLogStreamTransportTCP  TelemetryAuditLogStreamTransport = "tcp"
	TelemetryAuditLogStreamTransportHTTP TelemetryAuditLogStreamTransport = "http"
)

// DatadogSite is the site parameter of a Datadog organization, not the
// display name of one. It selects the intake endpoint; see LogsIntakeURL.
//
// It is a free-form string, not a closed enum: which sites exist is
// Datadog's deployment to govern, not this project's, so a site newer
// than this code (or one this code has never heard of) still works. A
// value Datadog does not recognise fails at delivery time, the same
// failure mode as a wrong http.endpoint, not at config save time.
//
// DatadogSiteUS1 is the only named constant, because it is the only site
// this package itself needs to refer to (the default). A second named
// constant here would just be an unenforced, staleness-prone copy of
// Datadog's site list -- the exact thing relaxing this field away from an
// enum was meant to avoid.
type DatadogSite string

const (
	DatadogSiteUS1 DatadogSite = "datadoghq.com"
)

// LogsIntakeURL returns the logs HTTP intake URL of s. Every Datadog site
// spells this host the same way, so this is a single interpolation
// rather than a switch -- and it is why s does not need to be validated
// against a known list for this to work correctly.
func (s DatadogSite) LogsIntakeURL() string {
	return fmt.Sprintf("https://http-intake.logs.%s/api/v2/logs", string(s))
}

type TelemetryAuditLogStreamDatadogConfig struct {
	Site    DatadogSite       `json:"site,omitempty"`
	Service string            `json:"service,omitempty"`
	Source  string            `json:"source,omitempty"`
	Tags    map[string]string `json:"tags,omitempty"`
}

func (c *TelemetryAuditLogStreamDatadogConfig) SetDefaults() {
	if c.Site == "" {
		c.Site = DatadogSiteUS1
	}
	if c.Service == "" {
		c.Service = "authgear"
	}
	if c.Source == "" {
		c.Source = "authgear"
	}
}

type TelemetryAuditLogStreamHTTPConfig struct {
	Endpoint string `json:"endpoint,omitempty"`
}

var _ = Schema.Add("TelemetryAuditLogStreamConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"name": { "type": "string", "pattern": "^[a-zA-Z0-9_-]{1,63}$" },
		"type": { "type": "string", "enum": ["syslog", "datadog"] },
		"transport": { "type": "string", "enum": ["tcp", "http"] },
		"tcp": { "$ref": "#/$defs/TelemetryAuditLogStreamTCPConfig" },
		"http": { "$ref": "#/$defs/TelemetryAuditLogStreamHTTPConfig" },
		"syslog": { "$ref": "#/$defs/TelemetryAuditLogStreamSyslogConfig" },
		"datadog": { "$ref": "#/$defs/TelemetryAuditLogStreamDatadogConfig" }
	},
	"required": ["name", "type", "transport"],
	"allOf": [
		{
			"if": { "properties": { "type": { "const": "syslog" } }, "required": ["type"] },
			"then": {
				"required": ["syslog"],
				"not": { "required": ["datadog"] },
				"properties": { "transport": { "enum": ["tcp"] } }
			}
		},
		{
			"if": { "properties": { "type": { "const": "datadog" } }, "required": ["type"] },
			"then": {
				"not": { "required": ["syslog"] },
				"properties": { "transport": { "enum": ["http"] } }
			}
		},
		{
			"if": { "properties": { "transport": { "const": "tcp" } }, "required": ["transport"] },
			"then": {
				"required": ["tcp"],
				"not": { "required": ["http"] }
			}
		},
		{
			"if": { "properties": { "transport": { "const": "http" } }, "required": ["transport"] },
			"then": { "not": { "required": ["tcp"] } }
		},
		{
			"if": {
				"properties": { "http": { "required": ["endpoint"] } },
				"required": ["http"]
			},
			"then": { "properties": { "datadog": { "not": { "required": ["site"] } } } }
		}
	]
}
`)

var _ = Schema.Add("TelemetryAuditLogStreamHTTPConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"endpoint": { "type": "string", "format": "x_http_url" }
	}
}
`)

var _ = Schema.Add("TelemetryAuditLogStreamDatadogConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"site": { "type": "string", "minLength": 1, "maxLength": 253 },
		"service": { "type": "string", "minLength": 1, "maxLength": 100 },
		"source": { "type": "string", "minLength": 1, "maxLength": 100 },
		"tags": {
			"type": "object",
			"maxProperties": 20,
			"additionalProperties": { "type": "string", "minLength": 1 }
		}
	}
}
`)

type TelemetryAuditLogStreamConfig struct {
	Name      string                           `json:"name,omitempty"`
	Type      TelemetryAuditLogStreamType      `json:"type,omitempty"`
	Transport TelemetryAuditLogStreamTransport `json:"transport,omitempty"`

	// The four objects below are nullable so that SetFieldDefaults does not
	// materialize them. A stream carries exactly the encoding object its
	// type selects and exactly the transport object its transport selects;
	// SetDefaults below allocates those two and leaves the other two nil.
	TCP     *TelemetryAuditLogStreamTCPConfig     `json:"tcp,omitempty" nullable:"true"`
	HTTP    *TelemetryAuditLogStreamHTTPConfig    `json:"http,omitempty" nullable:"true"`
	Syslog  *TelemetryAuditLogStreamSyslogConfig  `json:"syslog,omitempty" nullable:"true"`
	Datadog *TelemetryAuditLogStreamDatadogConfig `json:"datadog,omitempty" nullable:"true"`
}

func (c *TelemetryAuditLogStreamConfig) SetDefaults() {
	switch c.Type {
	case TelemetryAuditLogStreamTypeSyslog:
		if c.Syslog == nil {
			c.Syslog = &TelemetryAuditLogStreamSyslogConfig{}
		}
		c.Syslog.SetDefaults()
	case TelemetryAuditLogStreamTypeDatadog:
		if c.Datadog == nil {
			c.Datadog = &TelemetryAuditLogStreamDatadogConfig{}
		}
		c.Datadog.SetDefaults()
	}

	switch c.Transport {
	case TelemetryAuditLogStreamTransportTCP:
		if c.TCP == nil {
			c.TCP = &TelemetryAuditLogStreamTCPConfig{}
		}
		c.TCP.SetDefaults()
	case TelemetryAuditLogStreamTransportHTTP:
		if c.HTTP == nil {
			c.HTTP = &TelemetryAuditLogStreamHTTPConfig{}
		}
	}
}

var _ = Schema.Add("TelemetryAuditLogStreamTCPConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"address": { "type": "string", "format": "x_host_port" },
		"tls": { "$ref": "#/$defs/TelemetryAuditLogStreamTCPTLSConfig" }
	},
	"required": ["address"]
}
`)

type TelemetryAuditLogStreamTCPConfig struct {
	Address string                               `json:"address,omitempty"`
	TLS     *TelemetryAuditLogStreamTCPTLSConfig `json:"tls,omitempty" nullable:"true"`
}

func (c *TelemetryAuditLogStreamTCPConfig) SetDefaults() {
	if c.TLS == nil {
		c.TLS = &TelemetryAuditLogStreamTCPTLSConfig{}
	}
}

var _ = Schema.Add("TelemetryAuditLogStreamTCPTLSConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"enabled": { "type": "boolean" }
	},
	"required": ["enabled"]
}
`)

type TelemetryAuditLogStreamTCPTLSConfig struct {
	Enabled bool `json:"enabled"`
}

type SyslogFormat string

const (
	SyslogFormatRFC5424 SyslogFormat = "rfc5424"
)

type SyslogFraming string

const (
	SyslogFramingOctetCounting SyslogFraming = "octet_counting"
	SyslogFramingNewline       SyslogFraming = "newline"
)

// SyslogFacility names a syslog facility, matching the set redis.conf's
// syslog-facility setting accepts. See Code() for the RFC 5424 Table 1
// numeric code each name maps to.
type SyslogFacility string

const (
	SyslogFacilityUser   SyslogFacility = "user"
	SyslogFacilityLocal0 SyslogFacility = "local0"
	SyslogFacilityLocal1 SyslogFacility = "local1"
	SyslogFacilityLocal2 SyslogFacility = "local2"
	SyslogFacilityLocal3 SyslogFacility = "local3"
	SyslogFacilityLocal4 SyslogFacility = "local4"
	SyslogFacilityLocal5 SyslogFacility = "local5"
	SyslogFacilityLocal6 SyslogFacility = "local6"
	SyslogFacilityLocal7 SyslogFacility = "local7"
)

// Code returns f's numeric syslog facility code, per Table 1 in RFC 5424
// section 6.2.1. Schema validation restricts
// TelemetryAuditLogStreamSyslogConfig.Facility to exactly the named
// constants above, so every value reaching here has a case.
func (f SyslogFacility) Code() int {
	switch f {
	case SyslogFacilityUser:
		return 1
	case SyslogFacilityLocal0:
		return 16
	case SyslogFacilityLocal1:
		return 17
	case SyslogFacilityLocal2:
		return 18
	case SyslogFacilityLocal3:
		return 19
	case SyslogFacilityLocal4:
		return 20
	case SyslogFacilityLocal5:
		return 21
	case SyslogFacilityLocal6:
		return 22
	case SyslogFacilityLocal7:
		return 23
	default:
		panic("config: unknown syslog facility: " + string(f))
	}
}

var _ = Schema.Add("TelemetryAuditLogStreamSyslogConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"format": { "type": "string", "enum": ["rfc5424"] },
		"framing": { "type": "string", "enum": ["octet_counting", "newline"] },
		"facility": { "type": "string", "enum": ["user", "local0", "local1", "local2", "local3", "local4", "local5", "local6", "local7"] },
		"app_name": { "type": "string", "pattern": "^[!-~]{1,48}$" },
		"structured_data_id": { "type": "string", "pattern": "^[!#-<>-\\\\^-~]{1,32}$" }
	},
	"required": ["format", "framing"]
}
`)

type TelemetryAuditLogStreamSyslogConfig struct {
	Format           SyslogFormat   `json:"format,omitempty"`
	Framing          SyslogFraming  `json:"framing,omitempty"`
	Facility         SyslogFacility `json:"facility,omitempty"`
	AppName          string         `json:"app_name,omitempty"`
	StructuredDataID string         `json:"structured_data_id,omitempty"`
}

func (c *TelemetryAuditLogStreamSyslogConfig) SetDefaults() {
	if c.Facility == "" {
		c.Facility = SyslogFacilityLocal0
	}
	if c.AppName == "" {
		c.AppName = "authgear"
	}
	if c.StructuredDataID == "" {
		c.StructuredDataID = "authgear"
	}
}
