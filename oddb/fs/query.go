package fs

import (
	"github.com/oursky/ourd/oddb"
)

type queryMatcher oddb.Query

func (q *queryMatcher) match(record *oddb.Record) bool {
	// currently, fs implement only matches record type
	return record.Type == q.Type
}
