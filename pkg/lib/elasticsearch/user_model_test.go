package elasticsearch

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	libuser "github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	. "github.com/authgear/authgear-server/pkg/util/testing"
)

func TestMakeSearchBody(t *testing.T) {
	appID := "APP_ID"

	test := func(searchKeyword string, sortOption libuser.SortOption, expected string) {
		val := MakeSearchBody(config.AppID(appID), searchKeyword, sortOption)
		bytes, err := json.Marshal(val)
		So(err, ShouldBeNil)
		So(bytes, ShouldEqualJSON, expected)
	}

	Convey("QueryUserOptions.SearchBody keyword only", t, func() {
		test("KEYWORD", libuser.SortOption{}, `
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
		test("example.com", libuser.SortOption{}, `
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
		test("KEYWORD", libuser.SortOption{SortBy: libuser.SortByCreatedAt}, `
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
		test("KEYWORD", libuser.SortOption{SortBy: libuser.SortByLastLoginAt}, `
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
		test("KEYWORD", libuser.SortOption{
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
