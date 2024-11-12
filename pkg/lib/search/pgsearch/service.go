package pgsearch

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"

	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/searchdb"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type Service struct {
	AppID       *config.AppID
	SQLBuilder  *searchdb.SQLBuilder
	SQLExecutor *searchdb.SQLExecutor
}

func (s *Service) QueryUser(
	ctx context.Context,
	searchKeyword string,
	sortOption user.SortOption,
	pageArgs graphqlutil.PageArgs) ([]apimodel.PageItemRef, error) {
	if s.SQLExecutor == nil {
		return nil, fmt.Errorf("search database credential is not provided")
	}
	var refs []apimodel.PageItemRef
	err := s.SQLExecutor.Database.ReadOnly(ctx, func(ctx context.Context) error {
		q := s.searchQuery(searchKeyword)
		q = sortOption.Apply(q, string(pageArgs.After))

		if pageArgs.First != nil {
			q = q.Limit(*pageArgs.First)
		}

		rows, err := s.SQLExecutor.QueryWith(ctx, q)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var id string
			var createdAt *string
			var lastLoginAt *string
			err := rows.Scan(&id, &createdAt, &lastLoginAt)
			if err != nil {
				return err
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
		return nil
	})
	if err != nil {
		return nil, err
	}

	return refs, nil
}

func (s *Service) searchQuery(searchKeyword string) db.SelectBuilder {
	appID := string(*s.AppID)
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
