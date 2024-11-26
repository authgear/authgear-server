package elasticsearch

import (
	libuser "github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

const IndexNameUser = "user"

const PrefixMinChars = 3

func makeSearchConditions(searchKeyword string) []interface{} {
	should := []interface{}{
		map[string]interface{}{
			"term": map[string]interface{}{
				"id": searchKeyword,
			},
		},
		map[string]interface{}{
			"term": map[string]interface{}{
				"email": map[string]interface{}{
					"value":            searchKeyword,
					"case_insensitive": true,
				},
			},
		},
		map[string]interface{}{
			"term": map[string]interface{}{
				"email_local_part": map[string]interface{}{
					"value":            searchKeyword,
					"case_insensitive": true,
				},
			},
		},
		map[string]interface{}{
			"term": map[string]interface{}{
				"email_domain": map[string]interface{}{
					"value":            searchKeyword,
					"case_insensitive": true,
				},
			},
		},
		map[string]interface{}{
			"term": map[string]interface{}{
				"preferred_username": map[string]interface{}{
					"value":            searchKeyword,
					"case_insensitive": true,
				},
			},
		},
		map[string]interface{}{
			"term": map[string]interface{}{
				"phone_number": map[string]interface{}{
					"value":            searchKeyword,
					"case_insensitive": true,
				},
			},
		},
		map[string]interface{}{
			"term": map[string]interface{}{
				"phone_number_country_code": map[string]interface{}{
					"value":            searchKeyword,
					"case_insensitive": true,
				},
			},
		},
		map[string]interface{}{
			"term": map[string]interface{}{
				"phone_number_national_number": map[string]interface{}{
					"value":            searchKeyword,
					"case_insensitive": true,
				},
			},
		},
		map[string]interface{}{
			"term": map[string]interface{}{
				"oauth_subject_id": map[string]interface{}{
					"value":            searchKeyword,
					"case_insensitive": true,
				},
			},
		},
		// standard attributes
		map[string]interface{}{
			"match": map[string]interface{}{
				"family_name": map[string]interface{}{
					"query": searchKeyword,
				},
			},
		},
		map[string]interface{}{
			"match": map[string]interface{}{
				"given_name": map[string]interface{}{
					"query": searchKeyword,
				},
			},
		},
		map[string]interface{}{
			"match": map[string]interface{}{
				"middle_name": map[string]interface{}{
					"query": searchKeyword,
				},
			},
		},
		map[string]interface{}{
			"match": map[string]interface{}{
				"name": map[string]interface{}{
					"query": searchKeyword,
				},
			},
		},
		map[string]interface{}{
			"match": map[string]interface{}{
				"nickname": map[string]interface{}{
					"query": searchKeyword,
				},
			},
		},
		map[string]interface{}{
			"match": map[string]interface{}{
				"formatted": map[string]interface{}{
					"query": searchKeyword,
				},
			},
		},
		map[string]interface{}{
			"match": map[string]interface{}{
				"street_address": map[string]interface{}{
					"query": searchKeyword,
				},
			},
		},
		map[string]interface{}{
			"match": map[string]interface{}{
				"locality": map[string]interface{}{
					"query": searchKeyword,
				},
			},
		},
		map[string]interface{}{
			"match": map[string]interface{}{
				"region": map[string]interface{}{
					"query": searchKeyword,
				},
			},
		},
		map[string]interface{}{
			"term": map[string]interface{}{
				"gender": map[string]interface{}{
					"value":            searchKeyword,
					"case_insensitive": true,
				},
			},
		},
		map[string]interface{}{
			"term": map[string]interface{}{
				"zoneinfo": map[string]interface{}{
					"value":            searchKeyword,
					"case_insensitive": true,
				},
			},
		},
		map[string]interface{}{
			"term": map[string]interface{}{
				"locale": map[string]interface{}{
					"value":            searchKeyword,
					"case_insensitive": true,
				},
			},
		},
		map[string]interface{}{
			"term": map[string]interface{}{
				"country": map[string]interface{}{
					"value":            searchKeyword,
					"case_insensitive": true,
				},
			},
		},
		map[string]interface{}{
			"term": map[string]interface{}{
				"postal_code": map[string]interface{}{
					"value":            searchKeyword,
					"case_insensitive": true,
				},
			},
		},
		// Roles and Groups
		map[string]interface{}{
			"term": map[string]interface{}{
				"role_key": map[string]interface{}{
					"value":            searchKeyword,
					"case_insensitive": true,
				},
			},
		},
		map[string]interface{}{
			"match": map[string]interface{}{
				"role_name": map[string]interface{}{
					"query": searchKeyword,
				},
			},
		},
		map[string]interface{}{
			"term": map[string]interface{}{
				"group_key": map[string]interface{}{
					"value":            searchKeyword,
					"case_insensitive": true,
				},
			},
		},
		map[string]interface{}{
			"match": map[string]interface{}{
				"group_name": map[string]interface{}{
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
		should = append(should, []interface{}{
			map[string]interface{}{
				"prefix": map[string]interface{}{
					"email_text": map[string]interface{}{
						"value":            searchKeyword,
						"case_insensitive": true,
						"rewrite":          "constant_score_boolean",
					},
				},
			},
			map[string]interface{}{
				"prefix": map[string]interface{}{
					"email_local_part_text": map[string]interface{}{
						"value":            searchKeyword,
						"case_insensitive": true,
						"rewrite":          "constant_score_boolean",
					},
				},
			},
			map[string]interface{}{
				"prefix": map[string]interface{}{
					"email_domain_text": map[string]interface{}{
						"value":            searchKeyword,
						"case_insensitive": true,
						"rewrite":          "constant_score_boolean",
					},
				},
			},
			map[string]interface{}{
				"prefix": map[string]interface{}{
					"preferred_username_text": map[string]interface{}{
						"value":            searchKeyword,
						"case_insensitive": true,
						"rewrite":          "constant_score_boolean",
					},
				},
			},
			map[string]interface{}{
				"prefix": map[string]interface{}{
					"phone_number_text": map[string]interface{}{
						"value":            searchKeyword,
						"case_insensitive": true,
						"rewrite":          "constant_score_boolean",
					},
				},
			},
			map[string]interface{}{
				"prefix": map[string]interface{}{
					"phone_number_national_number_text": map[string]interface{}{
						"value":            searchKeyword,
						"case_insensitive": true,
						"rewrite":          "constant_score_boolean",
					},
				},
			},
			map[string]interface{}{
				"prefix": map[string]interface{}{
					"oauth_subject_id_text": map[string]interface{}{
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
	filterOptions libuser.FilterOptions) []interface{} {

	filters := []interface{}{}

	if len(filterOptions.RoleKeys) > 0 {
		roleKeyShoulds := slice.Map(filterOptions.RoleKeys, func(roleKey string) interface{} {
			return map[string]interface{}{
				"term": map[string]interface{}{
					"role_key": map[string]interface{}{
						"value": roleKey,
					},
				},
			}
		})
		filters = append(filters, map[string]interface{}{
			"bool": map[string]interface{}{
				"minimum_should_match": 1,
				"should":               roleKeyShoulds,
			},
		})
	}

	if len(filterOptions.GroupKeys) > 0 {
		groupKeyShoulds := slice.Map(filterOptions.GroupKeys, func(groupKey string) interface{} {
			return map[string]interface{}{
				"term": map[string]interface{}{
					"group_key": map[string]interface{}{
						"value": groupKey,
					},
				},
			}
		})
		filters = append(filters, map[string]interface{}{
			"bool": map[string]interface{}{
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
) map[string]interface{} {
	var should []interface{}
	if searchKeyword != "" {
		should = makeSearchConditions(searchKeyword)
	} else {
		should = []interface{}{map[string]interface{}{
			"match_all": map[string]interface{}{},
		}}
	}

	filter := []interface{}{
		map[string]interface{}{
			"term": map[string]interface{}{
				"app_id": appID,
			},
		},
	}

	if filterOptions.IsFilterEnabled() {

		filter = append(filter, makeFilterConditions(filterOptions)...)
	}

	body := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
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

	var sort []interface{}
	sort = append(sort, map[string]interface{}{
		string(sortBy): map[string]interface{}{
			"order": dir,
		},
	})
	body["sort"] = sort

	return body
}
