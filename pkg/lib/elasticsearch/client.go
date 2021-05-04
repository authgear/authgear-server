package elasticsearch

import (
	"fmt"

	"github.com/elastic/go-elasticsearch/v7"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func NewClient(credentials *config.ElasticsearchCredentials) *elasticsearch.Client {
	cfg := elasticsearch.Config{
		Addresses: []string{
			credentials.ElasticsearchURL,
		},
	}
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		panic(fmt.Errorf("failed to create elasticsearch client: %w", err))
	}
	return client
}
