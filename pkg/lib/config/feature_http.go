package config

var _ = FeatureConfigSchema.Add("HTTPFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"insecure_fetch_address_allowed": { "type": "boolean" },
		"insecure_fetch_address_allowed_hosts": {
			"type": "array",
			"items": { "type": "string", "minLength": 1 }
		}
	}
}
`)

// HTTPFeatureConfig governs the outbound fetches whose destination comes from
// a project's own configuration: event webhooks, the custom SMS provider, the
// account migration hook, the phone number verification hook, and an OAuth
// provider's OIDC discovery document (and in turn its jwks_uri).
//
// Those URLs are written by whoever administers the project, which on a
// deployment with open sign-up is any user who created one. Reaching the
// deployment's own network is therefore not theirs to grant, which is why this
// lives in the feature config: authgear.features.yaml is not writable through
// the portal or the Admin API (configsource.AuthgearFeatureYAMLDescriptor's
// UpdateResource refuses outright), so only the operator can set it, at
// whichever layer suits the deployment.
type HTTPFeatureConfig struct {
	// InsecureFetchAddressAllowed permits those fetches to connect to a
	// non-publicly-routable address, including 169.254.169.254.
	//
	// Off by default. A self-hosted deployment whose webhook receiver or
	// identity provider genuinely sits on a private network is the reason it
	// exists; a deployment serving projects it does not control should leave
	// it off.
	InsecureFetchAddressAllowed *bool `json:"insecure_fetch_address_allowed,omitempty"`
	// InsecureFetchAddressAllowedHosts names the hosts exempt from the
	// restriction, so that an operator can allow their own internal receiver
	// without allowing every non-public address.
	//
	// A hostname, or a leading "*." matching exactly one label. Matched against
	// the host as configured, before resolution, so whatever the host resolves
	// to is then accepted -- list only hosts this deployment controls.
	//
	// omitzero, not omitempty: nil means "not set, inherit" while an explicit
	// [] means "clear what a lower layer set", and those must stay distinct.
	InsecureFetchAddressAllowedHosts []string `json:"insecure_fetch_address_allowed_hosts,omitzero"`
}

func (c *HTTPFeatureConfig) SetDefaults() {
	if c.InsecureFetchAddressAllowed == nil {
		c.InsecureFetchAddressAllowed = new(false)
	}
}

func (c *HTTPFeatureConfig) IsInsecureFetchAddressAllowed() bool {
	return c != nil && c.InsecureFetchAddressAllowed != nil && *c.InsecureFetchAddressAllowed
}

func (c *HTTPFeatureConfig) GetInsecureFetchAddressAllowedHosts() []string {
	if c == nil {
		return nil
	}
	return c.InsecureFetchAddressAllowedHosts
}

var _ MergeableFeatureConfig = &HTTPFeatureConfig{}

// Merge is field-level rather than a whole-section replace, so that a later
// layer setting one field does not reset the siblings a lower layer set. The
// section has a single field today; merging per-field anyway means adding the
// second one cannot silently reintroduce that bug.
func (c *HTTPFeatureConfig) Merge(layer *FeatureConfig) MergeableFeatureConfig {
	if layer.HTTP == nil {
		return c
	}

	merged := c
	if merged == nil {
		merged = &HTTPFeatureConfig{}
	}

	if layer.HTTP.InsecureFetchAddressAllowed != nil {
		merged.InsecureFetchAddressAllowed = layer.HTTP.InsecureFetchAddressAllowed
	}
	if layer.HTTP.InsecureFetchAddressAllowedHosts != nil {
		merged.InsecureFetchAddressAllowedHosts = layer.HTTP.InsecureFetchAddressAllowedHosts
	}

	return merged
}
