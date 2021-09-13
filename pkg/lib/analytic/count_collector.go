package analytic

import (
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
)

type CountCollector struct {
	GlobalHandle  *globaldb.Handle
	GlobalDBStore *GlobalDBStore
	AppDBHandle   *appdb.Handle
	AppDBStore    *AppDBStore
	AuditDBHandle *auditdb.WriteHandle
	AuditDBStore  *AuditDBStore
}

func (c *CountCollector) CollectDaily(date *time.Time) (updatedCount int, err error) {
	utc := date.UTC()
	rangeFrom := time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
	rangeTo := rangeFrom.AddDate(0, 0, 1)

	var appIDs []string
	err = c.GlobalHandle.WithTx(func() error {
		appIDs, err = c.GlobalDBStore.GetAppIDs()
		if err != nil {
			return fmt.Errorf("failed to fetch app ids: %w", err)
		}
		return nil
	})
	if err != nil {
		return
	}

	var counts []*Count
	for _, appID := range appIDs {
		appCounts, e := c.CollectDailyCountForApp(appID, rangeFrom, rangeTo)
		if e != nil {
			err = e
			return
		}
		counts = append(counts, appCounts...)
	}

	if len(counts) > 0 {
		err = c.AuditDBHandle.WithTx(func() error {
			// Store the counts to audit db
			err = c.AuditDBStore.CreateCounts(counts)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return 0, fmt.Errorf("failed to store count: %w", err)
		}
	}

	updatedCount = len(counts)
	return updatedCount, nil
}

func (c *CountCollector) CollectDailyCountForApp(appID string, date time.Time, nextDay time.Time) (counts []*Count, err error) {
	// Cumulative number of user count
	err = c.AppDBHandle.WithTx(func() error {
		userCount, err := c.AppDBStore.GetUserCountBeforeTime(appID, &nextDay)
		if err != nil {
			return err
		}
		if userCount == 0 {
			// no user in the app, skip the cumulative number of user count
			return nil
		}
		counts = append(counts, NewCount(
			appID,
			userCount,
			date,
			CumulativeUserCountType,
		))
		return nil
	})
	if err != nil {
		err = fmt.Errorf("failed to calculate cumulative number of user: %w", err)
		return
	}

	return
}
