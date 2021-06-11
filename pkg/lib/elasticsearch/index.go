package elasticsearch

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	libuser "github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

const IndexNameUser = "user"

const PrefixMinChars = 3

func MakeSearchBody(
	appID config.AppID,
	searchKeyword string,
	sortOption libuser.SortOption,
) interface{} {
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
	}

	// For unknown reason, if the search keyword is shorter than the prefix min chars,
	// elasticsearch will throw runtime exception.
	if len(searchKeyword) >= PrefixMinChars {
		should = append(should, []interface{}{
			map[string]interface{}{
				"prefix": map[string]interface{}{
					"email_text": map[string]interface{}{
						"value":            searchKeyword,
						"case_insensitive": true,
					},
				},
			},
			map[string]interface{}{
				"prefix": map[string]interface{}{
					"email_local_part_text": map[string]interface{}{
						"value":            searchKeyword,
						"case_insensitive": true,
					},
				},
			},
			map[string]interface{}{
				"prefix": map[string]interface{}{
					"email_domain_text": map[string]interface{}{
						"value":            searchKeyword,
						"case_insensitive": true,
					},
				},
			},
			map[string]interface{}{
				"prefix": map[string]interface{}{
					"preferred_username_text": map[string]interface{}{
						"value":            searchKeyword,
						"case_insensitive": true,
					},
				},
			},
			map[string]interface{}{
				"prefix": map[string]interface{}{
					"phone_number_text": map[string]interface{}{
						"value":            searchKeyword,
						"case_insensitive": true,
					},
				},
			},
			map[string]interface{}{
				"prefix": map[string]interface{}{
					"phone_number_national_number_text": map[string]interface{}{
						"value":            searchKeyword,
						"case_insensitive": true,
					},
				},
			},
		}...)
	}

	body := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"minimum_should_match": 1,
				"filter": []interface{}{
					map[string]interface{}{
						"term": map[string]interface{}{
							"app_id": appID,
						},
					},
				},
				"should": should,
			},
		},
	}

	var sort []interface{}
	if sortOption.SortBy == libuser.SortByDefault {
		sort = append(sort, "_score")
	} else {
		dir := sortOption.SortDirection
		if dir == model.SortDirectionDefault {
			dir = model.SortDirectionDesc
		}
		sort = append(sort, map[string]interface{}{
			string(sortOption.SortBy): map[string]interface{}{
				"order": dir,
			},
		})
	}
	body["sort"] = sort

	return body
}
