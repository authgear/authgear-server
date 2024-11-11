package redisqueue

import (
	"context"
	"encoding/json"

	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/elasticsearch"
	"github.com/authgear/authgear-server/pkg/lib/infra/redisqueue"
)

func UserReindex(ctx context.Context, appProvider *deps.AppProvider, task *redisqueue.Task) (output json.RawMessage, err error) {
	esService := newElasticsearchService(ctx, appProvider)
	var request elasticsearch.ReindexRequest
	err = json.Unmarshal(task.Input, &request)
	if err != nil {
		return
	}
	result := esService.ExecReindexUser(ctx, request)
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return
	}

	output = json.RawMessage(resultBytes)
	return
}
