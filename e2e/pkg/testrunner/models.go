package testrunner

type BeforeHook struct {
	Type       BeforeHookType      `json:"type"`
	UserImport string              `json:"user_import"`
	CustomSQL  BeforeHookCustomSQL `json:"custom_sql"`
}

var _ = TestCaseSchema.Add("AuthgearYAMLSource", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"extend": { "type": "string", "description": "Path to the base authgear.yaml" },
		"override": { "type": "string", "description": "Inline snippet to override the base authgear.yaml" }
	}
}
`)

type AuthgearYAMLSource struct {
	Extend   string `json:"extend"`
	Override string `json:"override"`
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
		"path": { "type": "string", "description": "Path to the custom SQL script" }
	},
	"required": ["path"]
}
`)

type BeforeHookCustomSQL struct {
	Path string `json:"path"`
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
		"action": { "type": "string", "enum": ["create", "input", "oauth_redirect", "generate_totp_code", "query"] },
		"input": { "type": "string" },
		"to": { "type": "string" },
		"redirect_uri": { "type": "string" },
		"totp_secret": { "type": "string" },
		"output": { "$ref": "#/$defs/Output" },
		"query": { "type": "string" },
		"query_output": { "$ref": "#/$defs/QueryOutput" }
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
        },
				{
				  "if": {
							"properties": {
									"action": { "const": "generate_totp_code" }
							}
					},
					"then": {
							"required": ["totp_secret"]
					}
				},
				{
				  "if": {
							"properties": {
									"action": { "const": "query" }
							}
					},
					"then": {
							"required": ["query"]
					}
				}
    ]
}
`)

type Step struct {
	Name   string     `json:"name"`
	Action StepAction `json:"action"`

	// `action` == "create" or "input"
	Input string `json:"input"`

	// `action` == "oauth_redirect"
	To          string `json:"to"`
	RedirectURI string `json:"redirect_uri"`

	// `action` == "generate_totp_code"
	TOTPSecret string `json:"totp_secret"`

	// `action` == "input"
	Output *Output `json:"output"`

	// `action` == "query"
	Query       string       `json:"query"`
	QueryOutput *QueryOutput `json:"query_output"`
}

type StepAction string

const (
	StepActionCreate           StepAction = "create"
	StepActionInput            StepAction = "input"
	StepActionOAuthRedirect    StepAction = "oauth_redirect"
	StepActionGenerateTOTPCode StepAction = "generate_totp_code"
	StepActionQuery            StepAction = "query"
)

var _ = TestCaseSchema.Add("QueryOutput", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"rows": { "type": "string" }
	}
}
`)

type QueryOutput struct {
	Rows string `json:"rows"`
}

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
	Result string `json:"result"`
	Error  string `json:"error"`
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
