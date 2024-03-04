package pgsearch

import (
	"log"

	"github.com/authgear/authgear-server/cmd/authgear/search"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/search/pgsearch"
	"github.com/authgear/authgear-server/pkg/lib/search/reindex"
)

type Reindexer struct {
	*search.Reindexer
}

func (q *Reindexer) Reindex(store *pgsearch.Store) (err error) {

	allUserIDs, err := q.reindex(store)
	if err != nil {
		return
	}

	deletedCount, err := q.cleanupDeletedUsers(store, allUserIDs)
	if err != nil {
		return err
	}

	log.Printf("App (%v): %v indexed; %v deleted;\n", q.AppID, len(allUserIDs), deletedCount)
	return nil
}

func (q *Reindexer) reindex(store *pgsearch.Store) (allUserIDs map[string]struct{}, err error) {
	allUserIDs = make(map[string]struct{})

	var first uint64 = 500
	var after model.PageCursor = ""
	var items []search.Item

	for {
		err = q.Handle.WithTx(func() (err error) {
			items, err = q.QueryPage(after, first)
			if err != nil {
				return
			}
			return nil
		})
		if err != nil {
			return
		}

		// Termination condition
		if len(items) <= 0 {
			break
		}

		// Prepare for next iteration
		after = items[len(items)-1].Cursor

		sources := []*model.SearchUserSource{}

		// Process the items
		for _, item := range items {
			user := item.Value.(*model.SearchUserRaw)
			source := reindex.RawToSource(user)
			sources = append(sources, source)
			allUserIDs[user.ID] = struct{}{}
		}

		err := store.UpsertUsers(sources)
		if err != nil {
			return nil, err
		}
	}

	return allUserIDs, nil
}

func (q *Reindexer) cleanupDeletedUsers(store *pgsearch.Store, allUserIDs map[string]struct{}) (int64, error) {
	allUserIDsSlice := []string{}
	for id := range allUserIDs {
		allUserIDsSlice = append(allUserIDsSlice, id)
	}
	return store.CleanupUsers(string(q.AppID), allUserIDsSlice)

}
