package config

var _ = Schema.Add("TextTemplate", `
{
	"type": "object",
	"required": ["text_template"],
	"additionalProperties": false,
	"properties": {
		"text_template": {
			"type": "object",
			"required": ["template"],
			"additionalProperties": false,
			"properties": {
				"template": { "type": "string", "format": "x_text_template" }
			}
		}
	}
}
`)

type TextTemplate struct {
	TextTemplate *TextTemplateBody `json:"text_template,omitempty" nullable:"true"`
}

type TextTemplateBody struct {
	Template string `json:"template,omitempty"`
}
