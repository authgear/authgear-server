package service

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/usage"
	"github.com/authgear/authgear-server/pkg/util/periodical"
)

// fakeGlobalDBStore returns records keyed by RecordName so tests can assert
// that the service queries the correct record names.
// Use the zero value (nil map) to return empty results for every name.
type fakeGlobalDBStore struct {
	byName map[usage.RecordName][]*usage.UsageRecord
	err    error
}

func (f *fakeGlobalDBStore) FetchUsageRecordsInRange(_ context.Context, _ string, name usage.RecordName, _ periodical.Type, from, toExclusive time.Time) ([]*usage.UsageRecord, error) {
	if f.err != nil {
		return nil, f.err
	}
	var out []*usage.UsageRecord
	for _, r := range f.byName[name] {
		t := r.StartTime.UTC().Truncate(24 * time.Hour)
		if !t.Before(from) && t.Before(toExclusive) {
			out = append(out, r)
		}
	}
	return out, nil
}

// fakeDB runs the closure immediately (no real transaction).
type fakeDB struct{}

func (f *fakeDB) WithTx(_ context.Context, do func(context.Context) error) error {
	return do(context.Background())
}

func TestUsageService(t *testing.T) {
	Convey("GetMessagingUsage", t, func() {
		// Note: startDate > endDate and range > 1 year are validated by the transport layer, not the service.
		// No test cases for them here — they belong in handler param tests.

		Convey("returns all-zero counts when no records exist", func() {
			svc := &UsageService{GlobalDatabase: &fakeDB{}, GlobalDBStore: &fakeGlobalDBStore{}}
			result, err := svc.GetMessagingUsage(context.Background(), "app1", "2024-01-01", "2024-01-31")
			So(err, ShouldBeNil)
			So(result.SmsNorthAmericaCount, ShouldEqual, 0)
			So(result.SmsOtherRegionsCount, ShouldEqual, 0)
			So(result.WhatsappNorthAmericaCount, ShouldEqual, 0)
			So(result.WhatsappOtherRegionsCount, ShouldEqual, 0)
		})

		Convey("sums counts per record name independently", func() {
			day1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
			day2 := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
			store := &fakeGlobalDBStore{byName: map[usage.RecordName][]*usage.UsageRecord{
				usage.RecordNameSMSSentNorthAmerica:      {{StartTime: day1, Count: 10}, {StartTime: day2, Count: 5}},
				usage.RecordNameSMSSentOtherRegions:      {{StartTime: day1, Count: 3}},
				usage.RecordNameWhatsappSentNorthAmerica: {{StartTime: day1, Count: 7}},
				usage.RecordNameWhatsappSentOtherRegions: {},
			}}
			svc := &UsageService{GlobalDatabase: &fakeDB{}, GlobalDBStore: store}
			result, err := svc.GetMessagingUsage(context.Background(), "app1", "2024-01-01", "2024-01-31")
			So(err, ShouldBeNil)
			So(result.SmsNorthAmericaCount, ShouldEqual, 15)
			So(result.SmsOtherRegionsCount, ShouldEqual, 3)
			So(result.WhatsappNorthAmericaCount, ShouldEqual, 7)
			So(result.WhatsappOtherRegionsCount, ShouldEqual, 0)
		})
	})

	Convey("GetMonthlyActiveUsersUsage", t, func() {
		Convey("fills in 0 for months with no record", func() {
			// Only January has a record; February should be 0.
			jan := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
			store := &fakeGlobalDBStore{byName: map[usage.RecordName][]*usage.UsageRecord{
				usage.RecordNameActiveUser: {{StartTime: jan, Count: 500}},
			}}
			svc := &UsageService{GlobalDatabase: &fakeDB{}, GlobalDBStore: store}
			result, err := svc.GetMonthlyActiveUsersUsage(context.Background(), "app1", 2024, 1, 2024, 2)
			So(err, ShouldBeNil)
			So(result.Counts, ShouldHaveLength, 2)
			So(result.Counts[0].Count, ShouldEqual, 500)
			So(result.Counts[1].Count, ShouldEqual, 0)
		})

		Convey("handles year rollover (Dec to Jan)", func() {
			dec := time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC)
			jan := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
			store := &fakeGlobalDBStore{byName: map[usage.RecordName][]*usage.UsageRecord{
				usage.RecordNameActiveUser: {{StartTime: dec, Count: 300}, {StartTime: jan, Count: 400}},
			}}
			svc := &UsageService{GlobalDatabase: &fakeDB{}, GlobalDBStore: store}
			result, err := svc.GetMonthlyActiveUsersUsage(context.Background(), "app1", 2023, 12, 2024, 1)
			So(err, ShouldBeNil)
			So(result.Counts, ShouldHaveLength, 2)
			So(result.Counts[0].Count, ShouldEqual, 300)
			So(result.Counts[1].Count, ShouldEqual, 400)
		})

		Convey("endMonth=12 computes correct exclusive upper bound", func() {
			// time.Month(13) normalises to January of the next year in time.Date — no special case needed.
			dec := time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC)
			store := &fakeGlobalDBStore{byName: map[usage.RecordName][]*usage.UsageRecord{
				usage.RecordNameActiveUser: {{StartTime: dec, Count: 100}},
			}}
			svc := &UsageService{GlobalDatabase: &fakeDB{}, GlobalDBStore: store}
			result, err := svc.GetMonthlyActiveUsersUsage(context.Background(), "app1", 2024, 12, 2024, 12)
			So(err, ShouldBeNil)
			So(result.Counts, ShouldHaveLength, 1)
			So(result.Counts[0].Count, ShouldEqual, 100)
		})
	})
}
