# Part 5: Site Admin API — Usages API Real Data

## Context

Replace dummy data in `MessagingUsageHandler` and `MonthlyActiveUsersUsageHandler` with
real data from the `_portal_usage_record` table in the global database.

The two endpoints affected are:

- `GET /api/v1/apps/:appID/usage/messaging` — SMS and WhatsApp message counts for a date range
- `GET /api/v1/apps/:appID/usage/monthly-active-users` — monthly active user counts for a month range

**Key data sources:**

| Field | Source |
|---|---|
| `sms_north_america_count`, `sms_other_regions_count` | `_portal_usage_record`, `name` = `sms-sent.north-america` / `sms-sent.other-regions`, `period` = `daily` |
| `whatsapp_north_america_count`, `whatsapp_other_regions_count` | `_portal_usage_record`, `name` = `whatsapp-sent.north-america` / `whatsapp-sent.other-regions`, `period` = `daily` |
| `counts[*].count` | `_portal_usage_record`, `name` = `active-user`, `period` = `monthly` |

**Design decisions:**

- Both endpoints read from `_portal_usage_record` (pre-aggregated by the usage collector cron
  job). No Redis or audit DB dependency is needed.
- A single `siteadminservice.UsageService` handles both endpoints, following the same pattern
  as `siteadminservice.AppService`.
- **Messaging**: sum daily records whose `start_time` falls in `[startDate, endDate]` (inclusive)
  for each of the four record names. The transport layer's `parseMessagingUsageParams` validates
  two constraints using `makeValidationError`: (1) `startDate <= endDate`, and (2) the range does
  not exceed 1 year (`end <= start + 1 year` via `start.AddDate(1, 0, 0)`, handles leap years). The service receives already-validated strings
  and only needs to `time.Parse` them — no validation error path in the service layer.
- **MAU**: fetch all monthly `active-user` records for the month range in one query, index by
  `start_time`, then iterate the requested months filling in 0 for any missing entry.
  `time.Date(endYear, time.Month(endMonth)+1, 1, ...)` is used to compute the exclusive upper
  bound — Go's `time.Date` normalises month 13 to January of the following year, so no special
  case is needed when `endMonth == 12`.
- The existing `usage.GlobalDBStore.FetchUsageRecordsInRange` is reused via a narrow interface —
  no new SQL needed.
- Siteadmin already wires `*globaldb.SQLBuilder` and `*globaldb.SQLExecutor`, so
  `usage.GlobalDBStore` can be added as a partial struct with no new infra dependencies.

> **Note:** MAU data is populated by the `usage.CountCollector.CollectMonthlyActiveUser` cron
> job. Data may be up to one day stale.

---

## Architecture Overview

```
MessagingUsageHandler / MonthlyActiveUsersUsageHandler (transport)
    │  depends on
    ▼
UsageService (pkg/siteadmin/service/usage.go)
    │  depends on
    ├── AppServiceDatabase         → *globaldb.Handle      (transaction)
    └── UsageServiceGlobalDBStore → *usage.GlobalDBStore  (SQL queries)
```

### `GetMessagingUsage` flow

```
Transport (parseMessagingUsageParams):
1. Parse startDate / endDate strings → validate format via getDateParam
2. If start > end → makeValidationError on "end_date"

Service (GetMessagingUsage):
3. time.Parse startDate / endDate → time.Time (UTC midnight; always succeeds, transport validated)
4. endExclusive = endDate + 24h  (exclusive upper bound for daily records)
5. GlobalDatabase.WithTx:
   a. FetchUsageRecordsInRange(appID, "sms-sent.north-america",      daily, start, endExclusive) → sum
   b. FetchUsageRecordsInRange(appID, "sms-sent.other-regions",      daily, start, endExclusive) → sum
   c. FetchUsageRecordsInRange(appID, "whatsapp-sent.north-america", daily, start, endExclusive) → sum
   d. FetchUsageRecordsInRange(appID, "whatsapp-sent.other-regions", daily, start, endExclusive) → sum
6. Return siteadmin.MessagingUsage
```

### `GetMonthlyActiveUsersUsage` flow

```
1. fromTime = firstDayOfMonth(startYear, startMonth)
2. toTime   = firstDayOfMonth(endYear, endMonth+1)  [exclusive]
3. GlobalDatabase.WithTx:
   a. FetchUsageRecordsInRange(appID, "active-user", monthly, fromTime, toTime) → index by start_time
4. Iterate months from start to end inclusive:
   a. look up count by firstDayOfMonth(year, month); default 0 if missing
   b. append MonthlyActiveUsersCount{Year, Month, Count}
5. Return siteadmin.MonthlyActiveUsersUsage
```

---

## Key Dependencies

| What | Where |
|---|---|
| `usage.GlobalDBStore.FetchUsageRecordsInRange` | `pkg/lib/usage/globaldb_store.go` |
| `usage.RecordNameSMSSentNorthAmerica` etc. | `pkg/lib/usage/usage_record.go` |
| `usage.RecordNameActiveUser` | `pkg/lib/usage/usage_record.go` |
| `periodical.Daily` / `periodical.Monthly` | `pkg/util/periodical` |
| `timeutil.FirstDayOfTheMonth` | `pkg/util/timeutil` |
| `siteadmin.MessagingUsage` | `pkg/api/siteadmin/gen.go` |
| `siteadmin.MonthlyActiveUsersUsage` | `pkg/api/siteadmin/gen.go` |
| `AppServiceDatabase` | already defined in `pkg/siteadmin/service/app.go` (reused) |

---

## Files to Create

### 1. `pkg/siteadmin/service/usage.go`

```go
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

	var counts []siteadmin.MonthlyActiveUsersCount
	year, month := startYear, startMonth
	for {
		t := timeutil.FirstDayOfTheMonth(time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC))
		counts = append(counts, siteadmin.MonthlyActiveUsersCount{
			Year:  year,
			Month: month,
			Count: byMonth[t],
		})
		if year == endYear && month == endMonth {
			break
		}
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
```

### 2. `pkg/siteadmin/service/usage_test.go`

Added in **Commit 1** alongside `usage.go`. Uses hand-written fakes and GoConvey, matching
the pattern in `pkg/siteadmin/service/app_test.go`.

```go
package service_test

import (
    "context"
    "testing"
    "time"

    . "github.com/smartystreets/goconvey/convey"

    "github.com/authgear/authgear-server/pkg/lib/usage"
    "github.com/authgear/authgear-server/pkg/util/periodical"
    . "github.com/authgear/authgear-server/pkg/siteadmin/service"
)

// fakeGlobalDBStore returns records keyed by RecordName so tests can assert
// that the service queries the correct record names.
// Use the zero value (nil map) to return empty results for every name.
type fakeGlobalDBStore struct {
    byName map[usage.RecordName][]*usage.UsageRecord
    err    error
}

func (f *fakeGlobalDBStore) FetchUsageRecordsInRange(_ context.Context, _ string, name usage.RecordName, _ periodical.Type, _, _ time.Time) ([]*usage.UsageRecord, error) {
    return f.byName[name], f.err
}

// fakeDB runs the closure immediately (no real transaction).
type fakeDB struct{}

func (f *fakeDB) WithTx(_ context.Context, do func(context.Context) error) error {
    return do(context.Background())
}
```

#### Test cases

```go
func TestUsageService(t *testing.T) {
    Convey("GetMessagingUsage", t, func() {
        // Note: startDate > endDate is validated by the transport layer, not the service.
        // No test case for it here — it belongs in a handler param test.

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
            store := &fakeGlobalDBStore{byName: map[usage.RecordName][]*usage.UsageRecord{
                usage.RecordNameSMSSentNorthAmerica:      {{Count: 10}, {Count: 5}},
                usage.RecordNameSMSSentOtherRegions:      {{Count: 3}},
                usage.RecordNameWhatsappSentNorthAmerica: {{Count: 7}},
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

        Convey("handles year rollover (Dec → Jan)", func() {
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
```

> **Note on `fakeGlobalDBStore`**: the fake is keyed by `RecordName` so each test case controls
> exactly which records are returned per name. This ensures the service queries the correct
> record names and sums are truly independent.

---

## Files to Modify

### `pkg/siteadmin/service/deps.go`

Add `UsageService` to the dependency set:

```go
var DependencySet = wire.NewSet(
	wire.Struct(new(AppOwnerStore), "*"),
	wire.Bind(new(AppServiceOwnerStore), new(*AppOwnerStore)),
	wire.Struct(new(AppService), "*"),
	NewHTTPClient,
	wire.Struct(new(CollaboratorService), "*"),
	wire.Struct(new(UsageService), "*"),  // NEW
)
```

### `pkg/siteadmin/deps.go`

Two changes:

**1. Add partial `usage.GlobalDBStore` and its binding** (import alias `usagepkg`):

```go
// usage.GlobalDBStore satisfies UsageServiceGlobalDBStore
wire.Struct(new(usagepkg.GlobalDBStore), "SQLBuilder", "SQLExecutor"),
wire.Bind(new(siteadminservice.UsageServiceGlobalDBStore), new(*usagepkg.GlobalDBStore)),
```

**2. Add transport bindings:**

```go
wire.Bind(new(transport.MessagingUsageService), new(*siteadminservice.UsageService)),
wire.Bind(new(transport.MonthlyActiveUsersUsageService), new(*siteadminservice.UsageService)),
```

### `pkg/siteadmin/transport/handler_messaging_usage.go`

Add service interface and `Service` field; add `startDate > endDate` and range-exceeds-1-year
validations to `parseMessagingUsageParams`; replace dummy `ServeHTTP` body:

```go
type MessagingUsageService interface {
	GetMessagingUsage(ctx context.Context, appID string, startDate string, endDate string) (*siteadmin.MessagingUsage, error)
}

type MessagingUsageHandler struct {
	Service MessagingUsageService
}

// parseMessagingUsageParams validates that:
//  1. startDate <= endDate
//  2. the range does not exceed 1 year (end_date - start_date <= 365 days)
//
// Both checks live here (not in the service) because makeValidationError is
// transport-package-local.
func parseMessagingUsageParams(r *http.Request) (MessagingUsageParams, error) {
	q := r.URL.Query()

	startDate, err := getDateParam(q, "start_date")
	if err != nil {
		return MessagingUsageParams{}, err
	}

	endDate, err := getDateParam(q, "end_date")
	if err != nil {
		return MessagingUsageParams{}, err
	}

	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)
	if start.After(end) {
		return MessagingUsageParams{}, makeValidationError(func(ctx *validation.Context) {
			ctx.Child("end_date").EmitError("range", map[string]interface{}{"details": "end_date must not be before start_date"})
		})
	}
	if end.After(start.AddDate(1, 0, 0)) {
		return MessagingUsageParams{}, makeValidationError(func(ctx *validation.Context) {
			ctx.Child("end_date").EmitError("range", map[string]interface{}{"details": "date range must not exceed 1 year"})
		})
	}

	return MessagingUsageParams{
		AppID:     httproute.GetParam(r, "appID"),
		StartDate: startDate,
		EndDate:   endDate,
	}, nil
}

func (h *MessagingUsageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params, err := parseMessagingUsageParams(r)
	if err != nil {
		writeError(w, r, err)
		return
	}

	usage, err := h.Service.GetMessagingUsage(r.Context(), params.AppID, params.StartDate, params.EndDate)
	if err != nil {
		writeError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(usage)
}
```

### `pkg/siteadmin/transport/handler_monthly_active_users_usage.go`

Add service interface and `Service` field; replace dummy `ServeHTTP` body:

```go
type MonthlyActiveUsersUsageService interface {
	GetMonthlyActiveUsersUsage(ctx context.Context, appID string, startYear int, startMonth int, endYear int, endMonth int) (*siteadmin.MonthlyActiveUsersUsage, error)
}

type MonthlyActiveUsersUsageHandler struct {
	Service MonthlyActiveUsersUsageService
}

func (h *MonthlyActiveUsersUsageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params, err := parseMonthlyActiveUsersUsageParams(r)
	if err != nil {
		writeError(w, r, err)
		return
	}

	usage, err := h.Service.GetMonthlyActiveUsersUsage(
		r.Context(), params.AppID,
		params.StartYear, params.StartMonth,
		params.EndYear, params.EndMonth,
	)
	if err != nil {
		writeError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(usage)
}
```

### `pkg/siteadmin/wire_gen.go`

Regenerated via `wire gen ./pkg/siteadmin/...` after all deps changes. Do **not** hand-edit.

---

## Implementation Roadmap: 4 Atomic Commits

### **Commit 1: Add `UsageService` to siteadmin service layer**

**Files Created:**
- `pkg/siteadmin/service/usage.go`
- `pkg/siteadmin/service/usage_test.go`

**Files Modified:**
- `pkg/siteadmin/service/deps.go` — add `wire.Struct(new(UsageService), "*")`

**Scope:** Business logic and unit tests — no DI wiring, no handler changes.

**Verification:**
```bash
go test ./pkg/siteadmin/service/...
make lint
```

**Commit Message:** `"Add siteadmin UsageService for messaging and MAU usage"`

---

### **Commit 2: Add service interfaces to transport handlers**

**Files Modified:**
- `pkg/siteadmin/transport/handler_messaging_usage.go` — add `MessagingUsageService` interface + `Service` field; add 1-year range validation to `parseMessagingUsageParams`
- `pkg/siteadmin/transport/handler_monthly_active_users_usage.go` — add `MonthlyActiveUsersUsageService` interface + `Service` field

**Scope:** Struct and interface declarations only. `ServeHTTP` bodies unchanged (still dummy).

**Verification:**
```bash
go build ./pkg/siteadmin/...
make lint
```

**Commit Message:** `"Wire service interfaces into usage transport handlers"`

---

### **Commit 3: Wire UsageService into DI and regenerate**

**Files Modified:**
- `pkg/siteadmin/deps.go` — add partial `usage.GlobalDBStore`; add `UsageServiceGlobalDBStore` and transport bindings
- `pkg/siteadmin/wire_gen.go` — regenerated

**Build Steps:**
```bash
wire gen ./pkg/siteadmin/...
go mod tidy
go build ./pkg/siteadmin/...
go build ./cmd/portal/...
make lint
```

**Commit Message:** `"Wire UsageService into siteadmin DI and regenerate"`

---

### **Commit 4: Replace handler bodies with real service calls**

**Files Modified:**
- `pkg/siteadmin/transport/handler_messaging_usage.go` — call `h.Service.GetMessagingUsage`
- `pkg/siteadmin/transport/handler_monthly_active_users_usage.go` — call `h.Service.GetMonthlyActiveUsersUsage`

**Scope:** `ServeHTTP` bodies only — no interface or DI changes.

**Verification:**
```bash
go build ./pkg/siteadmin/...
go build ./cmd/portal/...
go test ./pkg/siteadmin/...
make lint
go run ./devtools/goanalysis ./cmd/... ./pkg/...   # update .vettedpositions for any line-number drift
make sort-vettedpositions
make check-tidy
```

**Commit Message:** `"Replace usage handler stubs with real service calls"`

---

## Dependency Graph

```
Commit 1 (UsageService + service/deps.go)
    ↓
Commit 2 (transport handler interfaces + Service fields)
    ↓
Commit 3 (deps.go wiring + wire gen)
    ↓
Commit 4 (handler ServeHTTP bodies)
```

**Key Properties:**
- ✅ Each commit is independently reviewable
- ✅ No new infra dependencies — reuses existing globaldb wiring already in siteadmin
- ✅ Unit tests ship with the service in Commit 1
- ✅ Build passes after Commit 3
- ✅ Endpoints return real data after Commit 4

---

## Verification

### Messaging usage

```bash
curl -s -H "Authorization: Bearer <token>" \
  "http://localhost:3005/api/v1/apps/myapp/usage/messaging?start_date=2024-01-01&end_date=2024-01-31" | jq .
```

Expected response:
```json
{
  "start_date": "2024-01-01",
  "end_date": "2024-01-31",
  "sms_north_america_count": 120,
  "sms_other_regions_count": 45,
  "whatsapp_north_america_count": 30,
  "whatsapp_other_regions_count": 15
}
```

If no usage records exist for the app and date range, all counts will be `0`.

### Monthly active users

```bash
curl -s -H "Authorization: Bearer <token>" \
  "http://localhost:3005/api/v1/apps/myapp/usage/monthly-active-users?start_year=2024&start_month=1&end_year=2024&end_month=3" | jq .
```

Expected response:
```json
{
  "counts": [
    { "year": 2024, "month": 1, "count": 500 },
    { "year": 2024, "month": 2, "count": 520 },
    { "year": 2024, "month": 3, "count": 0 }
  ]
}
```

Months with no collected usage record return `count: 0`.

