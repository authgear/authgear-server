package pgsearch

import (
	"context"
	"encoding/json"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/authgear/authgear-server/pkg/api/model"
	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/searchdb"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/lib/pq"
)

func NewStore(
	appID config.AppID,
	sqlBuilder *searchdb.SQLBuilder,
	sqlExecutor *searchdb.SQLExecutor,
) *Store {
	return &Store{
		AppID:       appID,
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
}

type Store struct {
	AppID       config.AppID
	SQLBuilder  *searchdb.SQLBuilder
	SQLExecutor *searchdb.SQLExecutor
}

func (s *Store) QueryUser(
	ctx context.Context,
	searchKeyword string,
	sortOption user.SortOption,
	pageArgs graphqlutil.PageArgs) ([]apimodel.PageItemRef, error) {
	var refs []apimodel.PageItemRef
	q := s.searchQuery(searchKeyword)
	q = sortOption.Apply(q, string(pageArgs.After))

	if pageArgs.First != nil {
		q = q.Limit(*pageArgs.First)
	}

	rows, err := s.SQLExecutor.QueryWith(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var createdAt *string
		var lastLoginAt *string
		err := rows.Scan(&id, &createdAt, &lastLoginAt)
		if err != nil {
			return nil, err
		}
		var cursor string
		switch sortOption.GetSortBy() {
		case user.SortByCreatedAt:
			cursor = *createdAt
		case user.SortByLastLoginAt:
			cursor = *lastLoginAt
		default:
			panic("pgsearch: unknown user cursor column")

		}
		ref := apimodel.PageItemRef{
			ID:     id,
			Cursor: apimodel.PageCursor(cursor),
		}
		refs = append(refs, ref)
	}

	return refs, nil
}

func (s *Store) DeleteUser(ctx context.Context, appID string, userID string) error {
	q := s.SQLBuilder.WithAppID(appID).
		Delete(s.SQLBuilder.TableName("_search_user")).
		Where("id = ?", userID)
	_, err := s.SQLExecutor.ExecWith(ctx, q)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) UpsertUser(ctx context.Context, user *model.SearchUserSource) error {
	// details is for free text search
	details := map[string]string{
		"email_text":                        strings.Join(user.EmailText, " "),
		"email_local_part_text":             strings.Join(user.EmailLocalPartText, " "),
		"email_domain_text":                 strings.Join(user.EmailDomainText, " "),
		"preferred_username_text":           strings.Join(user.PreferredUsernameText, " "),
		"phone_number_text":                 strings.Join(user.PhoneNumberText, " "),
		"phone_number_national_number_text": strings.Join(user.PhoneNumberNationalNumberText, " "),
		"oauth_subject_id_text":             strings.Join(user.OAuthSubjectIDText, " "),
		"family_name":                       user.FamilyName,
		"given_name":                        user.GivenName,
		"middle_name":                       user.MiddleName,
		"name":                              user.Name,
		"nickname":                          user.Nickname,
		"formatted":                         user.Formatted,
		"street_address":                    user.StreetAddress,
		"locality":                          user.Locality,
		"region":                            user.Region,
	}

	defailsBytes, err := json.Marshal(details)
	if err != nil {
		return err
	}

	nonNilArray := func(arr []string) []string {
		newArray := []string{}
		if arr != nil {
			newArray = arr
		}
		return newArray
	}

	toSingleElementArray := func(el string) []string {
		arr := []string{}
		if el != "" {
			arr = append(arr, el)
		}
		return arr
	}

	q := s.SQLBuilder.WithoutAppID().
		Insert(s.SQLBuilder.TableName("_search_user")).
		Columns(
			"id",
			"app_id",
			"created_at",
			"updated_at",
			"last_login_at",
			"is_disabled",
			"emails",
			"email_local_parts",
			"email_domains",
			"preferred_usernames",
			"phone_numbers",
			"phone_number_country_codes",
			"phone_number_national_numbers",
			"oauth_subject_ids",
			"gender",
			"zoneinfo",
			"locale",
			"postal_code",
			"country",
			"details",
		).
		Values(
			user.ID,
			user.AppID,
			user.CreatedAt,
			user.UpdatedAt,
			user.LastLoginAt,
			user.IsDisabled,
			pq.Array(nonNilArray(user.Email)),
			pq.Array(nonNilArray(user.EmailLocalPart)),
			pq.Array(nonNilArray(user.EmailDomain)),
			pq.Array(nonNilArray(user.PreferredUsername)),
			pq.Array(nonNilArray(user.PhoneNumber)),
			pq.Array(nonNilArray(user.PhoneNumberCountryCode)),
			pq.Array(nonNilArray(user.PhoneNumberNationalNumber)),
			pq.Array(nonNilArray(user.OAuthSubjectID)),
			pq.Array(toSingleElementArray(user.Gender)),
			pq.Array(toSingleElementArray(user.Zoneinfo)),
			pq.Array(toSingleElementArray(user.Locale)),
			pq.Array(toSingleElementArray(user.PostalCode)),
			pq.Array(toSingleElementArray(user.Country)),
			defailsBytes,
		).Suffix(`ON CONFLICT (id) DO UPDATE SET
		updated_at = EXCLUDED.updated_at,
		last_login_at = EXCLUDED.last_login_at,
		is_disabled = EXCLUDED.is_disabled,
		emails = EXCLUDED.emails,
		email_local_parts = EXCLUDED.email_local_parts,
		email_domains = EXCLUDED.email_domains,
		preferred_usernames = EXCLUDED.preferred_usernames,
		phone_numbers = EXCLUDED.phone_numbers,
		phone_number_country_codes = EXCLUDED.phone_number_country_codes,
		phone_number_national_numbers = EXCLUDED.phone_number_national_numbers,
		oauth_subject_ids = EXCLUDED.oauth_subject_ids,
		gender = EXCLUDED.gender,
		zoneinfo = EXCLUDED.zoneinfo,
		locale = EXCLUDED.locale,
		postal_code = EXCLUDED.postal_code,
		country = EXCLUDED.country,
		details = EXCLUDED.details
	`)
	_, err = s.SQLExecutor.ExecWith(ctx, q)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) searchQuery(searchKeyword string) db.SelectBuilder {
	appID := string(s.AppID)
	searchKeywordArr := pq.Array([]string{searchKeyword})
	q := s.SQLBuilder.WithAppID(appID).
		Select(
			"su.id",
			"su.created_at",
			"su.last_login_at",
		).
		From(s.SQLBuilder.TableName("_search_user"), "su").
		Where(sq.And{
			sq.Expr("su.app_ids @> ?", pq.Array([]string{appID})),
			sq.Or{
				sq.Expr("su.emails @> ?", searchKeywordArr),
				sq.Expr("su.email_local_parts @> ?", searchKeywordArr),
				sq.Expr("su.email_domains @> ?", searchKeywordArr),
				sq.Expr("su.preferred_usernames @> ?", searchKeywordArr),
				sq.Expr("su.phone_numbers @> ?", searchKeywordArr),
				sq.Expr("su.phone_number_country_codes @> ?", searchKeywordArr),
				sq.Expr("su.phone_number_national_numbers @> ?", searchKeywordArr),
				sq.Expr("su.oauth_subject_ids @> ?", searchKeywordArr),
				sq.Expr("su.gender @> ?", searchKeywordArr),
				sq.Expr("su.zoneinfo @> ?", searchKeywordArr),
				sq.Expr("su.locale @> ?", searchKeywordArr),
				sq.Expr("su.postal_code @> ?", searchKeywordArr),
				sq.Expr("su.country @> ?", searchKeywordArr),
				sq.Expr("su.details_tsvector @@ websearch_to_tsquery(?)", searchKeyword),
			},
		})
	return q
}
