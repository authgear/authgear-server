package config

var _ = Schema.Add("SearchConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"implementation": {
			"type": "string",
			"enum": ["elasticsearch", "postgresql"]
		}
	}
}
`)

type SearchImplementation string

const (
	SearchImplementationDefault       SearchImplementation = ""
	SearchImplementationElasticsearch SearchImplementation = "elasticsearch"
	SearchImplementationPostgresql    SearchImplementation = "postgresql"
)

type SearchConfig struct {
	Implementation SearchImplementation `json:"implementation,omitempty"`
}

func (c *SearchConfig) GetImplementation() SearchImplementation {
	switch c.Implementation {
	case SearchImplementationElasticsearch:
		fallthrough
	case SearchImplementationPostgresql:
		return c.Implementation
	default:
		return SearchImplementationElasticsearch
	}
}
