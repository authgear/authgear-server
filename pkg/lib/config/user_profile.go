package config

var _ = Schema.Add("UserProfileConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"standard_attributes": { "$ref": "#/$defs/StandardAttributesConfig" }
	}
}
`)

var _ = Schema.Add("StandardAttributesPopulationConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"strategy": {
			"type": "string",
			"enum": ["none", "on_signup"]
		}
	}
}
`)

var _ = Schema.Add("StandardAttributesConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"population": { "$ref": "#/$defs/StandardAttributesPopulationConfig" },
		"access_control": {
			"type": "array",
			"items": { "$ref": "#/$defs/StandardAttributesAccessControlConfig" }
		}
	}
}
`)

var _ = Schema.Add("StandardAttributesAccessControlConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"pointer": {
			"type": "string",
			"format": "json-pointer",
			"enum": [
				"/email",
				"/email_verified",
				"/phone_number",
				"/phone_number_verified",
				"/preferred_username",
				"/family_name",
				"/given_name",
				"/picture",
				"/gender",
				"/birthdate",
				"/zoneinfo",
				"/locale",
				"/name",
				"/nickname",
				"/middle_name",
				"/profile",
				"/website",
				"/address"
			]
		},
		"access_control": { "$ref": "#/$defs/StandardAttributesAccessControl" }
	}
}
`)

var _ = Schema.Add("StandardAttributesAccessControl", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"end_user": { "$ref": "#/$defs/AccessControlLevelString" },
		"bearer": { "$ref": "#/$defs/AccessControlLevelString" },
		"portal_ui": { "$ref": "#/$defs/AccessControlLevelString" }
	},
	"enum": [
		{
			"end_user": "hidden",
			"bearer": "hidden",
			"portal_ui": "hidden"
		},
		{
			"end_user": "hidden",
			"bearer": "hidden",
			"portal_ui": "readonly"
		},
		{
			"end_user": "hidden",
			"bearer": "hidden",
			"portal_ui": "readwrite"
		},
		{
			"end_user": "hidden",
			"bearer": "readonly",
			"portal_ui": "readonly"
		},
		{
			"end_user": "hidden",
			"bearer": "readonly",
			"portal_ui": "readwrite"
		},
		{
			"end_user": "hidden",
			"bearer": "readwrite",
			"portal_ui": "readwrite"
		},
		{
			"end_user": "readonly",
			"bearer": "readonly",
			"portal_ui": "readonly"
		},
		{
			"end_user": "readonly",
			"bearer": "readonly",
			"portal_ui": "readwrite"
		},
		{
			"end_user": "readonly",
			"bearer": "readwrite",
			"portal_ui": "readwrite"
		},
		{
			"end_user": "readwrite",
			"bearer": "readwrite",
			"portal_ui": "readwrite"
		}
	]
}
`)

var _ = Schema.Add("AccessControlLevelString", `
{
	"type": "string",
	"enum": ["hidden", "readonly", "readwrite"]
}
`)

type UserProfileConfig struct {
	StandardAttributes *StandardAttributesConfig `json:"standard_attributes,omitempty"`
}

type StandardAttributesConfig struct {
	Population    *StandardAttributesPopulationConfig      `json:"population,omitempty"`
	AccessControl []*StandardAttributesAccessControlConfig `json:"access_control,omitempty"`
}

func (c *StandardAttributesConfig) SetDefaults() {
	readwrite := &StandardAttributesAccessControl{
		EndUser:  AccessControlLevelStringReadwrite,
		Bearer:   AccessControlLevelStringReadwrite,
		PortalUI: AccessControlLevelStringReadwrite,
	}

	hidden := &StandardAttributesAccessControl{
		EndUser:  AccessControlLevelStringHidden,
		Bearer:   AccessControlLevelStringHidden,
		PortalUI: AccessControlLevelStringHidden,
	}

	if c.AccessControl == nil {
		c.AccessControl = []*StandardAttributesAccessControlConfig{
			{
				Pointer:       "/email",
				AccessControl: readwrite,
			},
			{
				Pointer:       "/email_verified",
				AccessControl: readwrite,
			},
			{
				Pointer:       "/phone_number",
				AccessControl: readwrite,
			},
			{
				Pointer:       "/phone_number_verified",
				AccessControl: readwrite,
			},
			{
				Pointer:       "/preferred_username",
				AccessControl: readwrite,
			},
			{
				Pointer:       "/family_name",
				AccessControl: readwrite,
			},
			{
				Pointer:       "/given_name",
				AccessControl: readwrite,
			},
			{
				Pointer:       "/picture",
				AccessControl: readwrite,
			},
			{
				Pointer:       "/gender",
				AccessControl: readwrite,
			},
			{
				Pointer:       "/birthdate",
				AccessControl: readwrite,
			},
			{
				Pointer:       "/zoneinfo",
				AccessControl: readwrite,
			},
			{
				Pointer:       "/locale",
				AccessControl: readwrite,
			},
			{
				Pointer:       "/name",
				AccessControl: hidden,
			},
			{
				Pointer:       "/nickname",
				AccessControl: hidden,
			},
			{
				Pointer:       "/middle_name",
				AccessControl: hidden,
			},
			{
				Pointer:       "/profile",
				AccessControl: hidden,
			},
			{
				Pointer:       "/website",
				AccessControl: hidden,
			},
			{
				Pointer:       "/address",
				AccessControl: hidden,
			},
		}
	}
}

type StandardAttributesAccessControlConfig struct {
	Pointer       string                           `json:"pointer,omitempty"`
	AccessControl *StandardAttributesAccessControl `json:"access_control,omitempty"`
}

type StandardAttributesAccessControl struct {
	EndUser  AccessControlLevelString `json:"end_user,omitempty"`
	Bearer   AccessControlLevelString `json:"bearer,omitempty"`
	PortalUI AccessControlLevelString `json:"portal_ui,omitempty"`
}

type AccessControlLevelString string

const (
	AccessControlLevelStringDefault   AccessControlLevelString = ""
	AccessControlLevelStringHidden    AccessControlLevelString = "hidden"
	AccessControlLevelStringReadonly  AccessControlLevelString = "readonly"
	AccessControlLevelStringReadwrite AccessControlLevelString = "readwrite"
)

type StandardAttributesPopulationStrategy string

const (
	StandardAttributesPopulationStrategyDefault  StandardAttributesPopulationStrategy = ""
	StandardAttributesPopulationStrategyNone     StandardAttributesPopulationStrategy = "none"
	StandardAttributesPopulationStrategyOnSignup StandardAttributesPopulationStrategy = "on_signup"
)

type StandardAttributesPopulationConfig struct {
	Strategy StandardAttributesPopulationStrategy `json:"strategy,omitempty"`
}

func (c *StandardAttributesPopulationConfig) SetDefaults() {
	if c.Strategy == StandardAttributesPopulationStrategyDefault {
		c.Strategy = StandardAttributesPopulationStrategyOnSignup
	}
}
