package redisqueue

import (
	"context"
	"encoding/json"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/infra/redisqueue"
	"github.com/authgear/authgear-server/pkg/lib/userexport"
)

func UserExport(ctx context.Context, appProvider *deps.AppProvider, task *redisqueue.Task) (output json.RawMessage, err error) {
	userExportService := newUserExportService(ctx, appProvider)
	var request userexport.Request
	err = json.Unmarshal(task.Input, &request)
	if err != nil {
		return
	}

	outputFilename, taskError := userExportService.ExportRecords(ctx, &request, task)
	result := userexport.Result{
		Filename: outputFilename,
	}
	if taskError != nil {
		result.Error = apierrors.AsAPIError(taskError)
	}

	resultBytes, err := json.Marshal(result)
	if err != nil {
		return
	}

	output = json.RawMessage(resultBytes)
	return
}
