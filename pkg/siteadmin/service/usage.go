package service

import (
	"context"
	"time"

	"github.com/authgear/authgear-server/pkg/api/siteadmin"
	"github.com/authgear/authgear-server/pkg/lib/usage"
	"github.com/authgear/authgear-server/pkg/util/periodical"
	"github.com/authgear/authgear-server/pkg/util/timeutil"
)

// ---- Narrow interfaces -------------------------------------------------------

type UsageServiceGlobalDBStore interface {
	FetchUsageRecordsInRange(
		ctx context.Context,
		appID string,
		recordName usage.RecordName,
		period periodical.Type,
		fromStartTime time.Time,
		toEndTimeExclusive time.Time,
	) ([]*usage.UsageRecord, error)
}

// ---- UsageService ------------------------------------------------------------

type UsageService struct {
	GlobalDatabase AppServiceDatabase
	GlobalDBStore  UsageServiceGlobalDBStore
}

func (s *UsageService) GetMessagingUsage(ctx context.Context, appID string, startDate string, endDate string) (*siteadmin.MessagingUsage, error) {
	// Transport has already validated format and ordering; parse errors here are unexpected.
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, err
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, err
	}
	// endDate is inclusive; add one day for the exclusive upper bound.
	endExclusive := end.AddDate(0, 0, 1)

	var smsNA, smsOther, whatsappNA, whatsappOther int
	err = s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		var e error
		smsNA, e = sumDailyRecords(ctx, s.GlobalDBStore, appID, usage.RecordNameSMSSentNorthAmerica, start, endExclusive)
		if e != nil {
			return e
		}
		smsOther, e = sumDailyRecords(ctx, s.GlobalDBStore, appID, usage.RecordNameSMSSentOtherRegions, start, endExclusive)
		if e != nil {
			return e
		}
		whatsappNA, e = sumDailyRecords(ctx, s.GlobalDBStore, appID, usage.RecordNameWhatsappSentNorthAmerica, start, endExclusive)
		if e != nil {
			return e
		}
		whatsappOther, e = sumDailyRecords(ctx, s.GlobalDBStore, appID, usage.RecordNameWhatsappSentOtherRegions, start, endExclusive)
		return e
	})
	if err != nil {
		return nil, err
	}

	return &siteadmin.MessagingUsage{
		StartDate:                 startDate,
		EndDate:                   endDate,
		SmsNorthAmericaCount:      smsNA,
		SmsOtherRegionsCount:      smsOther,
		WhatsappNorthAmericaCount: whatsappNA,
		WhatsappOtherRegionsCount: whatsappOther,
	}, nil
}

func (s *UsageService) GetMonthlyActiveUsersUsage(ctx context.Context, appID string, startYear int, startMonth int, endYear int, endMonth int) (*siteadmin.MonthlyActiveUsersUsage, error) {
	fromTime := timeutil.FirstDayOfTheMonth(time.Date(startYear, time.Month(startMonth), 1, 0, 0, 0, 0, time.UTC))
	// toTime is the first day of the month after endMonth (exclusive upper bound).
	toTime := timeutil.FirstDayOfTheMonth(time.Date(endYear, time.Month(endMonth)+1, 1, 0, 0, 0, 0, time.UTC))

	var records []*usage.UsageRecord
	err := s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		var e error
		records, e = s.GlobalDBStore.FetchUsageRecordsInRange(ctx, appID, usage.RecordNameActiveUser, periodical.Monthly, fromTime, toTime)
		return e
	})
	if err != nil {
		return nil, err
	}

	// Index records by start_time for O(1) lookup.
	byMonth := make(map[time.Time]int, len(records))
	for _, r := range records {
		byMonth[r.StartTime.UTC().Truncate(24*time.Hour)] = r.Count
	}

	totalMonths := (endYear-startYear)*12 + (endMonth - startMonth) + 1
	if totalMonths <= 0 {
		return &siteadmin.MonthlyActiveUsersUsage{Counts: nil}, nil
	}
	counts := make([]siteadmin.MonthlyActiveUsersCount, 0, totalMonths)
	year, month := startYear, startMonth
	for i := 0; i < totalMonths; i++ {
		t := timeutil.FirstDayOfTheMonth(time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC))
		counts = append(counts, siteadmin.MonthlyActiveUsersCount{
			Year:  year,
			Month: month,
			Count: byMonth[t],
		})
		month++
		if month > 12 {
			month = 1
			year++
		}
	}

	return &siteadmin.MonthlyActiveUsersUsage{Counts: counts}, nil
}

// sumDailyRecords fetches daily usage records in the range and returns the total count.
func sumDailyRecords(ctx context.Context, store UsageServiceGlobalDBStore, appID string, name usage.RecordName, from time.Time, to time.Time) (int, error) {
	records, err := store.FetchUsageRecordsInRange(ctx, appID, name, periodical.Daily, from, to)
	if err != nil {
		return 0, err
	}
	total := 0
	for _, r := range records {
		total += r.Count
	}
	return total, nil
}
