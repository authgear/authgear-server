package elasticsearch

import (
	libuser "github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

const IndexNameUser = "user"

const PrefixMinChars = 3

func makeSearchConditions(searchKeyword string) []any {
	should := []any{
		map[string]any{
			"term": map[string]any{
				"id": searchKeyword,
			},
		},
		map[string]any{
			"term": map[string]any{
				"email": map[string]any{
					"value":            searchKeyword,
					"case_insensitive": true,
				},
			},
		},
		map[string]any{
			"term": map[string]any{
				"email_local_part": map[string]any{
					"value":            searchKeyword,
					"case_insensitive": true,
				},
			},
		},
		map[string]any{
			"term": map[string]any{
				"email_domain": map[string]any{
					"value":            searchKeyword,
					"case_insensitive": true,
				},
			},
		},
		map[string]any{
			"term": map[string]any{
				"preferred_username": map[string]any{
					"value":            searchKeyword,
					"case_insensitive": true,
				},
			},
		},
		map[string]any{
			"term": map[string]any{
				"phone_number": map[string]any{
					"value":            searchKeyword,
					"case_insensitive": true,
				},
			},
		},
		map[string]any{
			"term": map[string]any{
				"phone_number_country_code": map[string]any{
					"value":            searchKeyword,
					"case_insensitive": true,
				},
			},
		},
		map[string]any{
			"term": map[string]any{
				"phone_number_national_number": map[string]any{
					"value":            searchKeyword,
					"case_insensitive": true,
				},
			},
		},
		map[string]any{
			"term": map[string]any{
				"oauth_subject_id": map[string]any{
					"value":            searchKeyword,
					"case_insensitive": true,
				},
			},
		},
		// standard attributes
		map[string]any{
			"match": map[string]any{
				"family_name": map[string]any{
					"query": searchKeyword,
				},
			},
		},
		map[string]any{
			"match": map[string]any{
				"given_name": map[string]any{
					"query": searchKeyword,
				},
			},
		},
		map[string]any{
			"match": map[string]any{
				"middle_name": map[string]any{
					"query": searchKeyword,
				},
			},
		},
		map[string]any{
			"match": map[string]any{
				"name": map[string]any{
					"query": searchKeyword,
				},
			},
		},
		map[string]any{
			"match": map[string]any{
				"nickname": map[string]any{
					"query": searchKeyword,
				},
			},
		},
		map[string]any{
			"match": map[string]any{
				"formatted": map[string]any{
					"query": searchKeyword,
				},
			},
		},
		map[string]any{
			"match": map[string]any{
				"street_address": map[string]any{
					"query": searchKeyword,
				},
			},
		},
		map[string]any{
			"match": map[string]any{
				"locality": map[string]any{
					"query": searchKeyword,
				},
			},
		},
		map[string]any{
			"match": map[string]any{
				"region": map[string]any{
					"query": searchKeyword,
				},
			},
		},
		map[string]any{
			"term": map[string]any{
				"gender": map[string]any{
					"value":            searchKeyword,
					"case_insensitive": true,
				},
			},
		},
		map[string]any{
			"term": map[string]any{
				"zoneinfo": map[string]any{
					"value":            searchKeyword,
					"case_insensitive": true,
				},
			},
		},
		map[string]any{
			"term": map[string]any{
				"locale": map[string]any{
					"value":            searchKeyword,
					"case_insensitive": true,
				},
			},
		},
		map[string]any{
			"term": map[string]any{
				"country": map[string]any{
					"value":            searchKeyword,
					"case_insensitive": true,
				},
			},
		},
		map[string]any{
			"term": map[string]any{
				"postal_code": map[string]any{
					"value":            searchKeyword,
					"case_insensitive": true,
				},
			},
		},
		// Roles and Groups
		map[string]any{
			"term": map[string]any{
				"role_key": map[string]any{
					"value":            searchKeyword,
					"case_insensitive": true,
				},
			},
		},
		map[string]any{
			"match": map[string]any{
				"role_name": map[string]any{
					"query": searchKeyword,
				},
			},
		},
		map[string]any{
			"term": map[string]any{
				"group_key": map[string]any{
					"value":            searchKeyword,
					"case_insensitive": true,
				},
			},
		},
		map[string]any{
			"match": map[string]any{
				"group_name": map[string]any{
					"query": searchKeyword,
				},
			},
		},
	}

	// For unknown reason, if the search keyword is shorter than the prefix min chars,
	// elasticsearch will throw runtime exception.

	// For unknown reason, if the search keyword is mix of Chinese characters and English characters,
	// elasticsearch will throw a runtime exception of
	// Cannot invoke Object.hashCode() because this.rewriteMethod is null.
	// When "rewrite" is set to "constant_score_boolean", this error seems gone.
	if len(searchKeyword) >= PrefixMinChars {
		should = append(should, []any{
			map[string]any{
				"prefix": map[string]any{
					"email_text": map[string]any{
						"value":            searchKeyword,
						"case_insensitive": true,
						"rewrite":          "constant_score_boolean",
					},
				},
			},
			map[string]any{
				"prefix": map[string]any{
					"email_local_part_text": map[string]any{
						"value":            searchKeyword,
						"case_insensitive": true,
						"rewrite":          "constant_score_boolean",
					},
				},
			},
			map[string]any{
				"prefix": map[string]any{
					"email_domain_text": map[string]any{
						"value":            searchKeyword,
						"case_insensitive": true,
						"rewrite":          "constant_score_boolean",
					},
				},
			},
			map[string]any{
				"prefix": map[string]any{
					"preferred_username_text": map[string]any{
						"value":            searchKeyword,
						"case_insensitive": true,
						"rewrite":          "constant_score_boolean",
					},
				},
			},
			map[string]any{
				"prefix": map[string]any{
					"phone_number_text": map[string]any{
						"value":            searchKeyword,
						"case_insensitive": true,
						"rewrite":          "constant_score_boolean",
					},
				},
			},
			map[string]any{
				"prefix": map[string]any{
					"phone_number_national_number_text": map[string]any{
						"value":            searchKeyword,
						"case_insensitive": true,
						"rewrite":          "constant_score_boolean",
					},
				},
			},
			map[string]any{
				"prefix": map[string]any{
					"oauth_subject_id_text": map[string]any{
						"value":            searchKeyword,
						"case_insensitive": true,
						"rewrite":          "constant_score_boolean",
					},
				},
			},
		}...)
	}

	return should
}

func makeFilterConditions(
	filterOptions libuser.FilterOptions) []any {

	filters := []any{}

	if len(filterOptions.RoleKeys) > 0 {
		roleKeyShoulds := slice.Map(filterOptions.RoleKeys, func(roleKey string) any {
			return map[string]any{
				"term": map[string]any{
					"role_key": map[string]any{
						"value": roleKey,
					},
				},
			}
		})
		filters = append(filters, map[string]any{
			"bool": map[string]any{
				"minimum_should_match": 1,
				"should":               roleKeyShoulds,
			},
		})
	}

	if len(filterOptions.GroupKeys) > 0 {
		groupKeyShoulds := slice.Map(filterOptions.GroupKeys, func(groupKey string) any {
			return map[string]any{
				"term": map[string]any{
					"group_key": map[string]any{
						"value": groupKey,
					},
				},
			}
		})
		filters = append(filters, map[string]any{
			"bool": map[string]any{
				"minimum_should_match": 1,
				"should":               groupKeyShoulds,
			},
		})
	}

	return filters
}

func MakeSearchBody(
	appID config.AppID,
	searchKeyword string,
	filterOptions libuser.FilterOptions,
	sortOption libuser.SortOption,
) map[string]any {
	var should []any
	if searchKeyword != "" {
		should = makeSearchConditions(searchKeyword)
	} else {
		should = []any{map[string]any{
			"match_all": map[string]any{},
		}}
	}

	filter := []any{
		map[string]any{
			"term": map[string]any{
				"app_id": appID,
			},
		},
	}

	if filterOptions.IsFilterEnabled() {

		filter = append(filter, makeFilterConditions(filterOptions)...)
	}

	body := map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"minimum_should_match": 1,
				"filter":               filter,
				"should":               should,
			},
		},
	}

	// To order to use search_after, "sort" must appear in the response.
	// To make "sort" appear in the response, we MUST NOT sort by _score.
	sortBy := sortOption.GetSortBy()

	dir := sortOption.GetSortDirection()

	var sort []any
	sort = append(sort, map[string]any{
		string(sortBy): map[string]any{
			"order": dir,
		},
	})
	body["sort"] = sort

	return body
}
