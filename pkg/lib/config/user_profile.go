package config

import (
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
)

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

type UserProfileConfig struct {
	StandardAttributes *StandardAttributesConfig `json:"standard_attributes,omitempty"`
}

type StandardAttributesConfig struct {
	Population    *StandardAttributesPopulationConfig      `json:"population,omitempty"`
	AccessControl []*StandardAttributesAccessControlConfig `json:"access_control,omitempty"`
}

func (c *StandardAttributesConfig) SetDefaults() {
	defaultReadwrite := &StandardAttributesAccessControl{
		EndUser:  AccessControlLevelStringReadwrite,
		Bearer:   AccessControlLevelStringReadonly,
		PortalUI: AccessControlLevelStringReadwrite,
	}

	defaultHidden := &StandardAttributesAccessControl{
		EndUser:  AccessControlLevelStringHidden,
		Bearer:   AccessControlLevelStringHidden,
		PortalUI: AccessControlLevelStringHidden,
	}

	if c.AccessControl == nil {
		c.AccessControl = []*StandardAttributesAccessControlConfig{
			{
				Pointer:       "/family_name",
				AccessControl: defaultReadwrite,
			},
			{
				Pointer:       "/given_name",
				AccessControl: defaultReadwrite,
			},
			{
				Pointer:       "/picture",
				AccessControl: defaultReadwrite,
			},
			{
				Pointer:       "/gender",
				AccessControl: defaultReadwrite,
			},
			{
				Pointer:       "/birthdate",
				AccessControl: defaultReadwrite,
			},
			{
				Pointer:       "/zoneinfo",
				AccessControl: defaultReadwrite,
			},
			{
				Pointer:       "/locale",
				AccessControl: defaultReadwrite,
			},
			{
				Pointer:       "/name",
				AccessControl: defaultHidden,
			},
			{
				Pointer:       "/nickname",
				AccessControl: defaultHidden,
			},
			{
				Pointer:       "/middle_name",
				AccessControl: defaultHidden,
			},
			{
				Pointer:       "/profile",
				AccessControl: defaultHidden,
			},
			{
				Pointer:       "/website",
				AccessControl: defaultHidden,
			},
			{
				Pointer:       "/address",
				AccessControl: defaultHidden,
			},
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
