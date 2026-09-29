package config

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
	TelemetryAuditLogStreamTypeSyslog TelemetryAuditLogStreamType = "syslog"
)

type TelemetryAuditLogStreamTransport string

const (
	TelemetryAuditLogStreamTransportTCP TelemetryAuditLogStreamTransport = "tcp"
)

var _ = Schema.Add("TelemetryAuditLogStreamConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"name": { "type": "string", "pattern": "^[a-zA-Z0-9_-]{1,63}$" },
		"type": { "type": "string", "enum": ["syslog"] },
		"transport": { "type": "string", "enum": ["tcp"] },
		"tcp": { "$ref": "#/$defs/TelemetryAuditLogStreamTCPConfig" },
		"syslog": { "$ref": "#/$defs/TelemetryAuditLogStreamSyslogConfig" }
	},
	"required": ["name", "type", "transport"],
	"allOf": [
		{
			"if": { "properties": { "type": { "const": "syslog" } }, "required": ["type"] },
			"then": { "required": ["syslog"] }
		},
		{
			"if": { "properties": { "transport": { "const": "tcp" } }, "required": ["transport"] },
			"then": { "required": ["tcp"] }
		},
		{
			"if": { "properties": { "type": { "const": "syslog" } }, "required": ["type"] },
			"then": { "properties": { "transport": { "enum": ["tcp"] } } }
		}
	]
}
`)

type TelemetryAuditLogStreamConfig struct {
	Name      string                           `json:"name,omitempty"`
	Type      TelemetryAuditLogStreamType      `json:"type,omitempty"`
	Transport TelemetryAuditLogStreamTransport `json:"transport,omitempty"`

	TCP    *TelemetryAuditLogStreamTCPConfig    `json:"tcp,omitempty"`
	Syslog *TelemetryAuditLogStreamSyslogConfig `json:"syslog,omitempty"`
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
	TLS     *TelemetryAuditLogStreamTCPTLSConfig `json:"tls,omitempty"`
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
