package usage

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/audit"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/authgear/authgear-server/pkg/util/periodical"
	phoneutil "github.com/authgear/authgear-server/pkg/util/phone"
	"github.com/authgear/authgear-server/pkg/util/timeutil"
)

type ReadCounterStore interface {
	GetDailyActiveUserCount(appID config.AppID, date *time.Time) (count int, redisKey string, err error)
	GetWeeklyActiveUserCount(appID config.AppID, year int, week int) (count int, redisKey string, err error)
	GetMonthlyActiveUserCount(appID config.AppID, year int, month int) (count int, redisKey string, err error)
}

type MeterAuditDBStore interface {
	QueryPage(appID string, opts audit.QueryPageOptions, pageArgs graphqlutil.PageArgs) ([]*audit.Log, uint64, error)
	GetCountByActivityType(appID string, activityType string, rangeFrom *time.Time, rangeTo *time.Time) (int, error)
}

type smsCountResult struct {
	northAmerica int
	otherRegions int
	total        int
}

type whatsappCountResult struct {
	northAmerica int
	otherRegions int
	total        int
}

type CountCollector struct {
	GlobalHandle  *globaldb.Handle
	GlobalDBStore *GlobalDBStore
	ReadCounter   ReadCounterStore
	AuditHandle   *auditdb.ReadHandle
	Meters        MeterAuditDBStore
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
				RecordNameActiveUser,
				count,
				periodical.Monthly,
				startT,
				endT,
			))
		}
	}

	return c.upsertUsageRecords(usageRecords)
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
				RecordNameActiveUser,
				count,
				periodical.Weekly,
				startT,
				endT,
			))
		}
	}

	return c.upsertUsageRecords(usageRecords)
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
				RecordNameActiveUser,
				count,
				periodical.Daily,
				startT,
				endT,
			))
		}
	}

	return c.upsertUsageRecords(usageRecords)
}

func (c *CountCollector) CollectDailySMSSent(startTime *time.Time) (int, error) {
	startT := timeutil.TruncateToDate(*startTime)
	endT := startT.AddDate(0, 0, 1)
	appIDs, err := c.getAppIDs()
	if err != nil {
		return 0, err
	}

	usageRecords := []*UsageRecord{}
	for _, appID := range appIDs {

		result, err := c.querySMSCount(appID, &startT, &endT)
		if err != nil {
			return 0, err
		}

		if result.northAmerica > 0 {
			usageRecords = append(usageRecords, NewUsageRecord(appID, RecordNameSMSSentNorthAmerica, result.northAmerica, periodical.Daily, startT, endT))
		}

		if result.otherRegions > 0 {
			usageRecords = append(usageRecords, NewUsageRecord(appID, RecordNameSMSSentOtherRegions, result.otherRegions, periodical.Daily, startT, endT))
		}

		if result.total > 0 {
			usageRecords = append(usageRecords, NewUsageRecord(appID, RecordNameSMSSentTotal, result.total, periodical.Daily, startT, endT))
		}
	}

	return c.upsertUsageRecords(usageRecords)
}

func (c *CountCollector) CollectDailyEmailSent(startTime *time.Time) (int, error) {
	startT := timeutil.TruncateToDate(*startTime)
	endT := startT.AddDate(0, 0, 1)
	appIDs, err := c.getAppIDs()
	if err != nil {
		return 0, err
	}

	usageRecords := []*UsageRecord{}
	for _, appID := range appIDs {
		err := c.AuditHandle.ReadOnly(func() (e error) {
			count, err := c.Meters.GetCountByActivityType(appID, string(nonblocking.EmailSent), &startT, &endT)
			if err != nil {
				return err
			}
			if count > 0 {
				usageRecords = append(usageRecords, NewUsageRecord(
					appID,
					RecordNameEmailSent,
					count,
					periodical.Daily,
					startT,
					endT,
				))
			}
			return nil
		})
		if err != nil {
			return 0, err
		}
	}

	return c.upsertUsageRecords(usageRecords)
}

func (c *CountCollector) CollectDailyWhatsappSent(startTime *time.Time) (int, error) {
	startT := timeutil.TruncateToDate(*startTime)
	endT := startT.AddDate(0, 0, 1)
	appIDs, err := c.getAppIDs()
	if err != nil {
		return 0, err
	}

	usageRecords := []*UsageRecord{}
	for _, appID := range appIDs {
		result, err := c.queryWhatsappCount(appID, &startT, &endT)
		if err != nil {
			return 0, err
		}

		if result.northAmerica > 0 {
			usageRecords = append(usageRecords, NewUsageRecord(appID, RecordNameWhatsappSentNorthAmerica, result.northAmerica, periodical.Daily, startT, endT))
		}

		if result.otherRegions > 0 {
			usageRecords = append(usageRecords, NewUsageRecord(appID, RecordNameWhatsappSentOtherRegions, result.otherRegions, periodical.Daily, startT, endT))
		}

		if result.total > 0 {
			usageRecords = append(usageRecords, NewUsageRecord(appID, RecordNameWhatsappSentTotal, result.total, periodical.Daily, startT, endT))
		}
	}

	return c.upsertUsageRecords(usageRecords)
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

func (c *CountCollector) upsertUsageRecords(usageRecords []*UsageRecord) (int, error) {
	if len(usageRecords) > 0 {
		if err := c.GlobalHandle.WithTx(func() error {
			return c.GlobalDBStore.UpsertUsageRecords(usageRecords)
		}); err != nil {
			return 0, err
		}
		return len(usageRecords), nil
	}
	return 0, nil
}

func (c *CountCollector) querySMSCount(appID string, rangeFrom *time.Time, rangeTo *time.Time) (*smsCountResult, error) {
	result := &smsCountResult{}
	var first uint64 = 100
	var after model.PageCursor = ""
	for {
		var err error
		var events []*event.Event
		err = c.AuditHandle.ReadOnly(func() error {
			events, after, err = c.queryEvents(
				nonblocking.SMSSent,
				func() event.Payload { return &nonblocking.SMSSentEventPayload{} },
				appID, rangeFrom, rangeTo, first, after)
			return err
		})
		if err != nil {
			return nil, err
		}
		// Termination condition
		if len(events) == 0 {
			return result, nil
		}
		for _, e := range events {
			payload, ok := e.Payload.(*nonblocking.SMSSentEventPayload)
			if !ok {
				return nil, errors.New("usage: unexpected event payload")
			}
			if payload.IsNotCountedInUsage {
				continue
			}

			e164 := payload.Recipient

			isNorthAmericaNumber, err := phoneutil.IsNorthAmericaNumber(e164)
			if err != nil {
				return nil, fmt.Errorf("usage: failed to parse sms recipient %w", err)
			}

			result.total++

			if isNorthAmericaNumber {
				result.northAmerica++
			} else {
				result.otherRegions++
			}
		}
	}
}

func (c *CountCollector) queryWhatsappCount(appID string, rangeFrom *time.Time, rangeTo *time.Time) (*whatsappCountResult, error) {
	result := &whatsappCountResult{}
	var first uint64 = 100
	var after model.PageCursor = ""
	for {
		var err error
		var events []*event.Event
		err = c.AuditHandle.ReadOnly(func() error {
			events, after, err = c.queryEvents(
				nonblocking.WhatsappSent,
				func() event.Payload { return &nonblocking.WhatsappSentEventPayload{} },
				appID, rangeFrom, rangeTo, first, after)
			return err
		})
		if err != nil {
			return nil, err
		}
		// Termination condition
		if len(events) == 0 {
			return result, nil
		}
		for _, e := range events {
			payload, ok := e.Payload.(*nonblocking.WhatsappSentEventPayload)
			if !ok {
				return nil, errors.New("usage: unexpected event payload")
			}

			e164 := payload.Recipient

			isNorthAmericaNumber, err := phoneutil.IsNorthAmericaNumber(e164)
			if err != nil {
				return nil, fmt.Errorf("usage: failed to parse whatsapp recipient %w", err)
			}
			if payload.IsNotCountedInUsage {
				continue
			}

			result.total++

			if isNorthAmericaNumber {
				result.northAmerica++
			} else {
				result.otherRegions++
			}
		}
	}
}

func (c *CountCollector) queryEvents(
	eventType event.Type,
	payloadFactory func() event.Payload,
	appID string,
	rangeFrom *time.Time,
	rangeTo *time.Time,
	first uint64,
	after model.PageCursor) (events []*event.Event, lastCursor model.PageCursor, err error) {
	options := audit.QueryPageOptions{
		RangeFrom:     rangeFrom,
		RangeTo:       rangeTo,
		ActivityTypes: []string{string(eventType)},
	}

	logs, offset, err := c.Meters.QueryPage(appID, options, graphqlutil.PageArgs{
		First: &first,
		After: graphqlutil.Cursor(after),
	})
	if err != nil {
		return
	}
	events = make([]*event.Event, len(logs))
	for i, log := range logs {
		b, e := json.Marshal(log.Data)
		if e != nil {
			err = e
			return
		}
		eventObj := event.Event{
			Payload: payloadFactory(),
		}
		e = json.Unmarshal(b, &eventObj)
		if e != nil {
			err = e
			return
		}
		events[i] = &eventObj
	}

	pageKey := db.PageKey{Offset: offset + uint64(len(logs)) - 1}
	cursor, err := pageKey.ToPageCursor()
	if err != nil {
		return
	}
	after = cursor

	return events, after, nil
}
