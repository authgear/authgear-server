package db

import (
	"encoding/base64"
	"strconv"
	"strings"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
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
	if k == nil {
		return "", nil
	}

	cursor := pageCursorPrefix + strconv.FormatUint(k.Offset, 10)
	return model.PageCursor(base64.RawURLEncoding.EncodeToString([]byte(cursor))), nil
}

var InvalidQuery = apierrors.BadRequest.WithReason("InvalidQuery")

func ApplyPageArgs(builder SelectBuilder, pageArgs graphqlutil.PageArgs) (out SelectBuilder, offset uint64, err error) {
	query := builder.builder

	first := pageArgs.First
	last := pageArgs.Last

	after, err := NewFromPageCursor(model.PageCursor(pageArgs.After))
	if err != nil {
		return
	}
	before, err := NewFromPageCursor(model.PageCursor(pageArgs.Before))
	if err != nil {
		return
	}

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
		err = InvalidQuery.New("unsupported pagination")
		return
	}

	out = builder
	out.builder = query
	return
}
