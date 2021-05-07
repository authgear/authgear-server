package elasticsearch

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	libuser "github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

const IndexNameUser = "user"

type User struct {
	ID          string     `json:"id,omitempty"`
	AppID       string     `json:"app_id,omitempty"`
	CreatedAt   time.Time  `json:"created_at,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at,omitempty"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	IsDisabled  bool       `json:"is_disabled"`

	Email          []string `json:"email,omitempty"`
	EmailLocalPart []string `json:"email_local_part,omitempty"`
	EmailDomain    []string `json:"email_domain,omitempty"`

	PreferredUsername []string `json:"preferred_username,omitempty"`

	PhoneNumber               []string `json:"phone_number,omitempty"`
	PhoneNumberCountryCode    []string `json:"phone_number_country_code,omitempty"`
	PhoneNumberNationalNumber []string `json:"phone_number_national_number,omitempty"`
}

func MakeSearchBody(
	appID config.AppID,
	searchKeyword string,
	sortOption libuser.SortOption,
) interface{} {
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
				"should": []interface{}{
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
				},
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
