package elasticsearch

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	. "github.com/authgear/authgear-server/pkg/util/testing"
)

func TestQueryUserOptionsSearchBody(t *testing.T) {
	appID := "APP_ID"

	test := func(input *QueryUserOptions, expected string) {
		val := input.SearchBody(appID)
		bytes, err := json.Marshal(val)
		So(err, ShouldBeNil)
		So(bytes, ShouldEqualJSON, expected)
	}

	Convey("QueryUserOptions.SearchBody keyword only", t, func() {
		test(&QueryUserOptions{
			SearchKeyword: "KEYWORD",
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
					}
					]
				}
			},
			"sort": [
			"_score"
			]
		}
		`)
	})

	Convey("QueryUserOptions.SearchBody keyword with regexp characters", t, func() {
		test(&QueryUserOptions{
			SearchKeyword: "example.com",
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
					}
					]
				}
			},
			"sort": [
			"_score"
			]
		}
		`)
	})

	Convey("QueryUserOptions.SearchBody sort by created_at", t, func() {
		test(&QueryUserOptions{
			SearchKeyword: "KEYWORD",
			SortBy:        QueryUserSortByCreatedAt,
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
		test(&QueryUserOptions{
			SearchKeyword: "KEYWORD",
			SortBy:        QueryUserSortByLastLoginAt,
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
		test(&QueryUserOptions{
			SearchKeyword: "KEYWORD",
			SortBy:        QueryUserSortByCreatedAt,
			SortDirection: SortDirectionAsc,
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
}
