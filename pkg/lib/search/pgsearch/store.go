package pgsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/api/model"
	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/searchdb"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
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
	filters user.FilterOptions,
	sortOption user.SortOption,
	pageArgs graphqlutil.PageArgs) ([]apimodel.PageItemRef, error) {
	var refs []apimodel.PageItemRef
	q := s.searchQuery(searchKeyword, filters)
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
func (s *Store) CleanupUsers(ctx context.Context, appID string, keepUserIDs []string) (int64, error) {
	q := s.SQLBuilder.WithAppID(appID).
		Delete(s.SQLBuilder.TableName("_search_user")).
		Where("NOT id = ANY (?)", pq.Array(keepUserIDs))

	r, err := s.SQLExecutor.ExecWith(ctx, q)
	if err != nil {
		return 0, err
	}
	return r.RowsAffected()
}

func (s *Store) UpsertUsers(ctx context.Context, users []*model.SearchUserSource) error {
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
			"role_keys",
			"group_keys",
			"details",
		)

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

	for _, user := range users {
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
			"group_names":                       strings.Join(user.GroupName, " "),
			"role_names":                        strings.Join(user.RoleName, " "),
		}

		defailsBytes, err := json.Marshal(details)
		if err != nil {
			return err
		}

		q = q.
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
				pq.Array(nonNilArray(user.RoleKey)),
				pq.Array(nonNilArray(user.GroupKey)),
				defailsBytes,
			)
	}

	q = q.Suffix(`ON CONFLICT (id) DO UPDATE SET
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
		role_keys = EXCLUDED.role_keys,
		group_keys = EXCLUDED.group_keys,
		details = EXCLUDED.details
	`)
	_, err := s.SQLExecutor.ExecWith(ctx, q)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) searchQuery(searchKeyword string, filters user.FilterOptions) db.SelectBuilder {
	appID := string(s.AppID)
	unisegSearchKeyword := StringUnicodeSegmentation(searchKeyword)
	searchKeywordArr := pq.Array([]string{searchKeyword})

	ands := sq.And{
		sq.Expr("su.app_ids @> ?", pq.Array([]string{appID})),
	}

	searchKeywordOrs := sq.Or{}
	if searchKeyword != "" {
		searchKeywordOrs = append(searchKeywordOrs,
			sq.Expr("su.id = ?", searchKeyword),
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
			sq.Expr("su.details_tsvector @@ websearch_to_tsquery(?)", unisegSearchKeyword),
		)
	}

	if len(searchKeyword) >= 3 {
		// Only add prefix search if >= 3 characters were inputted to avoid matching too many rows
		prefixSearchQuery := fmt.Sprintf("'%s':*", strings.ReplaceAll(searchKeyword, "'", "''"))
		searchKeywordOrs = append(searchKeywordOrs,
			sq.Expr("su.details_tsvector @@ to_tsquery(?)", prefixSearchQuery))
	}

	if len(searchKeywordOrs) > 0 {
		ands = append(ands, searchKeywordOrs)
	}

	if filters.IsFilterEnabled() {
		if len(filters.GroupKeys) > 0 {
			ands = append(ands,
				sq.Expr("su.group_keys @> ?", pq.Array(filters.GroupKeys)))
		}
		if len(filters.RoleKeys) > 0 {
			ands = append(ands,
				sq.Expr("su.role_keys @> ?", pq.Array(filters.RoleKeys)))
		}
	}

	q := s.SQLBuilder.WithAppID(appID).
		Select(
			"su.id",
			"su.created_at",
			"su.last_login_at",
		).
		From(s.SQLBuilder.TableName("_search_user"), "su").
		Where(ands)
	return q
}
