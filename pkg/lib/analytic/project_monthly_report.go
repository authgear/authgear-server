package analytic

import (
	"errors"
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
)

type ProjectMonthlyReportOptions struct {
	Year  int
	Month int
}

type ProjectMonthlyReport struct {
	GlobalHandle  *globaldb.Handle
	GlobalDBStore *GlobalDBStore
	AuditDBHandle *auditdb.ReadHandle
	AuditDBStore  *AuditDBReadStore
}

func (r *ProjectMonthlyReport) Run(options *ProjectMonthlyReportOptions) (data *ReportData, err error) {
	firstDayOfMonth := time.Date(options.Year, time.Month(options.Month), 1, 0, 0, 0, 0, time.UTC)

	var appIDs []string
	if err = r.GlobalHandle.WithTx(func() error {
		appIDs, err = r.GlobalDBStore.GetAppIDs()
		if err != nil {
			return fmt.Errorf("failed to fetch app ids: %w", err)
		}
		return nil
	}); err != nil {
		return
	}

	values := make([][]interface{}, len(appIDs))
	for i, appID := range appIDs {
		monthlyActiveUserCount := 0
		err = r.AuditDBHandle.ReadOnly(func() (e error) {
			c, err := r.AuditDBStore.GetAnalyticCountByType(
				appID,
				MonthlyActiveUserCountType,
				&firstDayOfMonth,
			)
			if err != nil {
				if !errors.Is(err, ErrAnalyticCountNotFound) {
					return fmt.Errorf("failed to fetch monthly active user: %w", err)
				}
			} else {
				monthlyActiveUserCount = c.Count
			}
			return nil
		})
		if err != nil {
			return
		}

		values[i] = []interface{}{
			options.Year,
			options.Month,
			appID,
			monthlyActiveUserCount,
		}
	}

	data = &ReportData{
		Header: []interface{}{
			"Year",
			"Month",
			"Project ID",
			"Monthly active user",
		},
		Values: values,
	}

	return data, nil
}
