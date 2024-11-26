package redisqueue

import (
	"context"
	"encoding/json"

	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/infra/redisqueue"
	"github.com/authgear/authgear-server/pkg/lib/search/reindex"
)

func UserReindex(ctx context.Context, appProvider *deps.AppProvider, task *redisqueue.Task) (output json.RawMessage, err error) {
	reindexer := newSearchReindexer(ctx, appProvider)
	var request reindex.ReindexRequest
	err = json.Unmarshal(task.Input, &request)
	if err != nil {
		return
	}
	result := reindexer.ExecReindexUser(ctx, request)
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return
	}

	output = json.RawMessage(resultBytes)
	return
}
