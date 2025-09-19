package config

var _ = Schema.Add("SearchConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"implementation": {
			"type": "string",
			"enum": ["elasticsearch", "postgresql", "none"]
		}
	}
}
`)

type SearchImplementation string

const (
	SearchImplementationDefault       SearchImplementation = ""
	SearchImplementationElasticsearch SearchImplementation = "elasticsearch"
	SearchImplementationPostgresql    SearchImplementation = "postgresql"
	SearchImplementationNone          SearchImplementation = "none"
)

type SearchConfig struct {
	Implementation SearchImplementation `json:"implementation,omitempty"`
}

func (c *SearchConfig) GetImplementation(globalImpl GlobalSearchImplementation) SearchImplementation {
	switch c.Implementation {
	case SearchImplementationElasticsearch:
		fallthrough
	case SearchImplementationNone:
		fallthrough
	case SearchImplementationPostgresql:
		return c.Implementation
	default:
		return SearchImplementation(globalImpl)
	}
}
