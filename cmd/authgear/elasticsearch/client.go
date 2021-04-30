package elasticsearch

import (
	"github.com/elastic/go-elasticsearch/v7"
)

func MakeClient(elasticsearchURL string) (*elasticsearch.Client, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{
			elasticsearchURL,
		},
	}
	return elasticsearch.NewClient(cfg)
}
