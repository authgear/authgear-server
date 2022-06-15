package usage

import (
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/util/periodical"
	"github.com/authgear/authgear-server/pkg/util/timeutil"
)

type ReadCounterStore interface {
	GetDailyActiveUserCount(appID config.AppID, date *time.Time) (count int, redisKey string, err error)
	GetWeeklyActiveUserCount(appID config.AppID, year int, week int) (count int, redisKey string, err error)
	GetMonthlyActiveUserCount(appID config.AppID, year int, month int) (count int, redisKey string, err error)
}

type CountCollector struct {
	GlobalHandle  *globaldb.Handle
	GlobalDBStore *GlobalDBStore
	ReadCounter   ReadCounterStore
}

func (c *CountCollector) CollectMonthlyActiveUser(startTime *time.Time) (int, error) {
	startT := timeutil.FirstDayOfTheMonth(*startTime)
	endT := startTime.AddDate(0, 1, 0)
	appIDs, err := c.getAppIDs()
	if err != nil {
		return 0, err
	}

	usageRecords := []*UsageRecord{}
	for _, appID := range appIDs {
		count, _, err := c.ReadCounter.GetMonthlyActiveUserCount(
			config.AppID(appID),
			startTime.Year(),
			int(startTime.Month()),
		)
		if err != nil {
			return 0, err
		}
		if count != 0 {
			usageRecords = append(usageRecords, NewUsageRecord(
				appID,
				ActiveUser,
				count,
				periodical.Monthly,
				startT,
				endT,
			))
		}
	}
	if len(usageRecords) > 0 {
		if err := c.GlobalHandle.WithTx(func() error {
			return c.GlobalDBStore.UpsertUsageRecords(usageRecords)
		}); err != nil {
			return 0, err
		}
		return len(usageRecords), err
	}
	return 0, nil
}

func (c *CountCollector) CollectWeeklyActiveUser(startTime *time.Time) (int, error) {
	startT := timeutil.MondayOfTheWeek(*startTime)
	endT := startT.AddDate(0, 0, 7)
	appIDs, err := c.getAppIDs()
	if err != nil {
		return 0, err
	}
	usageRecords := []*UsageRecord{}
	for _, appID := range appIDs {
		y, w := startTime.ISOWeek()
		count, _, err := c.ReadCounter.GetWeeklyActiveUserCount(
			config.AppID(appID),
			y,
			w,
		)
		if err != nil {
			return 0, err
		}
		if count != 0 {
			usageRecords = append(usageRecords, NewUsageRecord(
				appID,
				ActiveUser,
				count,
				periodical.Weekly,
				startT,
				endT,
			))
		}
	}
	if len(usageRecords) > 0 {
		if err := c.GlobalHandle.WithTx(func() error {
			return c.GlobalDBStore.UpsertUsageRecords(usageRecords)
		}); err != nil {
			return 0, err
		}
		return len(usageRecords), err
	}
	return 0, nil
}

func (c *CountCollector) CollectDailyActiveUser(startTime *time.Time) (int, error) {
	startT := timeutil.TruncateToDate(*startTime)
	endT := startT.AddDate(0, 0, 1)
	appIDs, err := c.getAppIDs()
	if err != nil {
		return 0, err
	}
	usageRecords := []*UsageRecord{}
	for _, appID := range appIDs {
		count, _, err := c.ReadCounter.GetDailyActiveUserCount(
			config.AppID(appID),
			&startT,
		)
		if err != nil {
			return 0, err
		}
		if count != 0 {
			usageRecords = append(usageRecords, NewUsageRecord(
				appID,
				ActiveUser,
				count,
				periodical.Daily,
				startT,
				endT,
			))
		}
	}
	if len(usageRecords) > 0 {
		if err := c.GlobalHandle.WithTx(func() error {
			return c.GlobalDBStore.UpsertUsageRecords(usageRecords)
		}); err != nil {
			return 0, err
		}
		return len(usageRecords), err
	}
	return 0, nil
}

func (c *CountCollector) getAppIDs() (appIDs []string, err error) {
	err = c.GlobalHandle.WithTx(func() error {
		appIDs, err = c.GlobalDBStore.GetAppIDs()
		if err != nil {
			return fmt.Errorf("failed to fetch app ids: %w", err)
		}
		return nil
	})
	return
}
