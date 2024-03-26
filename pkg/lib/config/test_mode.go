package config

import "regexp"

type rule interface {
	GetRegex() *regexp.Regexp
}

type rules[R rule] interface {
	GetRules() []R
	MatchTarget(target string) (R, bool)
}

var _ = Schema.Add("TestModeConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"oob_otp": { "$ref": "#/$defs/TestModeOOBOTPConfig" },
		"sms": { "$ref": "#/$defs/TestModeSMSConfig" },
		"whatsapp": { "$ref": "#/$defs/TestModeWhatsappConfig" },
		"email": { "$ref": "#/$defs/TestModeEmailConfig" }
	}
}
`)

type TestModeConfig struct {
	FixedOOBOTP *TestModeOOBOTPConfig   `json:"oob_otp,omitempty"`
	SMS         *TestModeSMSConfig      `json:"sms,omitempty"`
	Whatsapp    *TestModeWhatsappConfig `json:"whatsapp,omitempty"`
	Email       *TestModeEmailConfig    `json:"email,omitempty"`
}

var _ = Schema.Add("TestModeOOBOTPConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"enabled": { "type": "boolean" },
		"rules": { "type": "array", "items": { "$ref": "#/$defs/TestModeOOBOTPRule" } }
	}
}
`)

type TestModeOOBOTPConfig struct {
	Enabled bool                  `json:"enabled,omitempty"`
	Rules   []*TestModeOOBOTPRule `json:"rules,omitempty"`
}

var _ rules[*TestModeOOBOTPRule] = &TestModeOOBOTPConfig{}

func (c *TestModeOOBOTPConfig) GetRules() []*TestModeOOBOTPRule {
	return c.Rules
}

func (c *TestModeOOBOTPConfig) MatchTarget(target string) (*TestModeOOBOTPRule, bool) {
	return matchTestModeRulesWithTarget[*TestModeOOBOTPRule](c, target)
}

var _ = Schema.Add("TestModeOOBOTPRule", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"regex": { "type": "string", "format": "x_re2_regex" },
		"fixed_code": { "type": "string" }
	},
	"required": ["regex"]
}
`)

type TestModeOOBOTPRule struct {
	Regex     string `json:"regex,omitempty"`
	FixedCode string `json:"fixed_code,omitempty"`
}

func (r *TestModeOOBOTPRule) GetRegex() *regexp.Regexp {
	return regexp.MustCompile(r.Regex)
}

var _ rule = &TestModeOOBOTPRule{}

var _ = Schema.Add("TestModeSMSConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"enabled": { "type": "boolean" },
		"rules": { "type": "array", "items": { "$ref": "#/$defs/TestModeSMSRule" } }
	}
}
`)

type TestModeSMSConfig struct {
	Enabled bool               `json:"enabled,omitempty"`
	Rules   []*TestModeSMSRule `json:"rules,omitempty"`
}

var _ rules[*TestModeSMSRule] = &TestModeSMSConfig{}

func (c *TestModeSMSConfig) GetRules() []*TestModeSMSRule {
	return c.Rules
}
func (c *TestModeSMSConfig) MatchTarget(target string) (*TestModeSMSRule, bool) {
	return matchTestModeRulesWithTarget[*TestModeSMSRule](c, target)
}

var _ = Schema.Add("TestModeSMSRule", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"regex": { "type": "string", "format": "x_re2_regex" },
		"suppressed": { "type": "boolean" }
	},
	"required": ["regex"]
}
`)

type TestModeSMSRule struct {
	Regex      string `json:"regex,omitempty"`
	Suppressed bool   `json:"suppressed,omitempty"`
}

var _ rule = &TestModeSMSRule{}

func (r *TestModeSMSRule) GetRegex() *regexp.Regexp {
	return regexp.MustCompile(r.Regex)
}

var _ = Schema.Add("TestModeWhatsappConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"enabled": { "type": "boolean" },
		"rules": { "type": "array", "items": { "$ref": "#/$defs/TestModeWhatsappRule" } }
	}
}
`)

type TestModeWhatsappConfig struct {
	Enabled bool                    `json:"enabled,omitempty"`
	Rules   []*TestModeWhatsappRule `json:"rules,omitempty"`
}

var _ rules[*TestModeWhatsappRule] = &TestModeWhatsappConfig{}

func (c *TestModeWhatsappConfig) GetRules() []*TestModeWhatsappRule {
	return c.Rules
}
func (c *TestModeWhatsappConfig) MatchTarget(target string) (*TestModeWhatsappRule, bool) {
	return matchTestModeRulesWithTarget[*TestModeWhatsappRule](c, target)
}

var _ = Schema.Add("TestModeWhatsappRule", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"regex": { "type": "string", "format": "x_re2_regex" },
		"suppressed": { "type": "boolean" }
	},
	"required": ["regex"]
}
`)

type TestModeWhatsappRule struct {
	Regex      string `json:"regex,omitempty"`
	Suppressed bool   `json:"suppressed,omitempty"`
}

var _ rule = &TestModeWhatsappRule{}

func (r *TestModeWhatsappRule) GetRegex() *regexp.Regexp {
	return regexp.MustCompile(r.Regex)
}

var _ = Schema.Add("TestModeEmailConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"enabled": { "type": "boolean" },
		"rules": { "type": "array", "items": { "$ref": "#/$defs/TestModeEmailRule" } }
	}
}
`)

type TestModeEmailConfig struct {
	Enabled bool                 `json:"enabled,omitempty"`
	Rules   []*TestModeEmailRule `json:"rules,omitempty"`
}

var _ rules[*TestModeEmailRule] = &TestModeEmailConfig{}

func (c *TestModeEmailConfig) GetRules() []*TestModeEmailRule {
	return c.Rules
}
func (c *TestModeEmailConfig) MatchTarget(target string) (*TestModeEmailRule, bool) {
	return matchTestModeRulesWithTarget[*TestModeEmailRule](c, target)
}

var _ = Schema.Add("TestModeEmailRule", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"regex": { "type": "string", "format": "x_re2_regex" },
		"suppressed": { "type": "boolean" },
		"fixed_code": { "type": "string" }
	},
	"required": ["regex"]
}
`)

type TestModeEmailRule struct {
	Regex      string `json:"regex,omitempty"`
	Suppressed bool   `json:"suppressed,omitempty"`
	FixedCode  string `json:"fixed_code,omitempty"`
}

var _ rule = &TestModeEmailRule{}

func (r *TestModeEmailRule) GetRegex() *regexp.Regexp {
	return regexp.MustCompile(r.Regex)
}

func matchTestModeRulesWithTarget[R rule](rs rules[R], target string) (R, bool) {
	for _, r := range rs.GetRules() {
		reg := r.GetRegex()
		if reg.Match([]byte(target)) {
			return r, true
		}
	}
	var zero R
	return zero, false
}
