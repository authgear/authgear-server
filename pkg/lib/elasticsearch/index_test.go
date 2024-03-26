package elasticsearch

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	libuser "github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestMakeSearchBody(t *testing.T) {
	appID := "APP_ID"

	test := func(searchKeyword string, filterOptions libuser.FilterOptions, sortOption libuser.SortOption, expected string) {
		val := MakeSearchBody(config.AppID(appID), searchKeyword, filterOptions, sortOption)
		bytes, err := json.Marshal(val)
		So(err, ShouldBeNil)
		So(string(bytes), ShouldEqualJSON, expected)
	}

	Convey("QueryUserOptions.SearchBody short keyword", t, func() {
		test("SH", libuser.FilterOptions{}, libuser.SortOption{}, `
		{
			"query": {
				"bool": {
					"minimum_should_match": 1,
					"filter": [
					{
						"term": {
							"app_id": "APP_ID"
						}
					}
					],
					"should": [
					{
						"term": {
							"id": "SH"
						}
					},
					{
						"term": {
							"email": {
								"value": "SH",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"email_local_part": {
								"value": "SH",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"email_domain": {
								"value": "SH",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"preferred_username": {
								"value": "SH",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"phone_number": {
								"value": "SH",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"phone_number_country_code": {
								"value": "SH",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"phone_number_national_number": {
								"value": "SH",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"oauth_subject_id": {
								"value": "SH",
								"case_insensitive": true
							}
						}
					},
					{
						"match": {
							"family_name": {
								"query": "SH"
							}
						}
					},
					{
						"match": {
							"given_name": {
								"query": "SH"
							}
						}
					},
					{
						"match": {
							"middle_name": {
								"query": "SH"
							}
						}
					},
					{
						"match": {
							"name": {
								"query": "SH"
							}
						}
					},
					{
						"match": {
							"nickname": {
								"query": "SH"
							}
						}
					},
					{
						"match": {
							"formatted": {
								"query": "SH"
							}
						}
					},
					{
						"match": {
							"street_address": {
								"query": "SH"
							}
						}
					},
					{
						"match": {
							"locality": {
								"query": "SH"
							}
						}
					},
					{
						"match": {
							"region": {
								"query": "SH"
							}
						}
					},
					{
						"term": {
							"gender": {
								"case_insensitive": true,
								"value": "SH"
							}
						}
					},
					{
						"term": {
							"zoneinfo": {
								"case_insensitive": true,
								"value": "SH"
							}
						}
					},
					{
						"term": {
							"locale": {
								"case_insensitive": true,
								"value": "SH"
							}
						}
					},
					{
						"term": {
							"country": {
								"case_insensitive": true,
								"value": "SH"
							}
						}
					},
					{
						"term": {
							"postal_code": {
								"case_insensitive": true,
								"value": "SH"
							}
						}
					},
					{
						"term": {
							"role_key": {
								"case_insensitive": true,
								"value": "SH"
							}
						}
					},
					{
						"match": {
							"role_name": {
								"query": "SH"
							}
						}
					},
					{
						"term": {
							"group_key": {
								"case_insensitive": true,
								"value": "SH"
							}
						}
					},
					{
						"match": {
							"group_name": {
								"query": "SH"
							}
						}
					}
					]
				}
			},
			"sort": [
			{ "created_at": { "order": "desc" } }
			]
		}
		`)
	})

	Convey("QueryUserOptions.SearchBody keyword only", t, func() {
		test("KEYWORD", libuser.FilterOptions{}, libuser.SortOption{}, `
		{
			"query": {
				"bool": {
					"minimum_should_match": 1,
					"filter": [
					{
						"term": {
							"app_id": "APP_ID"
						}
					}
					],
					"should": [
					{
						"term": {
							"id": "KEYWORD"
						}
					},
					{
						"term": {
							"email": {
								"value": "KEYWORD",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"email_local_part": {
								"value": "KEYWORD",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"email_domain": {
								"value": "KEYWORD",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"preferred_username": {
								"value": "KEYWORD",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"phone_number": {
								"value": "KEYWORD",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"phone_number_country_code": {
								"value": "KEYWORD",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"phone_number_national_number": {
								"value": "KEYWORD",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"oauth_subject_id": {
								"value": "KEYWORD",
								"case_insensitive": true
							}
						}
					},
					{
						"match": {
							"family_name": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"given_name": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"middle_name": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"name": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"nickname": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"formatted": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"street_address": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"locality": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"region": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"term": {
							"gender": {
								"case_insensitive": true,
								"value": "KEYWORD"
							}
						}
					},
					{
						"term": {
							"zoneinfo": {
								"case_insensitive": true,
								"value": "KEYWORD"
							}
						}
					},
					{
						"term": {
							"locale": {
								"case_insensitive": true,
								"value": "KEYWORD"
							}
						}
					},
					{
						"term": {
							"country": {
								"case_insensitive": true,
								"value": "KEYWORD"
							}
						}
					},
					{
						"term": {
							"postal_code": {
								"case_insensitive": true,
								"value": "KEYWORD"
							}
						}
					},
					{
						"term": {
							"role_key": {
								"case_insensitive": true,
								"value": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"role_name": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"term": {
							"group_key": {
								"case_insensitive": true,
								"value": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"group_name": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"prefix": {
							"email_text": {
								"value": "KEYWORD",
								"case_insensitive": true,
								"rewrite": "constant_score_boolean"
							}
						}
					},
					{
						"prefix": {
							"email_local_part_text": {
								"value": "KEYWORD",
								"case_insensitive": true,
								"rewrite": "constant_score_boolean"
							}
						}
					},
					{
						"prefix": {
							"email_domain_text": {
								"value": "KEYWORD",
								"case_insensitive": true,
								"rewrite": "constant_score_boolean"
							}
						}
					},
					{
						"prefix": {
							"preferred_username_text": {
								"value": "KEYWORD",
								"case_insensitive": true,
								"rewrite": "constant_score_boolean"
							}
						}
					},
					{
						"prefix": {
							"phone_number_text": {
								"value": "KEYWORD",
								"case_insensitive": true,
								"rewrite": "constant_score_boolean"
							}
						}
					},
					{
						"prefix": {
							"phone_number_national_number_text": {
								"value": "KEYWORD",
								"case_insensitive": true,
								"rewrite": "constant_score_boolean"
							}
						}
					},
					{
						"prefix": {
							"oauth_subject_id_text": {
								"value": "KEYWORD",
								"case_insensitive": true,
								"rewrite": "constant_score_boolean"
							}
						}
					}
					]
				}
			},
			"sort": [
			{ "created_at": { "order": "desc" } }
			]
		}
		`)
	})

	Convey("QueryUserOptions.SearchBody keyword with regexp characters", t, func() {
		test("example.com", libuser.FilterOptions{}, libuser.SortOption{}, `
		{
			"query": {
				"bool": {
					"minimum_should_match": 1,
					"filter": [
					{
						"term": {
							"app_id": "APP_ID"
						}
					}
					],
					"should": [
					{
						"term": {
							"id": "example.com"
						}
					},
					{
						"term": {
							"email": {
								"value": "example.com",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"email_local_part": {
								"value": "example.com",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"email_domain": {
								"value": "example.com",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"preferred_username": {
								"value": "example.com",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"phone_number": {
								"value": "example.com",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"phone_number_country_code": {
								"value": "example.com",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"phone_number_national_number": {
								"value": "example.com",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"oauth_subject_id": {
								"value": "example.com",
								"case_insensitive": true
							}
						}
					},
					{
						"match": {
							"family_name": {
								"query": "example.com"
							}
						}
					},
					{
						"match": {
							"given_name": {
								"query": "example.com"
							}
						}
					},
					{
						"match": {
							"middle_name": {
								"query": "example.com"
							}
						}
					},
					{
						"match": {
							"name": {
								"query": "example.com"
							}
						}
					},
					{
						"match": {
							"nickname": {
								"query": "example.com"
							}
						}
					},
					{
						"match": {
							"formatted": {
								"query": "example.com"
							}
						}
					},
					{
						"match": {
							"street_address": {
								"query": "example.com"
							}
						}
					},
					{
						"match": {
							"locality": {
								"query": "example.com"
							}
						}
					},
					{
						"match": {
							"region": {
								"query": "example.com"
							}
						}
					},
					{
						"term": {
							"gender": {
								"case_insensitive": true,
								"value": "example.com"
							}
						}
					},
					{
						"term": {
							"zoneinfo": {
								"case_insensitive": true,
								"value": "example.com"
							}
						}
					},
					{
						"term": {
							"locale": {
								"case_insensitive": true,
								"value": "example.com"
							}
						}
					},
					{
						"term": {
							"country": {
								"case_insensitive": true,
								"value": "example.com"
							}
						}
					},
					{
						"term": {
							"postal_code": {
								"case_insensitive": true,
								"value": "example.com"
							}
						}
					},
					{
						"term": {
							"role_key": {
								"case_insensitive": true,
								"value": "example.com"
							}
						}
					},
					{
						"match": {
							"role_name": {
								"query": "example.com"
							}
						}
					},
					{
						"term": {
							"group_key": {
								"case_insensitive": true,
								"value": "example.com"
							}
						}
					},
					{
						"match": {
							"group_name": {
								"query": "example.com"
							}
						}
					},
					{
						"prefix": {
							"email_text": {
								"value": "example.com",
								"case_insensitive": true,
								"rewrite": "constant_score_boolean"
							}
						}
					},
					{
						"prefix": {
							"email_local_part_text": {
								"value": "example.com",
								"case_insensitive": true,
								"rewrite": "constant_score_boolean"
							}
						}
					},
					{
						"prefix": {
							"email_domain_text": {
								"value": "example.com",
								"case_insensitive": true,
								"rewrite": "constant_score_boolean"
							}
						}
					},
					{
						"prefix": {
							"preferred_username_text": {
								"value": "example.com",
								"case_insensitive": true,
								"rewrite": "constant_score_boolean"
							}
						}
					},
					{
						"prefix": {
							"phone_number_text": {
								"value": "example.com",
								"case_insensitive": true,
								"rewrite": "constant_score_boolean"
							}
						}
					},
					{
						"prefix": {
							"phone_number_national_number_text": {
								"value": "example.com",
								"case_insensitive": true,
								"rewrite": "constant_score_boolean"
							}
						}
					},
					{
						"prefix": {
							"oauth_subject_id_text": {
								"value": "example.com",
								"case_insensitive": true,
								"rewrite": "constant_score_boolean"
							}
						}
					}
					]
				}
			},
			"sort": [
			{ "created_at": { "order": "desc" } }
			]
		}
		`)
	})

	Convey("QueryUserOptions.SearchBody sort by created_at", t, func() {
		test("KEYWORD", libuser.FilterOptions{}, libuser.SortOption{SortBy: libuser.SortByCreatedAt}, `
		{
			"query": {
				"bool": {
					"minimum_should_match": 1,
					"filter": [
					{
						"term": {
							"app_id": "APP_ID"
						}
					}
					],
					"should": [
					{
						"term": {
							"id": "KEYWORD"
						}
					},
					{
						"term": {
							"email": {
								"value": "KEYWORD",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"email_local_part": {
								"value": "KEYWORD",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"email_domain": {
								"value": "KEYWORD",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"preferred_username": {
								"value": "KEYWORD",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"phone_number": {
								"value": "KEYWORD",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"phone_number_country_code": {
								"value": "KEYWORD",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"phone_number_national_number": {
								"value": "KEYWORD",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"oauth_subject_id": {
								"value": "KEYWORD",
								"case_insensitive": true
							}
						}
					},
					{
						"match": {
							"family_name": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"given_name": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"middle_name": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"name": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"nickname": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"formatted": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"street_address": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"locality": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"region": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"term": {
							"gender": {
								"case_insensitive": true,
								"value": "KEYWORD"
							}
						}
					},
					{
						"term": {
							"zoneinfo": {
								"case_insensitive": true,
								"value": "KEYWORD"
							}
						}
					},
					{
						"term": {
							"locale": {
								"case_insensitive": true,
								"value": "KEYWORD"
							}
						}
					},
					{
						"term": {
							"country": {
								"case_insensitive": true,
								"value": "KEYWORD"
							}
						}
					},
					{
						"term": {
							"postal_code": {
								"case_insensitive": true,
								"value": "KEYWORD"
							}
						}
					},
					{
						"term": {
							"role_key": {
								"case_insensitive": true,
								"value": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"role_name": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"term": {
							"group_key": {
								"case_insensitive": true,
								"value": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"group_name": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"prefix": {
							"email_text": {
								"value": "KEYWORD",
								"case_insensitive": true,
								"rewrite": "constant_score_boolean"
							}
						}
					},
					{
						"prefix": {
							"email_local_part_text": {
								"value": "KEYWORD",
								"case_insensitive": true,
								"rewrite": "constant_score_boolean"
							}
						}
					},
					{
						"prefix": {
							"email_domain_text": {
								"value": "KEYWORD",
								"case_insensitive": true,
								"rewrite": "constant_score_boolean"
							}
						}
					},
					{
						"prefix": {
							"preferred_username_text": {
								"value": "KEYWORD",
								"case_insensitive": true,
								"rewrite": "constant_score_boolean"
							}
						}
					},
					{
						"prefix": {
							"phone_number_text": {
								"value": "KEYWORD",
								"case_insensitive": true,
								"rewrite": "constant_score_boolean"
							}
						}
					},
					{
						"prefix": {
							"phone_number_national_number_text": {
								"value": "KEYWORD",
								"case_insensitive": true,
								"rewrite": "constant_score_boolean"
							}
						}
					},
					{
						"prefix": {
							"oauth_subject_id_text": {
								"value": "KEYWORD",
								"case_insensitive": true,
								"rewrite": "constant_score_boolean"
							}
						}
					}
					]
				}
			},
			"sort": [
			{ "created_at": { "order": "desc" } }
			]
		}
		`)
	})

	Convey("QueryUserOptions.SearchBody sort by last_login_at", t, func() {
		test("KEYWORD", libuser.FilterOptions{}, libuser.SortOption{SortBy: libuser.SortByLastLoginAt}, `
		{
			"query": {
				"bool": {
					"minimum_should_match": 1,
					"filter": [
					{
						"term": {
							"app_id": "APP_ID"
						}
					}
					],
					"should": [
					{
						"term": {
							"id": "KEYWORD"
						}
					},
					{
						"term": {
							"email": {
								"value": "KEYWORD",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"email_local_part": {
								"value": "KEYWORD",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"email_domain": {
								"value": "KEYWORD",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"preferred_username": {
								"value": "KEYWORD",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"phone_number": {
								"value": "KEYWORD",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"phone_number_country_code": {
								"value": "KEYWORD",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"phone_number_national_number": {
								"value": "KEYWORD",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"oauth_subject_id": {
								"value": "KEYWORD",
								"case_insensitive": true
							}
						}
					},
					{
						"match": {
							"family_name": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"given_name": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"middle_name": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"name": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"nickname": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"formatted": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"street_address": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"locality": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"region": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"term": {
							"gender": {
								"case_insensitive": true,
								"value": "KEYWORD"
							}
						}
					},
					{
						"term": {
							"zoneinfo": {
								"case_insensitive": true,
								"value": "KEYWORD"
							}
						}
					},
					{
						"term": {
							"locale": {
								"case_insensitive": true,
								"value": "KEYWORD"
							}
						}
					},
					{
						"term": {
							"country": {
								"case_insensitive": true,
								"value": "KEYWORD"
							}
						}
					},
					{
						"term": {
							"postal_code": {
								"case_insensitive": true,
								"value": "KEYWORD"
							}
						}
					},
					{
						"term": {
							"role_key": {
								"case_insensitive": true,
								"value": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"role_name": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"term": {
							"group_key": {
								"case_insensitive": true,
								"value": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"group_name": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"prefix": {
							"email_text": {
								"value": "KEYWORD",
								"case_insensitive": true,
								"rewrite": "constant_score_boolean"
							}
						}
					},
					{
						"prefix": {
							"email_local_part_text": {
								"value": "KEYWORD",
								"case_insensitive": true,
								"rewrite": "constant_score_boolean"
							}
						}
					},
					{
						"prefix": {
							"email_domain_text": {
								"value": "KEYWORD",
								"case_insensitive": true,
								"rewrite": "constant_score_boolean"
							}
						}
					},
					{
						"prefix": {
							"preferred_username_text": {
								"value": "KEYWORD",
								"case_insensitive": true,
								"rewrite": "constant_score_boolean"
							}
						}
					},
					{
						"prefix": {
							"phone_number_text": {
								"value": "KEYWORD",
								"case_insensitive": true,
								"rewrite": "constant_score_boolean"
							}
						}
					},
					{
						"prefix": {
							"phone_number_national_number_text": {
								"value": "KEYWORD",
								"case_insensitive": true,
								"rewrite": "constant_score_boolean"
							}
						}
					},
					{
						"prefix": {
							"oauth_subject_id_text": {
								"value": "KEYWORD",
								"case_insensitive": true,
								"rewrite": "constant_score_boolean"
							}
						}
					}
					]
				}
			},
			"sort": [
			{ "last_login_at": { "order": "desc" } }
			]
		}
		`)
	})

	Convey("QueryUserOptions.SearchBody sort asc", t, func() {
		test("KEYWORD", libuser.FilterOptions{}, libuser.SortOption{
			SortBy:        libuser.SortByCreatedAt,
			SortDirection: model.SortDirectionAsc,
		}, `
		{
			"query": {
				"bool": {
					"minimum_should_match": 1,
					"filter": [
					{
						"term": {
							"app_id": "APP_ID"
						}
					}
					],
					"should": [
					{
						"term": {
							"id": "KEYWORD"
						}
					},
					{
						"term": {
							"email": {
								"value": "KEYWORD",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"email_local_part": {
								"value": "KEYWORD",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"email_domain": {
								"value": "KEYWORD",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"preferred_username": {
								"value": "KEYWORD",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"phone_number": {
								"value": "KEYWORD",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"phone_number_country_code": {
								"value": "KEYWORD",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"phone_number_national_number": {
								"value": "KEYWORD",
								"case_insensitive": true
							}
						}
					},
					{
						"term": {
							"oauth_subject_id": {
								"value": "KEYWORD",
								"case_insensitive": true
							}
						}
					},
					{
						"match": {
							"family_name": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"given_name": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"middle_name": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"name": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"nickname": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"formatted": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"street_address": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"locality": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"region": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"term": {
							"gender": {
								"case_insensitive": true,
								"value": "KEYWORD"
							}
						}
					},
					{
						"term": {
							"zoneinfo": {
								"case_insensitive": true,
								"value": "KEYWORD"
							}
						}
					},
					{
						"term": {
							"locale": {
								"case_insensitive": true,
								"value": "KEYWORD"
							}
						}
					},
					{
						"term": {
							"country": {
								"case_insensitive": true,
								"value": "KEYWORD"
							}
						}
					},
					{
						"term": {
							"postal_code": {
								"case_insensitive": true,
								"value": "KEYWORD"
							}
						}
					},
					{
						"term": {
							"role_key": {
								"case_insensitive": true,
								"value": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"role_name": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"term": {
							"group_key": {
								"case_insensitive": true,
								"value": "KEYWORD"
							}
						}
					},
					{
						"match": {
							"group_name": {
								"query": "KEYWORD"
							}
						}
					},
					{
						"prefix": {
							"email_text": {
								"value": "KEYWORD",
								"case_insensitive": true,
								"rewrite": "constant_score_boolean"
							}
						}
					},
					{
						"prefix": {
							"email_local_part_text": {
								"value": "KEYWORD",
								"case_insensitive": true,
								"rewrite": "constant_score_boolean"
							}
						}
					},
					{
						"prefix": {
							"email_domain_text": {
								"value": "KEYWORD",
								"case_insensitive": true,
								"rewrite": "constant_score_boolean"
							}
						}
					},
					{
						"prefix": {
							"preferred_username_text": {
								"value": "KEYWORD",
								"case_insensitive": true,
								"rewrite": "constant_score_boolean"
							}
						}
					},
					{
						"prefix": {
							"phone_number_text": {
								"value": "KEYWORD",
								"case_insensitive": true,
								"rewrite": "constant_score_boolean"
							}
						}
					},
					{
						"prefix": {
							"phone_number_national_number_text": {
								"value": "KEYWORD",
								"case_insensitive": true,
								"rewrite": "constant_score_boolean"
							}
						}
					},
					{
						"prefix": {
							"oauth_subject_id_text": {
								"value": "KEYWORD",
								"case_insensitive": true,
								"rewrite": "constant_score_boolean"
							}
						}
					}
					]
				}
			},
			"sort": [
			{ "created_at": { "order": "asc" } }
			]
		}
		`)
	})

	Convey("QueryUserOptions.SearchBody keyword with filter options", t, func() {
		test("KEYWORD", libuser.FilterOptions{
			GroupKeys: []string{"group1", "group2"},
			RoleKeys:  []string{"role1", "role2"},
		}, libuser.SortOption{}, `
		{
			"query": {
				"bool": {
					"minimum_should_match": 1,
					"filter": [
						{
							"term": {
								"app_id": "APP_ID"
							}
						},
						{
							"bool": {
								"minimum_should_match": 1,
								"should": [
									{
										"term": {
											"role_key": {
												"value": "role1"
											}
										}
									},
									{
										"term": {
											"role_key": {
												"value": "role2"
											}
										}
									}
								]
							}
						},
						{
							"bool": {
								"minimum_should_match": 1,
								"should": [
									{
										"term": {
											"group_key": {
												"value": "group1"
											}
										}
									},
									{
										"term": {
											"group_key": {
												"value": "group2"
											}
										}
									}
								]
							}
						}
					],
					"should": [
						{
							"term": {
								"id": "KEYWORD"
							}
						},
						{
							"term": {
								"email": {
									"value": "KEYWORD",
									"case_insensitive": true
								}
							}
						},
						{
							"term": {
								"email_local_part": {
									"value": "KEYWORD",
									"case_insensitive": true
								}
							}
						},
						{
							"term": {
								"email_domain": {
									"value": "KEYWORD",
									"case_insensitive": true
								}
							}
						},
						{
							"term": {
								"preferred_username": {
									"value": "KEYWORD",
									"case_insensitive": true
								}
							}
						},
						{
							"term": {
								"phone_number": {
									"value": "KEYWORD",
									"case_insensitive": true
								}
							}
						},
						{
							"term": {
								"phone_number_country_code": {
									"value": "KEYWORD",
									"case_insensitive": true
								}
							}
						},
						{
							"term": {
								"phone_number_national_number": {
									"value": "KEYWORD",
									"case_insensitive": true
								}
							}
						},
						{
							"term": {
								"oauth_subject_id": {
									"value": "KEYWORD",
									"case_insensitive": true
								}
							}
						},
						{
							"match": {
								"family_name": {
									"query": "KEYWORD"
								}
							}
						},
						{
							"match": {
								"given_name": {
									"query": "KEYWORD"
								}
							}
						},
						{
							"match": {
								"middle_name": {
									"query": "KEYWORD"
								}
							}
						},
						{
							"match": {
								"name": {
									"query": "KEYWORD"
								}
							}
						},
						{
							"match": {
								"nickname": {
									"query": "KEYWORD"
								}
							}
						},
						{
							"match": {
								"formatted": {
									"query": "KEYWORD"
								}
							}
						},
						{
							"match": {
								"street_address": {
									"query": "KEYWORD"
								}
							}
						},
						{
							"match": {
								"locality": {
									"query": "KEYWORD"
								}
							}
						},
						{
							"match": {
								"region": {
									"query": "KEYWORD"
								}
							}
						},
						{
							"term": {
								"gender": {
									"case_insensitive": true,
									"value": "KEYWORD"
								}
							}
						},
						{
							"term": {
								"zoneinfo": {
									"case_insensitive": true,
									"value": "KEYWORD"
								}
							}
						},
						{
							"term": {
								"locale": {
									"case_insensitive": true,
									"value": "KEYWORD"
								}
							}
						},
						{
							"term": {
								"country": {
									"case_insensitive": true,
									"value": "KEYWORD"
								}
							}
						},
						{
							"term": {
								"postal_code": {
									"case_insensitive": true,
									"value": "KEYWORD"
								}
							}
						},
						{
							"term": {
								"role_key": {
									"case_insensitive": true,
									"value": "KEYWORD"
								}
							}
						},
						{
							"match": {
								"role_name": {
									"query": "KEYWORD"
								}
							}
						},
						{
							"term": {
								"group_key": {
									"case_insensitive": true,
									"value": "KEYWORD"
								}
							}
						},
						{
							"match": {
								"group_name": {
									"query": "KEYWORD"
								}
							}
						},
						{
							"prefix": {
								"email_text": {
									"value": "KEYWORD",
									"case_insensitive": true,
									"rewrite": "constant_score_boolean"
								}
							}
						},
						{
							"prefix": {
								"email_local_part_text": {
									"value": "KEYWORD",
									"case_insensitive": true,
									"rewrite": "constant_score_boolean"
								}
							}
						},
						{
							"prefix": {
								"email_domain_text": {
									"value": "KEYWORD",
									"case_insensitive": true,
									"rewrite": "constant_score_boolean"
								}
							}
						},
						{
							"prefix": {
								"preferred_username_text": {
									"value": "KEYWORD",
									"case_insensitive": true,
									"rewrite": "constant_score_boolean"
								}
							}
						},
						{
							"prefix": {
								"phone_number_text": {
									"value": "KEYWORD",
									"case_insensitive": true,
									"rewrite": "constant_score_boolean"
								}
							}
						},
						{
							"prefix": {
								"phone_number_national_number_text": {
									"value": "KEYWORD",
									"case_insensitive": true,
									"rewrite": "constant_score_boolean"
								}
							}
						},
						{
							"prefix": {
								"oauth_subject_id_text": {
									"value": "KEYWORD",
									"case_insensitive": true,
									"rewrite": "constant_score_boolean"
								}
							}
						}
					]
				}
			},
			"sort": [{ "created_at": { "order": "desc" } }]
		}		
		`)
	})

	Convey("QueryUserOptions.SearchBody filter without search keyword", t, func() {
		test("", libuser.FilterOptions{
			GroupKeys: []string{"group1", "group2"},
			RoleKeys:  []string{"role1", "role2"},
		}, libuser.SortOption{}, `
		{
			"query": {
				"bool": {
					"minimum_should_match": 1,
					"filter": [
						{
							"term": {
								"app_id": "APP_ID"
							}
						},
						{
							"bool": {
								"minimum_should_match": 1,
								"should": [
									{
										"term": {
											"role_key": {
												"value": "role1"
											}
										}
									},
									{
										"term": {
											"role_key": {
												"value": "role2"
											}
										}
									}
								]
							}
						},
						{
							"bool": {
								"minimum_should_match": 1,
								"should": [
									{
										"term": {
											"group_key": {
												"value": "group1"
											}
										}
									},
									{
										"term": {
											"group_key": {
												"value": "group2"
											}
										}
									}
								]
							}
						}
					],
					"should": [
						{
							"match_all": {}
						}
					]
				}
			},
			"sort": [{ "created_at": { "order": "desc" } }]
		}
		`)

		test("", libuser.FilterOptions{
			RoleKeys: []string{"role1", "role2"},
		}, libuser.SortOption{}, `
		{
			"query": {
				"bool": {
					"minimum_should_match": 1,
					"filter": [
						{
							"term": {
								"app_id": "APP_ID"
							}
						},
						{
							"bool": {
								"minimum_should_match": 1,
								"should": [
									{
										"term": {
											"role_key": {
												"value": "role1"
											}
										}
									},
									{
										"term": {
											"role_key": {
												"value": "role2"
											}
										}
									}
								]
							}
						}
					],
					"should": [
						{
							"match_all": {}
						}
					]
				}
			},
			"sort": [{ "created_at": { "order": "desc" } }]
		}
		`)

		test("", libuser.FilterOptions{
			GroupKeys: []string{"group1", "group2"},
		}, libuser.SortOption{}, `
		{
			"query": {
				"bool": {
					"minimum_should_match": 1,
					"filter": [
						{
							"term": {
								"app_id": "APP_ID"
							}
						},
						{
							"bool": {
								"minimum_should_match": 1,
								"should": [
									{
										"term": {
											"group_key": {
												"value": "group1"
											}
										}
									},
									{
										"term": {
											"group_key": {
												"value": "group2"
											}
										}
									}
								]
							}
						}
					],
					"should": [
						{
							"match_all": {}
						}
					]
				}
			},
			"sort": [{ "created_at": { "order": "desc" } }]
		}		
		`)
	})
}
