package elasticsearch

import (
	"bytes"
	"fmt"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"

	libes "github.com/authgear/authgear-server/pkg/lib/elasticsearch"
)

// DO NOT delete or update properties
// Only add new properties
var IndexMappings = `
{
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
			"analyzer": "keyword",
			"index_prefixes": {
				"min_chars": 3,
				"max_chars": 19
			}
		},
		"email_local_part": { "type": "keyword" },
		"email_local_part_text": {
			"type": "text",
			"analyzer": "keyword",
			"index_prefixes": {
				"min_chars": 3,
				"max_chars": 10
			}
		},
		"email_domain": { "type": "keyword" },
		"email_domain_text": {
			"type": "text",
			"analyzer": "keyword",
			"index_prefixes": {
				"min_chars": 3,
				"max_chars": 10
			}
		},
		"preferred_username": { "type": "keyword" },
		"preferred_username_text": {
			"type": "text",
			"analyzer": "keyword",
			"index_prefixes": {
				"min_chars": 3,
				"max_chars": 19
			}
		},
		"phone_number": { "type": "keyword" },
		"phone_number_text": {
			"type": "text",
			"analyzer": "keyword",
			"index_prefixes": {
				"min_chars": 3,
				"max_chars": 19
			}
		},
		"phone_number_country_code": { "type": "keyword" },
		"phone_number_national_number": { "type": "keyword" },
		"phone_number_national_number_text": {
			"type": "text",
			"analyzer": "keyword",
			"index_prefixes": {
				"min_chars": 3,
				"max_chars": 15
			}
		},
		"oauth_subject_id": { "type": "keyword" },
		"oauth_subject_id_text": {
			"type": "text",
			"analyzer": "keyword",
			"index_prefixes": {
				"min_chars": 3,
				"max_chars": 19
			}
		},
		"family_name": { "type": "text" },
		"given_name": { "type": "text" },
		"middle_name": { "type": "text" },
		"name": { "type": "text" },
		"nickname": { "type": "text" },
		"formatted": { "type": "text" },
		"street_address": { "type": "text" },
		"locality": { "type": "text" },
		"region": { "type": "text" },
		"gender": { "type": "keyword" },
		"zoneinfo": { "type": "keyword" },
		"locale": { "type": "keyword" },
		"country": { "type": "keyword" },
		"postal_code": { "type": "keyword" },
		"role_key": { "type": "keyword" },
		"role_name": { "type": "text" },
		"group_key": { "type": "keyword" },
		"group_name": { "type": "text" }
	}
}
`

func CreateIndex(es *elasticsearch.Client) error {
	// index_prefixes is only available on text.
	// Therefore we have to store both keyword and text.
	// Note that we have to specify the analyzer as "keyword"
	// because we want elasticsearch to treat the whole field as a term.

	// If we even need to adjust index_prefixes.min_chars,
	// we have to make sure that PrefixMinChars in pkg/lib/elasticsearch is also updated.

	// The default analyzer "standard" breaks email address into 3 parts (the local part, the @, and the domain part).
	// For example, suppose the standard attribute "name" of UserA is "usera@example.com".
	// Suppose we are searching for "userb@example.com".
	// The result will have 2 hits, one for UserA, one for UserB.
	// UserA is a hit because the search keyword is being broken into 3 parts,
	// and the last part matches the "name" field.
	//
	// To prevent this, we want to prevent the search query from being broken into 3 parts,
	// and instead treat the whole query as a term.
	//
	// uax_url_email is standard tokenizer except it treats URL and email address as single tokens. See https://www.elastic.co/guide/en/elasticsearch/reference/current/analysis-uaxurlemail-tokenizer.html
	//
	// The official way to customize the standard analyzer is documented at https://www.elastic.co/guide/en/elasticsearch/reference/current/analysis-standard-analyzer.html#_definition_4
	//
	// Our approach is to create a new analyzer that is equivalent to the standard analyzer with URL and email address treated as single tokens.
	// In the index, we create an analyzer named "default_search".
	// This name has a special meaning. See https://www.elastic.co/guide/en/elasticsearch/reference/current/specify-analyzer.html#specify-search-default-analyzer
	// Note that the standard analyzer is still being used to analyze the field.
	// So if the query search is "example.com", it can still match the "name" field in the above example.
	// By applying the analyzer to the search query instead of the field,
	// we keep the characteristics of "more specific search query gives more specific search result".
	bodyStr := fmt.Sprintf(`
	{
		"settings": {
			"analysis": {
				"analyzer": {
					"default_search": {
						"type": "custom",
						"tokenizer": "uax_url_email",
						"filter": ["lowercase"]
					}
				}
			}
		},
		"mappings": %s
		}
	}
	`, IndexMappings)
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

func UpdateIndex(es *elasticsearch.Client) error {
	bodyStr := IndexMappings
	res, err := es.Indices.PutMapping(bytes.NewReader([]byte(bodyStr)), func(o *esapi.IndicesPutMappingRequest) {
		o.Index = []string{libes.IndexNameUser}
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
