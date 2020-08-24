package db

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
)

var InvalidCursor = apierrors.Invalid.WithReason("InvalidCursor")
var ErrInvalidCursor = InvalidCursor.New("invalid pagination cursor")

type PageKey struct {
	Key string `json:"key"`
	ID  string `json:"id"`
}

func NewFromPageCursor(k model.PageCursor) (*PageKey, error) {
	if k == "" {
		return nil, nil
	}

	data, err := base64.RawURLEncoding.DecodeString(string(k))
	if err != nil {
		return nil, ErrInvalidCursor
	}

	var pageKey PageKey
	if err := json.Unmarshal(data, &pageKey); err != nil {
		return nil, ErrInvalidCursor
	}

	return &pageKey, nil
}

func (k *PageKey) ToPageCursor() (model.PageCursor, error) {
	cursor, err := json.Marshal(k)
	if err != nil {
		return "", err
	}
	return model.PageCursor(base64.RawURLEncoding.EncodeToString(cursor)), nil
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
