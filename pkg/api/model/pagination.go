package model

import (
	"encoding/base64"
	"encoding/json"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
)

var InvalidCursor = apierrors.Invalid.WithReason("InvalidCursor")
var ErrInvalidCursor = InvalidCursor.New("invalid pagination cursor")

type pageCursor struct {
	Key string `json:"key"`
	ID  string `json:"id"`
}

type PageCursor string

func NewCursor(key, id string) (PageCursor, error) {
	cursor, err := json.Marshal(pageCursor{
		Key: key,
		ID:  id,
	})
	if err != nil {
		return "", err
	}
	return PageCursor(base64.RawURLEncoding.EncodeToString(cursor)), nil
}

func (k PageCursor) AsDBKey() (*db.PageKey, error) {
	if k == "" {
		return nil, nil
	}

	data, err := base64.RawURLEncoding.DecodeString(string(k))
	if err != nil {
		return nil, ErrInvalidCursor
	}

	var cursor pageCursor
	if err := json.Unmarshal(data, &cursor); err != nil {
		return nil, ErrInvalidCursor
	}

	return &db.PageKey{
		Key: cursor.Key,
		ID:  cursor.ID,
	}, nil
}

type PageItem struct {
	Value  interface{}
	Cursor PageCursor
}
