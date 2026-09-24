package config

var _ = FeatureConfigSchema.Add("TelemetryFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"audit_logs": { "$ref": "#/$defs/TelemetryAuditLogsFeatureConfig" }
	}
}
`)

type TelemetryFeatureConfig struct {
	AuditLogs *TelemetryAuditLogsFeatureConfig `json:"audit_logs,omitempty"`
}

var _ MergeableFeatureConfig = &TelemetryFeatureConfig{}

func (c *TelemetryFeatureConfig) Merge(layer *FeatureConfig) MergeableFeatureConfig {
	return c.merge(layer.Telemetry)
}

func (c *TelemetryFeatureConfig) merge(layer *TelemetryFeatureConfig) *TelemetryFeatureConfig {
	if c == nil && layer == nil {
		return nil
	}
	if c == nil {
		return layer
	}
	if layer == nil {
		return c
	}
	c.AuditLogs = c.AuditLogs.merge(layer.AuditLogs)
	return c
}

var _ = FeatureConfigSchema.Add("TelemetryAuditLogsFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"streaming": { "$ref": "#/$defs/TelemetryAuditLogsStreamingFeatureConfig" }
	}
}
`)

type TelemetryAuditLogsFeatureConfig struct {
	Streaming *TelemetryAuditLogsStreamingFeatureConfig `json:"streaming,omitempty"`
}

func (c *TelemetryAuditLogsFeatureConfig) merge(layer *TelemetryAuditLogsFeatureConfig) *TelemetryAuditLogsFeatureConfig {
	if c == nil && layer == nil {
		return nil
	}
	if c == nil {
		return layer
	}
	if layer == nil {
		return c
	}
	c.Streaming = c.Streaming.merge(layer.Streaming)
	return c
}

var _ = FeatureConfigSchema.Add("TelemetryAuditLogsStreamingFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"disabled": { "type": "boolean" }
	}
}
`)

type TelemetryAuditLogsStreamingFeatureConfig struct {
	Disabled *bool `json:"disabled,omitempty"`
}

func (c *TelemetryAuditLogsStreamingFeatureConfig) SetDefaults() {
	if c.Disabled == nil {
		c.Disabled = new(false)
	}
}

func (c *TelemetryAuditLogsStreamingFeatureConfig) merge(layer *TelemetryAuditLogsStreamingFeatureConfig) *TelemetryAuditLogsStreamingFeatureConfig {
	if c == nil && layer == nil {
		return nil
	}
	if c == nil {
		return layer
	}
	if layer == nil {
		return c
	}
	if layer.Disabled != nil {
		c.Disabled = layer.Disabled
	}
	return c
}
