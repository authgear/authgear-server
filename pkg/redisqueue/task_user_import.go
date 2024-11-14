package redisqueue

import (
	"context"
	"encoding/json"

	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/infra/redisqueue"
	"github.com/authgear/authgear-server/pkg/lib/userimport"
)

func UserImport(ctx context.Context, appProvider *deps.AppProvider, task *redisqueue.Task) (output json.RawMessage, err error) {
	userImportService := newUserImportService(ctx, appProvider)
	var request userimport.Request
	err = json.Unmarshal(task.Input, &request)
	if err != nil {
		return
	}
	result := userImportService.ImportRecords(ctx, &request)
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return
	}

	output = json.RawMessage(resultBytes)
	return
}
