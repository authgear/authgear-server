package elasticsearch

import (
	"encoding/base64"
	"encoding/json"

	"github.com/authgear/authgear-server/pkg/api/model"
)

func CursorToSearchAfter(cursor model.PageCursor) (searchAfter any, err error) {
	if cursor == "" {
		return
	}

	bytes, err := base64.RawURLEncoding.DecodeString(string(cursor))
	if err != nil {
		return
	}

	err = json.Unmarshal(bytes, &searchAfter)
	if err != nil {
		return
	}

	return
}

func SortToCursor(sort any) (cursor model.PageCursor, err error) {
	if sort == nil {
		return
	}

	bytes, err := json.Marshal(sort)
	if err != nil {
		return
	}

	return model.PageCursor(base64.RawURLEncoding.EncodeToString(bytes)), nil
}
