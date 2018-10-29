package handler

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

var timeNow = func() time.Time { return time.Now().UTC() }

type serializedError struct {
	id  string
	err skyerr.Error
}

func newSerializedError(id string, err skyerr.Error) serializedError {
	return serializedError{
		id:  id,
		err: err,
	}
}

func (s serializedError) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"_type":   "error",
		"name":    s.err.Name(),
		"code":    s.err.Code(),
		"message": s.err.Message(),
	}
	if s.id != "" {
		m["_id"] = s.id

		ss := strings.SplitN(s.id, "/", 2)
		if len(ss) == 2 {
			m["_recordType"] = ss[0]
			m["_recordID"] = ss[1]
		}
	}
	if s.err.Info() != nil {
		m["info"] = s.err.Info()
	}

	return json.Marshal(m)
}
