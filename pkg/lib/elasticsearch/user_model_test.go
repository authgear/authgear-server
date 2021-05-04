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
							"email": "KEYWORD"
						}
					},
					{
						"term": {
							"preferred_username": "KEYWORD"
						}
					},
					{
						"term": {
							"phone_number": "KEYWORD"
						}
					},
					{
						"regexp": {
							"email": {
								"case_insensitive": true,
								"value": "\\@KEYWORD$"
							}
						}
					},
					{
						"regexp": {
							"phone_number": {
								"value": "KEYWORD$"
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
							"email": "example.com"
						}
					},
					{
						"term": {
							"preferred_username": "example.com"
						}
					},
					{
						"term": {
							"phone_number": "example.com"
						}
					},
					{
						"regexp": {
							"email": {
								"case_insensitive": true,
								"value": "\\@example\\.com$"
							}
						}
					},
					{
						"regexp": {
							"phone_number": {
								"value": "example\\.com$"
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
							"email": "KEYWORD"
						}
					},
					{
						"term": {
							"preferred_username": "KEYWORD"
						}
					},
					{
						"term": {
							"phone_number": "KEYWORD"
						}
					},
					{
						"regexp": {
							"email": {
								"case_insensitive": true,
								"value": "\\@KEYWORD$"
							}
						}
					},
					{
						"regexp": {
							"phone_number": {
								"value": "KEYWORD$"
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
							"email": "KEYWORD"
						}
					},
					{
						"term": {
							"preferred_username": "KEYWORD"
						}
					},
					{
						"term": {
							"phone_number": "KEYWORD"
						}
					},
					{
						"regexp": {
							"email": {
								"case_insensitive": true,
								"value": "\\@KEYWORD$"
							}
						}
					},
					{
						"regexp": {
							"phone_number": {
								"value": "KEYWORD$"
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
							"email": "KEYWORD"
						}
					},
					{
						"term": {
							"preferred_username": "KEYWORD"
						}
					},
					{
						"term": {
							"phone_number": "KEYWORD"
						}
					},
					{
						"regexp": {
							"email": {
								"case_insensitive": true,
								"value": "\\@KEYWORD$"
							}
						}
					},
					{
						"regexp": {
							"phone_number": {
								"value": "KEYWORD$"
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
