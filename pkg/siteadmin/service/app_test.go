package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/api/siteadmin"
	relay "github.com/authgear/authgear-server/pkg/graphqlgo/relay"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	portalservice "github.com/authgear/authgear-server/pkg/portal/service"
	"github.com/authgear/authgear-server/pkg/portal/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

var _ = relay.ToGlobalID

// ---- Fakes -------------------------------------------------------------------

type fakeConfigSourceStore struct {
	sources    []*configsource.DatabaseSource
	totalCount int
}

func (f *fakeConfigSourceStore) GetDatabaseSourceByAppID(_ context.Context, appID string) (*configsource.DatabaseSource, error) {
	for _, s := range f.sources {
		if s.AppID == appID {
			return s, nil
		}
	}
	return nil, configsource.ErrAppNotFound
}

func (f *fakeConfigSourceStore) CountAll(_ context.Context) (int, error) {
	return f.totalCount, nil
}

func (f *fakeConfigSourceStore) ListPaged(_ context.Context, limit uint64, offset uint64) ([]*configsource.DatabaseSource, error) {
	if offset >= uint64(len(f.sources)) {
		return nil, nil
	}
	end := offset + limit
	if end > uint64(len(f.sources)) {
		end = uint64(len(f.sources))
	}
	return f.sources[int(offset):int(end)], nil
}

func (f *fakeConfigSourceStore) GetManyByAppIDs(_ context.Context, appIDs []string) ([]*configsource.DatabaseSource, error) {
	var result []*configsource.DatabaseSource
	for _, id := range appIDs {
		for _, s := range f.sources {
			if s.AppID == id {
				result = append(result, s)
				break
			}
		}
	}
	return result, nil
}

// fakeOwnerStore implements AppServiceOwnerStore using in-memory data.
// ListAppsWithStats mirrors real SQL semantics: filter, sort, paginate.
type fakeOwnerStore struct {
	// rows is the full set the fake "database" contains.
	rows []AppStoreRow
	// owners maps appID → ownerUserID (used only by GetOwnerByAppID).
	owners map[string]string
}

func (f *fakeOwnerStore) GetOwnerByAppID(_ context.Context, appID string) (string, error) {
	if uid, ok := f.owners[appID]; ok {
		return uid, nil
	}
	return "", ErrOwnerNotFound
}

func (f *fakeOwnerStore) ListAppsWithStats(_ context.Context, params ListAppsStoreParams) ([]AppStoreRow, int, error) {
	// Filter
	var filtered []AppStoreRow
	for _, r := range f.rows {
		if params.AppID != "" && r.AppID != params.AppID {
			continue
		}
		if params.PlanName != "" && r.PlanName != params.PlanName {
			continue
		}
		if params.OwnerUserID != "" && r.OwnerUserID != params.OwnerUserID {
			continue
		}
		filtered = append(filtered, r)
	}

	// Sort — mirrors SQL: primary field with configurable direction, app_id ASC secondary.
	sort.Slice(filtered, func(i, j int) bool {
		a, b := filtered[i], filtered[j]
		if params.Sort == "mau" {
			if a.LastMonthMAU != b.LastMonthMAU {
				if params.Order == "asc" {
					return a.LastMonthMAU < b.LastMonthMAU
				}
				return a.LastMonthMAU > b.LastMonthMAU
			}
		} else {
			if !a.CreatedAt.Equal(b.CreatedAt) {
				if params.Order == "asc" {
					return a.CreatedAt.Before(b.CreatedAt)
				}
				return a.CreatedAt.After(b.CreatedAt)
			}
		}
		return a.AppID < b.AppID // stable secondary (always ASC)
	})

	total := len(filtered)

	// Paginate
	offset := int((params.Page - 1) * params.PageSize)
	if offset >= total {
		return nil, total, nil
	}
	end := offset + int(params.PageSize)
	if end > total {
		end = total
	}
	return filtered[offset:end], total, nil
}

// fakeDatabase satisfies AppServiceDatabase by directly executing the callback
// without a real DB transaction — suitable for unit tests that use in-memory fakes.
type fakeDatabase struct{}

func (fakeDatabase) WithTx(ctx context.Context, do func(context.Context) error) error {
	return do(ctx)
}

type fakeAdminAPI struct {
	serverURL string
}

func (f *fakeAdminAPI) SelfDirector(_ context.Context, _ string, _ portalservice.Usage) (func(*http.Request), error) {
	return func(r *http.Request) {
		u, _ := url.Parse(f.serverURL)
		r.URL = u
		r.Host = u.Host
	}, nil
}

// ---- Helpers -----------------------------------------------------------------

// adminAPIServer starts a test HTTP server that always responds with the given
// value serialised as JSON.
func adminAPIServer(response interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
}

// getUsersByEmailResponse builds a GraphQL data envelope for the
// getUsersByStandardAttribute query.
func getUsersByEmailResponse(globalIDs ...string) map[string]interface{} {
	users := make([]interface{}, len(globalIDs))
	for i, id := range globalIDs {
		users[i] = map[string]interface{}{"id": id}
	}
	return map[string]interface{}{
		"data": map[string]interface{}{"users": users},
	}
}

// getNodesResponse builds a GraphQL data envelope for the getUserNodes query.
func getNodesResponse(nodes ...interface{}) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{"nodes": nodes},
	}
}

func ctxWithSession() context.Context {
	return session.WithSessionInfo(context.Background(), &model.SessionInfo{
		IsValid: true,
		UserID:  "actor-user",
	})
}

// ---- Tests -------------------------------------------------------------------

func TestAppService(t *testing.T) {
	now := time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC)
	fixedClock := clock.NewMockClockAtTime(now)

	makeService := func(svr *httptest.Server, cs *fakeConfigSourceStore, os *fakeOwnerStore) *AppService {
		return &AppService{
			GlobalDatabase:    fakeDatabase{},
			ConfigSourceStore: cs,
			OwnerStore:        os,
			AdminAPI: &AdminAPIService{
				AdminAPI:   &fakeAdminAPI{serverURL: svr.URL},
				HTTPClient: SiteAdminHTTPClient{Client: &http.Client{}},
			},
			AuditDatabase: nil,
			AuditStore:    nil,
			Clock:         fixedClock,
		}
	}

	Convey("AppService", t, func() {

		Convey("ListApps with owner_email: user not found returns empty", func() {
			svr := adminAPIServer(getUsersByEmailResponse())
			defer svr.Close()

			svc := makeService(svr, &fakeConfigSourceStore{}, &fakeOwnerStore{})
			result, err := svc.ListApps(ctxWithSession(), ListAppsParams{
				Page: 1, PageSize: 10, OwnerEmail: "nobody@example.com",
			})

			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 0)
			So(result.Apps, ShouldResemble, []siteadmin.App{})
		})

		Convey("ListApps with owner_email: found user but no apps returns empty", func() {
			globalID := relay.ToGlobalID("User", "user-1")
			svr := adminAPIServer(getUsersByEmailResponse(globalID))
			defer svr.Close()

			svc := makeService(svr, &fakeConfigSourceStore{}, &fakeOwnerStore{rows: []AppStoreRow{}})
			result, err := svc.ListApps(ctxWithSession(), ListAppsParams{
				Page: 1, PageSize: 10, OwnerEmail: "alice@example.com",
			})

			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 0)
			So(result.Apps, ShouldResemble, []siteadmin.App{})
		})

		Convey("ListApps paged: returns apps with resolved emails and last_month_mau", func() {
			t1 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC) // newer
			t2 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC) // older

			globalID1 := relay.ToGlobalID("User", "user-1")
			globalID2 := relay.ToGlobalID("User", "user-2")
			svr := adminAPIServer(getNodesResponse(
				map[string]interface{}{"id": globalID1, "standardAttributes": map[string]interface{}{"email": "alice@example.com"}},
				map[string]interface{}{"id": globalID2, "standardAttributes": map[string]interface{}{"email": "bob@example.com"}},
			))
			defer svr.Close()

			os := &fakeOwnerStore{
				rows: []AppStoreRow{
					{AppID: "app-1", PlanName: "free", CreatedAt: t1, OwnerUserID: "user-1", LastMonthMAU: 10},
					{AppID: "app-2", PlanName: "starter", CreatedAt: t2, OwnerUserID: "user-2", LastMonthMAU: 5},
				},
			}

			svc := makeService(svr, &fakeConfigSourceStore{}, os)
			result, err := svc.ListApps(ctxWithSession(), ListAppsParams{Page: 1, PageSize: 10})

			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 2)
			So(result.Apps, ShouldHaveLength, 2)
			// Default sort: created_at DESC → app-1 (t1=Jan 2) first
			So(result.Apps[0].Id, ShouldEqual, "app-1")
			So(result.Apps[0].LastMonthMau, ShouldEqual, 10)
			So(result.Apps[1].Id, ShouldEqual, "app-2")
			So(result.Apps[1].LastMonthMau, ShouldEqual, 5)
		})

		Convey("ListApps: PageSize 0 is clamped to maxPageSize", func() {
			svr := adminAPIServer(getNodesResponse())
			defer svr.Close()

			svc := makeService(svr, &fakeConfigSourceStore{}, &fakeOwnerStore{})
			result, err := svc.ListApps(ctxWithSession(), ListAppsParams{Page: 1, PageSize: 0})

			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 0)
		})

		Convey("ListApps: plan filter returns only matching apps", func() {
			t1 := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
			t2 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
			t3 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

			svr := adminAPIServer(getNodesResponse())
			defer svr.Close()

			os := &fakeOwnerStore{
				rows: []AppStoreRow{
					{AppID: "app-1", PlanName: "free", CreatedAt: t1},
					{AppID: "app-2", PlanName: "starter", CreatedAt: t2},
					{AppID: "app-3", PlanName: "free", CreatedAt: t3},
				},
			}

			svc := makeService(svr, &fakeConfigSourceStore{}, os)
			result, err := svc.ListApps(ctxWithSession(), ListAppsParams{Page: 1, PageSize: 10, Plan: "free"})

			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 2)
			So(result.Apps, ShouldHaveLength, 2)
			So(result.Apps[0].Id, ShouldEqual, "app-1")
			So(result.Apps[1].Id, ShouldEqual, "app-3")
		})

		Convey("ListApps: sort=mau order=desc returns highest MAU first", func() {
			t1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

			svr := adminAPIServer(getNodesResponse())
			defer svr.Close()

			os := &fakeOwnerStore{
				rows: []AppStoreRow{
					{AppID: "app-a", PlanName: "free", CreatedAt: t1, LastMonthMAU: 30},
					{AppID: "app-b", PlanName: "free", CreatedAt: t1, LastMonthMAU: 100},
					{AppID: "app-c", PlanName: "free", CreatedAt: t1, LastMonthMAU: 5},
				},
			}

			svc := makeService(svr, &fakeConfigSourceStore{}, os)
			result, err := svc.ListApps(ctxWithSession(), ListAppsParams{
				Page: 1, PageSize: 10, Sort: "mau", Order: "desc",
			})

			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 3)
			So(result.Apps[0].Id, ShouldEqual, "app-b") // MAU=100
			So(result.Apps[1].Id, ShouldEqual, "app-a") // MAU=30
			So(result.Apps[2].Id, ShouldEqual, "app-c") // MAU=5
		})

		Convey("ListApps: sort=mau order=asc returns lowest MAU first", func() {
			t1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

			svr := adminAPIServer(getNodesResponse())
			defer svr.Close()

			os := &fakeOwnerStore{
				rows: []AppStoreRow{
					{AppID: "app-a", PlanName: "free", CreatedAt: t1, LastMonthMAU: 30},
					{AppID: "app-b", PlanName: "free", CreatedAt: t1, LastMonthMAU: 100},
					{AppID: "app-c", PlanName: "free", CreatedAt: t1, LastMonthMAU: 5},
				},
			}

			svc := makeService(svr, &fakeConfigSourceStore{}, os)
			result, err := svc.ListApps(ctxWithSession(), ListAppsParams{
				Page: 1, PageSize: 10, Sort: "mau", Order: "asc",
			})

			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 3)
			So(result.Apps[0].Id, ShouldEqual, "app-c") // MAU=5
			So(result.Apps[1].Id, ShouldEqual, "app-a") // MAU=30
			So(result.Apps[2].Id, ShouldEqual, "app-b") // MAU=100
		})

		Convey("ListApps: MAU ties broken by app_id ASC", func() {
			t1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

			svr := adminAPIServer(getNodesResponse())
			defer svr.Close()

			os := &fakeOwnerStore{
				rows: []AppStoreRow{
					{AppID: "app-z", PlanName: "free", CreatedAt: t1, LastMonthMAU: 50},
					{AppID: "app-a", PlanName: "free", CreatedAt: t1, LastMonthMAU: 50},
					{AppID: "app-m", PlanName: "free", CreatedAt: t1, LastMonthMAU: 50},
				},
			}

			svc := makeService(svr, &fakeConfigSourceStore{}, os)
			result, err := svc.ListApps(ctxWithSession(), ListAppsParams{
				Page: 1, PageSize: 10, Sort: "mau", Order: "desc",
			})

			So(err, ShouldBeNil)
			So(result.Apps[0].Id, ShouldEqual, "app-a")
			So(result.Apps[1].Id, ShouldEqual, "app-m")
			So(result.Apps[2].Id, ShouldEqual, "app-z")
		})

		Convey("ListApps: sort=created_at order=asc returns oldest first", func() {
			t1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
			t2 := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
			t3 := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)

			svr := adminAPIServer(getNodesResponse())
			defer svr.Close()

			os := &fakeOwnerStore{
				rows: []AppStoreRow{
					{AppID: "app-1", PlanName: "free", CreatedAt: t1},
					{AppID: "app-2", PlanName: "free", CreatedAt: t2},
					{AppID: "app-3", PlanName: "free", CreatedAt: t3},
				},
			}

			svc := makeService(svr, &fakeConfigSourceStore{}, os)
			result, err := svc.ListApps(ctxWithSession(), ListAppsParams{
				Page: 1, PageSize: 10, Sort: "created_at", Order: "asc",
			})

			So(err, ShouldBeNil)
			So(result.Apps[0].Id, ShouldEqual, "app-1") // Jan
			So(result.Apps[1].Id, ShouldEqual, "app-3") // Feb
			So(result.Apps[2].Id, ShouldEqual, "app-2") // Mar
		})

		Convey("ListApps: last_month_mau is 0 for apps with no usage record", func() {
			t1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

			svr := adminAPIServer(getNodesResponse())
			defer svr.Close()

			os := &fakeOwnerStore{
				rows: []AppStoreRow{
					{AppID: "app-1", PlanName: "free", CreatedAt: t1, LastMonthMAU: 0},
				},
			}

			svc := makeService(svr, &fakeConfigSourceStore{}, os)
			result, err := svc.ListApps(ctxWithSession(), ListAppsParams{Page: 1, PageSize: 10})

			So(err, ShouldBeNil)
			So(result.Apps[0].LastMonthMau, ShouldEqual, 0)
		})

		Convey("ListApps: plan + owner_email combined filter", func() {
			globalID := relay.ToGlobalID("User", "user-1")
			svr := adminAPIServer(getUsersByEmailResponse(globalID))
			defer svr.Close()

			t1 := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
			t2 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
			t3 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

			os := &fakeOwnerStore{
				rows: []AppStoreRow{
					{AppID: "app-1", PlanName: "starter", CreatedAt: t1, OwnerUserID: "user-1"},
					{AppID: "app-2", PlanName: "free", CreatedAt: t2, OwnerUserID: "user-1"},
					{AppID: "app-3", PlanName: "starter", CreatedAt: t3, OwnerUserID: "user-2"},
				},
			}

			svc := makeService(svr, &fakeConfigSourceStore{}, os)
			result, err := svc.ListApps(ctxWithSession(), ListAppsParams{
				Page: 1, PageSize: 10, Plan: "starter", OwnerEmail: "alice@example.com",
			})

			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 1)
			So(result.Apps[0].Id, ShouldEqual, "app-1")
			So(result.Apps[0].OwnerEmail, ShouldEqual, "alice@example.com")
		})

		Convey("GetApp: nil audit DB yields UserCount 0", func() {
			src := &configsource.DatabaseSource{AppID: "app-1", PlanName: "free", CreatedAt: now}
			globalID := relay.ToGlobalID("User", "user-1")
			svr := adminAPIServer(getNodesResponse(
				map[string]interface{}{"id": globalID, "standardAttributes": map[string]interface{}{"email": "alice@example.com"}},
			))
			defer svr.Close()

			cs := &fakeConfigSourceStore{sources: []*configsource.DatabaseSource{src}}
			os := &fakeOwnerStore{owners: map[string]string{"app-1": "user-1"}}

			svc := makeService(svr, cs, os)
			detail, err := svc.GetApp(ctxWithSession(), "app-1")

			So(err, ShouldBeNil)
			So(detail.UserCount, ShouldEqual, 0)
			So(detail.OwnerEmail, ShouldEqual, "alice@example.com")
		})

		Convey("GetApp: no owner returns empty owner email", func() {
			src := &configsource.DatabaseSource{AppID: "app-1", PlanName: "free", CreatedAt: now}
			svr := adminAPIServer(getNodesResponse())
			defer svr.Close()

			cs := &fakeConfigSourceStore{sources: []*configsource.DatabaseSource{src}}
			os := &fakeOwnerStore{}

			svc := makeService(svr, cs, os)
			detail, err := svc.GetApp(ctxWithSession(), "app-1")

			So(err, ShouldBeNil)
			So(detail.OwnerEmail, ShouldEqual, "")
		})
	})
}

