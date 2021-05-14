package elasticsearch

import (
	"bytes"
	"fmt"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"

	libes "github.com/authgear/authgear-server/pkg/lib/elasticsearch"
)

func CreateIndex(es *elasticsearch.Client) error {
	// index_prefixes is only available on text.
	// Therefore we have to store both keyword and text.
	bodyStr := `
	{
		"mappings": {
			"properties": {
				"app_id": { "type": "keyword" },
				"id": { "type": "keyword" },
				"created_at": { "type": "date" },
				"updated_at": { "type": "date" },
				"last_login_at": { "type": "date" },
				"is_disabled": { "type": "boolean" },
				"email": { "type": "keyword" },
				"email_text": {
					"type": "text",
					"index_prefixes": {
						"min_chars": 3,
						"max_chars": 19
					}
				},
				"email_local_part": { "type": "keyword" },
				"email_local_part_text": {
					"type": "text",
					"index_prefixes": {
						"min_chars": 3,
						"max_chars": 10
					}
				},
				"email_domain": { "type": "keyword" },
				"email_domain_text": {
					"type": "text",
					"index_prefixes": {
						"min_chars": 3,
						"max_chars": 10
					}
				},
				"preferred_username": { "type": "keyword" },
				"preferred_username_text": {
					"type": "text",
					"index_prefixes": {
						"min_chars": 3,
						"max_chars": 19
					}
				},
				"phone_number": { "type": "keyword" },
				"phone_number_text": {
					"type": "text",
					"index_prefixes": {
						"min_chars": 3,
						"max_chars": 19
					}
				},
				"phone_number_country_code": { "type": "keyword" },
				"phone_number_national_number": { "type": "keyword" },
				"phone_number_national_number_text": {
					"type": "text",
					"index_prefixes": {
						"min_chars": 3,
						"max_chars": 15
					}
				}
			}
		}
	}
	`
	res, err := es.Indices.Create(libes.IndexNameUser, func(o *esapi.IndicesCreateRequest) {
		o.Body = bytes.NewReader([]byte(bodyStr))
	})
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("%v", res)
	}
	return nil
}

func DeleteIndex(es *elasticsearch.Client) error {
	res, err := es.Indices.Delete([]string{libes.IndexNameUser})
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("%v", res)
	}
	return nil
}
