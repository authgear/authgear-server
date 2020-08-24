package db

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

type PageKey struct {
	Key string
	ID  string
}

var InvalidQuery = apierrors.BadRequest.WithReason("InvalidQuery")

type PageQuery func(builder SelectBuilder, after, before *PageKey, first, last *uint64) (sq.Sqlizer, error)

type QueryPageConfig struct {
	KeyColumn string
	IDColumn  string
}

func QueryPage(config QueryPageConfig) PageQuery {
	keyColumn := config.KeyColumn
	idColumn := config.IDColumn

	return func(builder SelectBuilder, after, before *PageKey, first, last *uint64) (sq.Sqlizer, error) {
		query := builder.builder

		if after != nil {
			query = query.Where(fmt.Sprintf("%s > ? OR (%s = ? AND %s > ?)", keyColumn, keyColumn, idColumn), after.Key, after.Key, after.ID)
		}
		if before != nil {
			query = query.Where(fmt.Sprintf("%s < ? OR (%s = ? AND %s < ?)", keyColumn, keyColumn, idColumn), before.Key, before.Key, before.ID)
		}

		switch {
		case first != nil && last != nil:
			// NOTE: Relay spec discourage using first & last simultaneously,
			// and implementing this requires complex SQL. Therefore, forbid this
			// combination for now.
			return nil, InvalidQuery.New("first & last is mutually exclusive")

		case last != nil:
			query = newSQLBuilder().
				Select("page.*").
				FromSelect(
					query.
						OrderBy(keyColumn+" DESC", idColumn+" DESC").
						Limit(*last),
					"page",
				).
				OrderBy(keyColumn, idColumn)

		case first != nil:
			query = query.
				OrderBy(keyColumn, idColumn).
				Limit(*first)

		default:
			query = query.
				OrderBy(keyColumn, idColumn)
		}

		return query, nil
	}
}
