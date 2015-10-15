package fs

import (
	"github.com/oursky/skygear/skydb"
)

type queryMatcher skydb.Query

func (q *queryMatcher) match(record *skydb.Record) bool {
	// currently, fs implement only matches record type
	return record.ID.Type == q.Type
}
