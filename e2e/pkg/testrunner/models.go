package testrunner

type BeforeHook struct {
	Type       BeforeHookType      `yaml:"type"`
	UserImport string              `yaml:"user_import"`
	CustomSQL  BeforeHookCustomSQL `yaml:"custom_sql"`
}

var _ = TestCaseSchema.Add("AuthgearYAMLSource", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"extend": { "type": "file-path", "description": "Path to the base authgear.yaml" },
		"override": { "type": "string", "description": "Inline snippet to override the base authgear.yaml" }
	}
}
`)

type AuthgearYAMLSource struct {
	Extend   string `yaml:"extend"`
	Override string `yaml:"override"`
}

type BeforeHookType string

const (
	BeforeHookTypeUserImport BeforeHookType = "user_import"
	BeforeHookTypeCustomSQL  BeforeHookType = "custom_sql"
)

var _ = TestCaseSchema.Add("BeforeHookCustomSQL", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"path": { "type": "string", "format": "file-path", "description": "Path to the custom SQL script" }
	},
	"required": ["path"]
}
`)

type BeforeHookCustomSQL struct {
	Path string `yaml:"path"`
}

var _ = TestCaseSchema.Add("BeforeHook", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"type": { "type": "string", "enum": ["user_import", "custom_sql"] },
		"user_import": { "type": "string" },
		"custom_sql": { "$ref": "#/$defs/BeforeHookCustomSQL" }
	},
	"required": ["type"],
	"allOf": [
			{
				"if": { "properties": { "type": { "const": "user_import" } } },
				"then": {
					"required": ["user_import"]
				}
			},
			{
				"if": { "properties": { "type": { "const": "custom_sql" } } },
				"then": {
					"required": ["custom_sql"]
				}
			}
		]
}
`)

var _ = TestCaseSchema.Add("Step", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"name": { "type": "string" },
		"action": { "type": "string", "enum": ["create", "input", "oauth_redirect"] },
		"input": { "type": "string" },
		"to": { "type": "string" },
		"redirect_uri": { "type": "string" },
		"output": { "$ref": "#/$defs/Output" }
	},
	"allOf": [
        {
            "if": {
                "properties": {
                    "action": { "const": "create" }
                }
            },
            "then": {
                "required": ["input"]
            }
        },
				{
					"if": {
							"properties": {
									"action": { "const": "input" }
							}
					},
					"then": {
							"required": ["input"]
					}
				},
        {
            "if": {
                "properties": {
                    "action": { "const": "oauth_redirect" }
                }
            },
            "then": {
                "required": ["to", "redirect_uri"]
            }
        }
    ]
}
`)

type Step struct {
	Name   string     `yaml:"name"`
	Action StepAction `yaml:"action"`

	// `action` == "create" or "input"
	Input string `yaml:"input"`

	// `action` == "oauth_redirect"
	To          string `yaml:"to"`
	RedirectURI string `yaml:"redirect_uri"`

	Output *Output `yaml:"output"`
}

type StepAction string

const (
	StepActionCreate        StepAction = "create"
	StepActionInput         StepAction = "input"
	StepActionOAuthRedirect StepAction = "oauth_redirect"
)

var _ = TestCaseSchema.Add("Output", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"result": { "type": "string" },
		"error": { "type": "string" }
	}
}
`)

type Output struct {
	Result string `yaml:"result"`
	Error  string `yaml:"error"`
}

var _ = TestCaseSchema.Add("StepResult", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"result": { "type": "string" },
		"error": { "type": "string" }
	}
}
`)

type StepResult struct {
	Result interface{} `json:"result"`
	Error  error       `json:"error"`
}
