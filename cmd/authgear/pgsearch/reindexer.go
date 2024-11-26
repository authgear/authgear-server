package pgsearch

import (
	"context"
	"log"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/searchdb"
	"github.com/authgear/authgear-server/pkg/lib/search/pgsearch"
	"github.com/authgear/authgear-server/pkg/lib/search/reindex"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

type Reindexer struct {
	Clock          clock.Clock
	AppDBHandle    *appdb.Handle
	SearchDBHandle *searchdb.Handle
	AppID          config.AppID
	UserStore      *user.Store
	PGSearchStore  *pgsearch.Store

	SourceProvider *reindex.SourceProvider
}

func (q *Reindexer) Reindex(ctx context.Context) (err error) {

	allUserIDs, err := q.reindex(ctx)
	if err != nil {
		return
	}

	deletedCount, err := q.cleanupDeletedUsers(ctx, allUserIDs)
	if err != nil {
		return err
	}

	log.Printf("App (%v): %v indexed; %v deleted;\n", q.AppID, len(allUserIDs), deletedCount)
	return nil
}

func (q *Reindexer) reindex(ctx context.Context) (allUserIDs map[string]struct{}, err error) {
	allUserIDs = make(map[string]struct{})

	var first uint64 = 500
	var after model.PageCursor = ""
	var items []reindex.ReindexItem
	var count = 0

	for {
		err = q.AppDBHandle.WithTx(ctx, func(ctx context.Context) (err error) {
			items, err = q.SourceProvider.QueryPage(ctx, after, first)
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
			count += 1
			source := item.Value
			sources = append(sources, source)
			allUserIDs[source.ID] = struct{}{}

			log.Printf("App (%v): processing user %v;\n", q.AppID, count)
		}

		err := q.SearchDBHandle.WithTx(ctx, func(ctx context.Context) error {
			return q.PGSearchStore.UpsertUsers(ctx, sources)
		})
		if err != nil {
			return nil, err
		}

		userIDs := slice.Map(sources, func(source *model.SearchUserSource) string { return source.ID })

		err = q.AppDBHandle.WithTx(ctx, func(ctx context.Context) error {
			return q.UserStore.UpdateLastIndexedAt(ctx, userIDs, q.Clock.NowUTC())
		})
		if err != nil {
			return nil, err
		}
	}

	return allUserIDs, nil
}

func (q *Reindexer) cleanupDeletedUsers(ctx context.Context, allUserIDs map[string]struct{}) (deletedCount int64, err error) {
	allUserIDsSlice := []string{}
	for id := range allUserIDs {
		allUserIDsSlice = append(allUserIDsSlice, id)
	}

	err = q.SearchDBHandle.WithTx(ctx, func(ctx context.Context) error {
		deletedCount, err = q.PGSearchStore.CleanupUsers(ctx, string(q.AppID), allUserIDsSlice)
		return err
	})
	return deletedCount, err
}
