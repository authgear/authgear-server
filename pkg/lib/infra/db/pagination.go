package db

import (
	"encoding/base64"
	"strconv"
	"strings"

	sq "github.com/Masterminds/squirrel"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
)

var InvalidCursor = apierrors.Invalid.WithReason("InvalidCursor")
var ErrInvalidCursor = InvalidCursor.New("invalid pagination cursor")

type PageKey struct {
	Offset uint64
}

const pageCursorPrefix = "offset:"

func NewFromPageCursor(k model.PageCursor) (*PageKey, error) {
	if k == "" {
		return nil, nil
	}

	data, err := base64.RawURLEncoding.DecodeString(string(k))
	if err != nil {
		return nil, ErrInvalidCursor
	}

	if !strings.HasPrefix(string(data), pageCursorPrefix) {
		return nil, ErrInvalidCursor
	}
	cursor := strings.TrimPrefix(string(data), pageCursorPrefix)

	offset, err := strconv.ParseUint(cursor, 10, 64)
	if err != nil {
		return nil, ErrInvalidCursor
	}

	return &PageKey{Offset: offset}, nil
}

func (k *PageKey) ToPageCursor() (model.PageCursor, error) {
	cursor := pageCursorPrefix + strconv.FormatUint(k.Offset, 10)
	return model.PageCursor(base64.RawURLEncoding.EncodeToString([]byte(cursor))), nil
}

var InvalidQuery = apierrors.BadRequest.WithReason("InvalidQuery")

type PageQuery func(builder SelectBuilder, after, before *PageKey, first, last *uint64) (sq.Sqlizer, uint64, error)

type QueryPageConfig struct {
	KeyColumn string
	IDColumn  string
}

func QueryPage(config QueryPageConfig) PageQuery {
	keyColumn := config.KeyColumn
	idColumn := config.IDColumn

	return func(builder SelectBuilder, after, before *PageKey, first, last *uint64) (sq.Sqlizer, uint64, error) {
		query := builder.builder

		var offset uint64
		switch {
		case after == nil && before == nil && first == nil && last == nil:
			offset = 0

		case after == nil && before == nil && first != nil && last == nil:
			offset = 0
			query = query.Limit(*first)

		case after != nil && before == nil && first != nil && last == nil:
			offset = after.Offset + 1
			query = query.Limit(*first).Offset(offset)

		case after != nil && before != nil:
			offset = after.Offset + 1
			limit := before.Offset - offset
			if first != nil && *first < limit {
				limit = *first
			}
			if last != nil && *last < limit {
				delta := limit - *last
				offset += delta
				limit -= delta
			}
			query = query.Limit(limit).Offset(offset)

		default:
			return nil, 0, InvalidQuery.New("unsupported pagination")
		}
		query = query.OrderBy(keyColumn, idColumn)

		return query, offset, nil
	}
}
