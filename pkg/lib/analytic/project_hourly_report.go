package analytic

import (
	"context"
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
)

type ProjectHourlyReportOptions struct {
	Time *time.Time
}

type ProjectHourlyReport struct {
	GlobalHandle  *globaldb.Handle
	GlobalDBStore *GlobalDBStore
	AppDBHandle   *appdb.Handle
	AppDBStore    *AppDBStore
}

func (r *ProjectHourlyReport) Run(ctx context.Context, options *ProjectHourlyReportOptions) (data *ReportData, err error) {
	var appIDs []string
	if err = r.GlobalHandle.ReadOnly(ctx, func(ctx context.Context) error {
		appIDs, err = r.GlobalDBStore.GetAppIDs(ctx)
		if err != nil {
			return fmt.Errorf("failed to fetch app ids: %w", err)
		}

		return nil
	}); err != nil {
		return
	}

	timeStr := options.Time.Format("2006-01-02T15")

	values := make([][]interface{}, len(appIDs))
	for i, appID := range appIDs {
		var count int
		err = r.AppDBHandle.ReadOnly(ctx, func(ctx context.Context) error {
			count, err = r.AppDBStore.GetUserCountBeforeTime(ctx, appID, options.Time)
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return
		}

		values[i] = []interface{}{
			timeStr,
			appID,
			count,
		}
	}

	data = &ReportData{
		Header: []interface{}{
			"Hour",
			"Project ID",
			"Number of users",
		},
		Values: values,
	}

	return
}
