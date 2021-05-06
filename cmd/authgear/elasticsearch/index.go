package elasticsearch

import (
	"bytes"
	"fmt"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

var IndexName = "user"

func CreateIndex(es *elasticsearch.Client) error {
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
				"preferred_username": { "type": "keyword" },
				"phone_number": { "type": "keyword" }
			}
		}
	}
	`
	res, err := es.Indices.Create(IndexName, func(o *esapi.IndicesCreateRequest) {
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
	res, err := es.Indices.Delete([]string{IndexName})
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("%v", res)
	}
	return nil
}
