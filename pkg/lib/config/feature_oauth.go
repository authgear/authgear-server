package config

var _ = FeatureConfigSchema.Add("OAuthFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"client": { "$ref": "#/$defs/OAuthClientFeatureConfig" },
		"client_id_metadata_document": { "$ref": "#/$defs/OAuthClientIDMetadataDocumentFeatureConfig" }
	}
}
`)

type OAuthFeatureConfig struct {
	Client                   *OAuthClientFeatureConfig                   `json:"client,omitempty"`
	ClientIDMetadataDocument *OAuthClientIDMetadataDocumentFeatureConfig `json:"client_id_metadata_document,omitempty"`
}

var _ MergeableFeatureConfig = &OAuthFeatureConfig{}

func (c *OAuthFeatureConfig) Merge(layer *FeatureConfig) MergeableFeatureConfig {
	if layer.OAuth == nil {
		return c
	}

	var merged *OAuthFeatureConfig = c
	if merged == nil {
		merged = &OAuthFeatureConfig{}
	}

	merged.Client = merged.Client.Merge(layer.OAuth.Client)
	merged.ClientIDMetadataDocument = merged.ClientIDMetadataDocument.Merge(layer.OAuth.ClientIDMetadataDocument)

	return merged
}

func (c *OAuthFeatureConfig) GetClientIDMetadataDocument() *OAuthClientIDMetadataDocumentFeatureConfig {
	if c == nil {
		return nil
	}
	return c.ClientIDMetadataDocument
}

var _ = FeatureConfigSchema.Add("OAuthClientFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"maximum": { "type": "integer" },
		"soft_maximum": { "type": "integer" },
		"custom_ui_enabled": { "type": "boolean" },
		"app2app_enabled": { "type": "boolean" }
	}
}
`)

type OAuthClientFeatureConfig struct {
	Maximum         *int  `json:"maximum,omitempty"`
	SoftMaximum     *int  `json:"soft_maximum,omitempty"`
	CustomUIEnabled *bool `json:"custom_ui_enabled,omitempty"`
	App2AppEnabled  *bool `json:"app2app_enabled,omitempty"`
}

func (c *OAuthClientFeatureConfig) SetDefaults() {
	if c.Maximum == nil {
		c.Maximum = new(99)
	}

	if c.SoftMaximum == nil {
		c.SoftMaximum = new(99)
	}

	if c.CustomUIEnabled == nil {
		c.CustomUIEnabled = new(false)
	}

	if c.App2AppEnabled == nil {
		c.App2AppEnabled = new(false)
	}
}

func (c *OAuthClientFeatureConfig) Merge(layer *OAuthClientFeatureConfig) *OAuthClientFeatureConfig {
	if c == nil && layer == nil {
		return nil
	}
	if c == nil {
		return layer
	}
	if layer == nil {
		return c
	}
	if layer.Maximum != nil {
		c.Maximum = layer.Maximum
	}
	if layer.SoftMaximum != nil {
		c.SoftMaximum = layer.SoftMaximum
	}
	if layer.App2AppEnabled != nil {
		c.App2AppEnabled = layer.App2AppEnabled
	}
	if layer.CustomUIEnabled != nil {
		c.CustomUIEnabled = layer.CustomUIEnabled
	}
	return c
}

var _ = FeatureConfigSchema.Add("OAuthClientIDMetadataDocumentFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"insecure_http_allowed": { "type": "boolean" },
		"insecure_fetch_address_allowed": { "type": "boolean" }
	}
}
`)

// OAuthClientIDMetadataDocumentFeatureConfig defeats the CIMD fetch's
// transport protections, for test and local-development projects only.
// Never set on a project serving real traffic, and never at the cluster or
// plan layer -- app-specific override only. See docs/specs/cimd.md § SSRF
// Protection.
type OAuthClientIDMetadataDocumentFeatureConfig struct {
	// InsecureHTTPAllowed permits http:// wherever CIMD requires https: the
	// client_id, the document's uri fields, and the logo fetch. Scheme only.
	InsecureHTTPAllowed *bool `json:"insecure_http_allowed,omitempty"`
	// InsecureFetchAddressAllowed permits connecting to a
	// non-publicly-routable address, including 169.254.169.254.
	InsecureFetchAddressAllowed *bool `json:"insecure_fetch_address_allowed,omitempty"`
}

func (c *OAuthClientIDMetadataDocumentFeatureConfig) SetDefaults() {
	if c.InsecureHTTPAllowed == nil {
		c.InsecureHTTPAllowed = new(false)
	}
	if c.InsecureFetchAddressAllowed == nil {
		c.InsecureFetchAddressAllowed = new(false)
	}
}

func (c *OAuthClientIDMetadataDocumentFeatureConfig) IsInsecureHTTPAllowed() bool {
	return c != nil && c.InsecureHTTPAllowed != nil && *c.InsecureHTTPAllowed
}

func (c *OAuthClientIDMetadataDocumentFeatureConfig) IsInsecureFetchAddressAllowed() bool {
	return c != nil && c.InsecureFetchAddressAllowed != nil && *c.InsecureFetchAddressAllowed
}

// Merge is field-level, not a whole-section replace: a higher layer setting
// only one of the two flags must not reset the other back to its code
// default from a lower layer. *bool (not bool) is what lets nil mean "this
// layer said nothing" -- the only way a higher layer can force a flag back
// off, and the only shape that survives the marshal-then-reparse round trip
// in AuthgearFeatureYAMLDescriptor.viewEffectiveResource.
func (c *OAuthClientIDMetadataDocumentFeatureConfig) Merge(layer *OAuthClientIDMetadataDocumentFeatureConfig) *OAuthClientIDMetadataDocumentFeatureConfig {
	if c == nil && layer == nil {
		return nil
	}
	if c == nil {
		return layer
	}
	if layer == nil {
		return c
	}
	if layer.InsecureHTTPAllowed != nil {
		c.InsecureHTTPAllowed = layer.InsecureHTTPAllowed
	}
	if layer.InsecureFetchAddressAllowed != nil {
		c.InsecureFetchAddressAllowed = layer.InsecureFetchAddressAllowed
	}
	return c
}
