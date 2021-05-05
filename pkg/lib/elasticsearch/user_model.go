package elasticsearch

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
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

type QueryUserSortBy string

const (
	QueryUserSortByDefault     QueryUserSortBy = ""
	QueryUserSortByCreatedAt   QueryUserSortBy = "created_at"
	QueryUserSortByLastLoginAt QueryUserSortBy = "last_login_at"
)

type SortDirection string

const (
	SortDirectionDefault SortDirection = ""
	SortDirectionAsc     SortDirection = "asc"
	SortDirectionDesc    SortDirection = "desc"
)

type QueryUserOptions struct {
	SearchKeyword string
	First         uint64
	After         model.PageCursor
	SortBy        QueryUserSortBy
	SortDirection SortDirection
}

func (o *QueryUserOptions) SearchBody(appID string) interface{} {
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
							"id": o.SearchKeyword,
						},
					},
					map[string]interface{}{
						"term": map[string]interface{}{
							"email": o.SearchKeyword,
						},
					},
					map[string]interface{}{
						"term": map[string]interface{}{
							"email_local_part": o.SearchKeyword,
						},
					},
					map[string]interface{}{
						"term": map[string]interface{}{
							"email_domain": o.SearchKeyword,
						},
					},
					map[string]interface{}{
						"term": map[string]interface{}{
							"preferred_username": o.SearchKeyword,
						},
					},
					map[string]interface{}{
						"term": map[string]interface{}{
							"phone_number": o.SearchKeyword,
						},
					},
					map[string]interface{}{
						"term": map[string]interface{}{
							"phone_number_country_code": o.SearchKeyword,
						},
					},
					map[string]interface{}{
						"term": map[string]interface{}{
							"phone_number_national_number": o.SearchKeyword,
						},
					},
				},
			},
		},
	}

	var sort []interface{}
	if o.SortBy == QueryUserSortByDefault {
		sort = append(sort, "_score")
	} else {
		dir := o.SortDirection
		if dir == SortDirectionDefault {
			dir = SortDirectionDesc
		}
		sort = append(sort, map[string]interface{}{
			string(o.SortBy): map[string]interface{}{
				"order": dir,
			},
		})
	}
	body["sort"] = sort

	return body
}
