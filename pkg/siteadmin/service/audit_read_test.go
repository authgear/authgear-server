package service

import (
	"context"
	"sort"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/siteadmin"
)

// ---- Fakes ------------------------------------------------------------------

type fakeAuditReadDatabase struct{}

func (f *fakeAuditReadDatabase) ReadOnly(ctx context.Context, do func(context.Context) error) error {
	return do(ctx)
}

type fakeAuditLogStore struct {
	rows []AuditLogEntry
}

func (f *fakeAuditLogStore) Count(_ context.Context, affectedAppID string) (int, error) {
	return len(f.filtered(affectedAppID)), nil
}

func (f *fakeAuditLogStore) List(_ context.Context, affectedAppID string, order siteadmin.OrderDirection, limit, offset uint64) ([]AuditLogEntry, error) {
	rows := make([]AuditLogEntry, len(f.filtered(affectedAppID)))
	copy(rows, f.filtered(affectedAppID))

	sort.SliceStable(rows, func(i, j int) bool {
		if order == siteadmin.Asc {
			return rows[i].CreatedAt.Before(rows[j].CreatedAt)
		}
		return rows[i].CreatedAt.After(rows[j].CreatedAt)
	})

	if offset >= uint64(len(rows)) {
		return nil, nil
	}
	end := min(offset+limit, uint64(len(rows)))
	return rows[offset:end], nil
}

func (f *fakeAuditLogStore) Get(_ context.Context, id string) (*AuditLogEntryDetail, error) {
	for _, r := range f.rows {
		if r.ID == id {
			return &AuditLogEntryDetail{AuditLogEntry: r, Data: map[string]any{"type": "test"}}, nil
		}
	}
	return nil, apierrors.NewNotFound("audit log not found")
}

func (f *fakeAuditLogStore) filtered(affectedAppID string) []AuditLogEntry {
	if affectedAppID == "" {
		return f.rows
	}
	var out []AuditLogEntry
	for _, r := range f.rows {
		if r.AffectedAppID == affectedAppID {
			out = append(out, r)
		}
	}
	return out
}

// ---- Tests ------------------------------------------------------------------

func TestSiteAdminAuditReadService(t *testing.T) {
	t1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) // oldest
	t2 := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	t3 := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC) // newest

	makeService := func(store *fakeAuditLogStore) *SiteAdminAuditReadService {
		return &SiteAdminAuditReadService{
			AuditDatabase: &fakeAuditReadDatabase{},
			Store:         store,
		}
	}

	Convey("SiteAdminAuditReadService", t, func() {

		Convey("ListAuditLogs: no entries returns empty", func() {
			svc := makeService(&fakeAuditLogStore{})
			result, err := svc.ListAuditLogs(ctxWithSession(), ListAuditLogsParams{Page: 1, PageSize: 10})

			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 0)
			So(result.Entries, ShouldResemble, []AuditLogEntry{})
		})

		Convey("ListAuditLogs: default order=desc returns newest first", func() {
			store := &fakeAuditLogStore{rows: []AuditLogEntry{
				{ID: "1", CreatedAt: t1, ActivityType: "site_admin.app.plan.updated"},
				{ID: "2", CreatedAt: t2, ActivityType: "site_admin.app.plan.updated"},
				{ID: "3", CreatedAt: t3, ActivityType: "site_admin.app.plan.updated"},
			}}
			svc := makeService(store)
			result, err := svc.ListAuditLogs(ctxWithSession(), ListAuditLogsParams{Page: 1, PageSize: 10})

			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 3)
			So(result.Entries[0].ID, ShouldEqual, "3") // newest first
			So(result.Entries[2].ID, ShouldEqual, "1")
		})

		Convey("ListAuditLogs: order=asc returns oldest first", func() {
			store := &fakeAuditLogStore{rows: []AuditLogEntry{
				{ID: "1", CreatedAt: t1, ActivityType: "site_admin.app.plan.updated"},
				{ID: "2", CreatedAt: t2, ActivityType: "site_admin.app.plan.updated"},
				{ID: "3", CreatedAt: t3, ActivityType: "site_admin.app.plan.updated"},
			}}
			svc := makeService(store)
			result, err := svc.ListAuditLogs(ctxWithSession(), ListAuditLogsParams{
				Page: 1, PageSize: 10, Order: siteadmin.Asc,
			})

			So(err, ShouldBeNil)
			So(result.Entries[0].ID, ShouldEqual, "1") // oldest first
			So(result.Entries[2].ID, ShouldEqual, "3")
		})

		Convey("ListAuditLogs: invalid order defaults to desc", func() {
			store := &fakeAuditLogStore{rows: []AuditLogEntry{
				{ID: "1", CreatedAt: t1, ActivityType: "site_admin.app.plan.updated"},
				{ID: "2", CreatedAt: t3, ActivityType: "site_admin.app.plan.updated"},
			}}
			svc := makeService(store)
			result, err := svc.ListAuditLogs(ctxWithSession(), ListAuditLogsParams{
				Page: 1, PageSize: 10, Order: "invalid",
			})

			So(err, ShouldBeNil)
			So(result.Entries[0].ID, ShouldEqual, "2") // newest first (desc default)
		})

		Convey("ListAuditLogs: AffectedAppID filter", func() {
			store := &fakeAuditLogStore{rows: []AuditLogEntry{
				{ID: "1", CreatedAt: t1, ActivityType: "site_admin.app.plan.updated", AffectedAppID: "app-a"},
				{ID: "2", CreatedAt: t2, ActivityType: "site_admin.app.plan.updated", AffectedAppID: "app-b"},
				{ID: "3", CreatedAt: t3, ActivityType: "site_admin.app.plan.updated", AffectedAppID: "app-a"},
			}}
			svc := makeService(store)
			result, err := svc.ListAuditLogs(ctxWithSession(), ListAuditLogsParams{
				Page: 1, PageSize: 10, AffectedAppID: "app-a",
			})

			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 2)
			So(result.Entries[0].AffectedAppID, ShouldEqual, "app-a")
			So(result.Entries[1].AffectedAppID, ShouldEqual, "app-a")
		})

		Convey("ListAuditLogs: page 2 returns correct slice", func() {
			store := &fakeAuditLogStore{rows: []AuditLogEntry{
				{ID: "1", CreatedAt: t1, ActivityType: "site_admin.app.plan.updated"},
				{ID: "2", CreatedAt: t2, ActivityType: "site_admin.app.plan.updated"},
				{ID: "3", CreatedAt: t3, ActivityType: "site_admin.app.plan.updated"},
			}}
			svc := makeService(store)
			result, err := svc.ListAuditLogs(ctxWithSession(), ListAuditLogsParams{
				Page: 2, PageSize: 2, // page 2 of 2-per-page = 3rd entry
			})

			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 3)
			So(result.Entries, ShouldHaveLength, 1)
			So(result.Entries[0].ID, ShouldEqual, "1") // oldest (3rd in desc order)
		})

		Convey("ListAuditLogs: nil AuditDatabase returns empty", func() {
			svc := &SiteAdminAuditReadService{AuditDatabase: nil, Store: &fakeAuditLogStore{}}
			result, err := svc.ListAuditLogs(ctxWithSession(), ListAuditLogsParams{Page: 1, PageSize: 10})

			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 0)
			So(result.Entries, ShouldResemble, []AuditLogEntry{})
		})

		Convey("GetAuditLog: found returns entry with Data", func() {
			store := &fakeAuditLogStore{rows: []AuditLogEntry{
				{ID: "abc123", CreatedAt: t1, ActivityType: "site_admin.app.plan.updated",
					ActorUserID: "user-1", AffectedAppID: "app-x"},
			}}
			svc := makeService(store)
			entry, err := svc.GetAuditLog(ctxWithSession(), "abc123")

			So(err, ShouldBeNil)
			So(entry.ID, ShouldEqual, "abc123")
			So(entry.ActorUserID, ShouldEqual, "user-1")
			So(entry.AffectedAppID, ShouldEqual, "app-x")
			So(entry.Data, ShouldNotBeNil)
		})

		Convey("GetAuditLog: not found returns NotFound error", func() {
			svc := makeService(&fakeAuditLogStore{})
			_, err := svc.GetAuditLog(ctxWithSession(), "nonexistent")

			So(err, ShouldNotBeNil)
			So(apierrors.IsAPIErrorWithCondition(err, func(e *apierrors.APIError) bool {
				return e.Code == 404
			}), ShouldBeTrue)
		})

		Convey("GetAuditLog: nil AuditDatabase returns NotFound", func() {
			svc := &SiteAdminAuditReadService{AuditDatabase: nil, Store: &fakeAuditLogStore{}}
			_, err := svc.GetAuditLog(ctxWithSession(), "abc123")

			So(err, ShouldNotBeNil)
			So(apierrors.IsAPIErrorWithCondition(err, func(e *apierrors.APIError) bool {
				return e.Code == 404
			}), ShouldBeTrue)
		})
	})
}
