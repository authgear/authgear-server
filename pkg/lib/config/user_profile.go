package config

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
)

var _ = Schema.Add("UserProfileConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"standard_attributes": { "$ref": "#/$defs/StandardAttributesConfig" },
		"custom_attributes": { "$ref": "#/$defs/CustomAttributesConfig" }
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
				"/phone_number",
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
		"access_control": { "$ref": "#/$defs/UserProfileAttributesAccessControl" }
	}
}
`)

var _ = Schema.Add("UserProfileAttributesAccessControl", `
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
			"end_user": "readwrite",
			"bearer": "readonly",
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

var defaultReadwriteStandardAttributesPointers []string = []string{
	"/email",
	"/phone_number",
	"/preferred_username",
	"/family_name",
	"/given_name",
	"/picture",
	"/gender",
	"/birthdate",
	"/zoneinfo",
	"/locale",
}

var defaultHiddenStandardAttributesPointers []string = []string{
	"/name",
	"/nickname",
	"/middle_name",
	"/profile",
	"/website",
	"/address",
}

var _ = Schema.Add("CustomAttributesConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"attributes": {
			"type": "array",
			"items": {
				"$ref": "#/$defs/CustomAttributesAttributeConfig"
			}
		}
	}
}
`)

// It seems impossible to write additionalProperties: false without duplicating the common schema in every variant :(
// See https://json-schema.org/understanding-json-schema/reference/combining.html
var _ = Schema.Add("CustomAttributesAttributeConfig", `
{
	"type": "object",
	"properties": {
		"id": {
			"type": "string",
			"minLength": 1
		},
		"pointer": {
			"type": "string",
			"pattern": "^/[a-zA-Z0-9_]+$",
			"not": {
				"enum": [
					"/iss",
					"/sub",
					"/aud",
					"/exp",
					"/nbf",
					"/iat",
					"/jti",

					"/sub",

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
					"/address",

					"/updated_at"
				]
			}
		},
		"type": {
			"type": "string",
			"enum": [
				"string",
				"number",
				"integer",
				"enum",
				"phone_number",
				"email",
				"url",
				"alpha2"
			]
		},
		"access_control": { "$ref": "#/$defs/UserProfileAttributesAccessControl" }
	},
	"required": ["id", "pointer", "type"],
	"allOf": [
		{
			"if": {
				"properties": { "type": { "const": "number" } }
			},
			"then": {
				"properties": {
					"minimum": {
						"type": "number"
					},
					"maximum": {
						"type": "number"
					}
				}
			}
		},
		{
			"if": {
				"properties": { "type": { "const": "integer" } }
			},
			"then": {
				"properties": {
					"minimum": {
						"type": "integer"
					},
					"maximum": {
						"type": "integer"
					}
				}
			}
		},
		{
			"if": {
				"properties": { "type": { "const": "enum" } }
			},
			"then": {
				"properties": {
					"enum": {
						"type": "array",
						"items": {
							"type": "string",
							"minLength": 1
						},
						"minItems": 1,
						"uniqueItems": true
					}
				},
				"required": ["enum"]
			}
		},
		{
			"if": {
				"properties": {
					"type": {
						"not": {
							"enum": [
								"number",
								"integer",
								"enum"
							]
						}
					}
				}
			},
			"then": true
		}
	]
}
`)

type UserProfileConfig struct {
	StandardAttributes *StandardAttributesConfig `json:"standard_attributes,omitempty"`
	CustomAttributes   *CustomAttributesConfig   `json:"custom_attributes,omitempty"`
}

type CustomAttributesConfig struct {
	Attributes []*CustomAttributesAttributeConfig `json:"attributes,omitempty"`
}

func (c *CustomAttributesConfig) GetAccessControl() accesscontrol.T {
	t := accesscontrol.T{}
	for _, a := range c.Attributes {
		subject := accesscontrol.Subject(a.Pointer)
		t[subject] = map[accesscontrol.Role]accesscontrol.Level{
			RoleEndUser:  a.AccessControl.EndUser.Level(),
			RoleBearer:   a.AccessControl.Bearer.Level(),
			RolePortalUI: a.AccessControl.PortalUI.Level(),
		}
	}
	return t
}

type CustomAttributesAttributeConfig struct {
	ID            string                              `json:"id,omitempty"`
	Pointer       string                              `json:"pointer,omitempty"`
	Type          CustomAttributeType                 `json:"type,omitempty"`
	AccessControl *UserProfileAttributesAccessControl `json:"access_control,omitempty"`
	Minimum       *float64                            `json:"minimum,omitempty"`
	Maximum       *float64                            `json:"maximum,omitempty"`
	Enum          []string                            `json:"enum,omitempty"`
}

func (c *CustomAttributesAttributeConfig) SetDefaults() {
	if c.AccessControl == nil {
		c.AccessControl = &UserProfileAttributesAccessControl{}
	}
	if c.AccessControl.EndUser == AccessControlLevelStringDefault {
		c.AccessControl.EndUser = AccessControlLevelStringHidden
	}
	if c.AccessControl.Bearer == AccessControlLevelStringDefault {
		c.AccessControl.Bearer = AccessControlLevelStringHidden
	}
	if c.AccessControl.PortalUI == AccessControlLevelStringDefault {
		c.AccessControl.PortalUI = AccessControlLevelStringReadwrite
	}
}

func (c *CustomAttributesAttributeConfig) ToJSONSchema() (schema map[string]interface{}, err error) {
	schema = make(map[string]interface{})

	switch c.Type {
	case CustomAttributeTypeString:
		schema["type"] = "string"
	case CustomAttributeTypeNumber:
		schema["type"] = "number"
		if c.Minimum != nil {
			schema["minimum"] = *c.Minimum
		}
		if c.Maximum != nil {
			schema["maximum"] = *c.Maximum
		}
	case CustomAttributeTypeInteger:
		schema["type"] = "integer"
		if c.Minimum != nil {
			schema["minimum"] = int64(*c.Minimum)
		}
		if c.Maximum != nil {
			schema["maximum"] = int64(*c.Maximum)
		}
	case CustomAttributeTypeEnum:
		schema["type"] = "string"
		schema["enum"] = c.Enum
	case CustomAttributeTypePhoneNumber:
		schema["type"] = "string"
		schema["format"] = "phone"
	case CustomAttributeTypeEmail:
		schema["type"] = "string"
		schema["format"] = "email"
	case CustomAttributeTypeURL:
		schema["type"] = "string"
		schema["format"] = "uri"
	case CustomAttributeTypeAlpha2:
		schema["type"] = "string"
		schema["format"] = "iso3166-1-alpha-2"
	default:
		err = fmt.Errorf("unknown type: %v", c.Type)
	}

	return
}

type CustomAttributeType string

const (
	CustomAttributeTypeString      CustomAttributeType = "string"
	CustomAttributeTypeNumber      CustomAttributeType = "number"
	CustomAttributeTypeInteger     CustomAttributeType = "integer"
	CustomAttributeTypeEnum        CustomAttributeType = "enum"
	CustomAttributeTypePhoneNumber CustomAttributeType = "phone_number"
	CustomAttributeTypeEmail       CustomAttributeType = "email"
	CustomAttributeTypeURL         CustomAttributeType = "url"
	CustomAttributeTypeAlpha2      CustomAttributeType = "alpha2"
)

type StandardAttributesConfig struct {
	Population    *StandardAttributesPopulationConfig      `json:"population,omitempty"`
	AccessControl []*StandardAttributesAccessControlConfig `json:"access_control,omitempty"`
}

func (c *StandardAttributesConfig) SetDefaults() {
	defaultReadwrite := &UserProfileAttributesAccessControl{
		EndUser:  AccessControlLevelStringReadwrite,
		Bearer:   AccessControlLevelStringReadonly,
		PortalUI: AccessControlLevelStringReadwrite,
	}

	defaultHidden := &UserProfileAttributesAccessControl{
		EndUser:  AccessControlLevelStringHidden,
		Bearer:   AccessControlLevelStringHidden,
		PortalUI: AccessControlLevelStringHidden,
	}

	for _, pointer := range defaultReadwriteStandardAttributesPointers {
		found := false
		for _, a := range c.AccessControl {
			if pointer == a.Pointer {
				found = true
			}
		}
		if !found {
			c.AccessControl = append(c.AccessControl, &StandardAttributesAccessControlConfig{
				Pointer:       pointer,
				AccessControl: defaultReadwrite,
			})
		}
	}
	for _, pointer := range defaultHiddenStandardAttributesPointers {
		found := false
		for _, a := range c.AccessControl {
			if pointer == a.Pointer {
				found = true
			}
		}
		if !found {
			c.AccessControl = append(c.AccessControl, &StandardAttributesAccessControlConfig{
				Pointer:       pointer,
				AccessControl: defaultHidden,
			})
		}
	}
}

func (c *StandardAttributesConfig) GetAccessControl() accesscontrol.T {
	t := accesscontrol.T{}
	for _, a := range c.AccessControl {
		subject := accesscontrol.Subject(a.Pointer)
		t[subject] = map[accesscontrol.Role]accesscontrol.Level{
			RoleEndUser:  a.AccessControl.EndUser.Level(),
			RoleBearer:   a.AccessControl.Bearer.Level(),
			RolePortalUI: a.AccessControl.PortalUI.Level(),
		}
	}
	return t
}

func (c *StandardAttributesConfig) IsEndUserAllHidden() bool {
	if len(c.AccessControl) <= 0 {
		return false
	}
	for _, a := range c.AccessControl {
		if a.AccessControl.EndUser.Level() != AccessControlLevelHidden {
			return false
		}
	}
	return true
}

type StandardAttributesAccessControlConfig struct {
	Pointer       string                              `json:"pointer,omitempty"`
	AccessControl *UserProfileAttributesAccessControl `json:"access_control,omitempty"`
}

type UserProfileAttributesAccessControl struct {
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

const (
	AccessControlLevelHidden    accesscontrol.Level = 1
	AccessControlLevelReadonly  accesscontrol.Level = 2
	AccessControlLevelReadwrite accesscontrol.Level = 3
)

func (l AccessControlLevelString) Level() accesscontrol.Level {
	switch l {
	case AccessControlLevelStringHidden:
		return AccessControlLevelHidden
	case AccessControlLevelStringReadonly:
		return AccessControlLevelReadonly
	case AccessControlLevelStringReadwrite:
		return AccessControlLevelReadwrite
	}
	return 0
}

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

const (
	RoleEndUser  accesscontrol.Role = "end_user"
	RoleBearer   accesscontrol.Role = "bearer"
	RolePortalUI accesscontrol.Role = "portal_ui"
)
