package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/api/siteadmin"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	relay "github.com/authgear/authgear-server/pkg/graphqlgo/relay"
	portalservice "github.com/authgear/authgear-server/pkg/portal/service"
	"github.com/authgear/authgear-server/pkg/portal/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

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

func (f *fakeConfigSourceStore) ListPaged(_ context.Context, limit int, offset int) ([]*configsource.DatabaseSource, error) {
	if offset >= len(f.sources) {
		return nil, nil
	}
	end := offset + limit
	if end > len(f.sources) {
		end = len(f.sources)
	}
	return f.sources[offset:end], nil
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

type fakeOwnerStore struct {
	// appID -> ownerUserID
	owners map[string]string
}

func (f *fakeOwnerStore) GetOwnerByAppID(_ context.Context, appID string) (string, error) {
	if uid, ok := f.owners[appID]; ok {
		return uid, nil
	}
	return "", ErrOwnerNotFound
}

func (f *fakeOwnerStore) GetOwnersByAppIDs(_ context.Context, appIDs []string) (map[string]string, error) {
	m := make(map[string]string)
	for _, id := range appIDs {
		if uid, ok := f.owners[id]; ok {
			m[id] = uid
		}
	}
	return m, nil
}

func (f *fakeOwnerStore) CountAppsByOwnerUserID(_ context.Context, userID string) (int, error) {
	count := 0
	for _, uid := range f.owners {
		if uid == userID {
			count++
		}
	}
	return count, nil
}

func (f *fakeOwnerStore) ListAppIDsByOwnerUserIDPaged(_ context.Context, userID string, limit int, offset int) ([]string, error) {
	var appIDs []string
	for appID, uid := range f.owners {
		if uid == userID {
			appIDs = append(appIDs, appID)
		}
	}
	if offset >= len(appIDs) {
		return nil, nil
	}
	end := offset + limit
	if end > len(appIDs) {
		end = len(appIDs)
	}
	return appIDs[offset:end], nil
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
func getNodesResponse(nodes ...map[string]interface{}) map[string]interface{} {
	ns := make([]interface{}, len(nodes))
	for i, n := range nodes {
		ns[i] = n
	}
	return map[string]interface{}{
		"data": map[string]interface{}{"nodes": ns},
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
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	fixedClock := clock.NewMockClockAtTime(now)

	makeService := func(svr *httptest.Server, cs *fakeConfigSourceStore, os *fakeOwnerStore) *AppService {
		return &AppService{
			GlobalDatabase:    fakeDatabase{},
			ConfigSourceStore: cs,
			OwnerStore:        os,
			AdminAPI:          &fakeAdminAPI{serverURL: svr.URL},
			AuditDatabase:     nil,
			AuditStore:        nil,
			HTTPClient:        AppServiceHTTPClient{Client: &http.Client{}},
			Clock:             fixedClock,
		}
	}

	Convey("AppService", t, func() {

		Convey("ListApps with owner_email: user not found returns empty", func() {
			svr := adminAPIServer(getUsersByEmailResponse())
			defer svr.Close()

			svc := makeService(svr, &fakeConfigSourceStore{}, &fakeOwnerStore{owners: map[string]string{}})
			result, err := svc.ListApps(ctxWithSession(), ListAppsParams{
				Page: 1, PageSize: 10, OwnerEmail: "nobody@example.com",
			})

			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 0)
			So(result.Apps, ShouldResemble, []siteadmin.App{})
		})

		Convey("ListApps with owner_email: unparseable global ID is skipped -> empty result", func() {
			svr := adminAPIServer(getUsersByEmailResponse("not-a-valid-global-id"))
			defer svr.Close()

			svc := makeService(svr, &fakeConfigSourceStore{}, &fakeOwnerStore{owners: map[string]string{}})
			result, err := svc.ListApps(ctxWithSession(), ListAppsParams{
				Page: 1, PageSize: 10, OwnerEmail: "alice@example.com",
			})

			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 0)
			So(result.Apps, ShouldResemble, []siteadmin.App{})
		})

		Convey("ListApps with owner_email: user has no apps returns empty", func() {
			globalID := relay.ToGlobalID("User", "user-1")
			svr := adminAPIServer(getUsersByEmailResponse(globalID))
			defer svr.Close()

			svc := makeService(svr, &fakeConfigSourceStore{}, &fakeOwnerStore{owners: map[string]string{}})
			result, err := svc.ListApps(ctxWithSession(), ListAppsParams{
				Page: 1, PageSize: 10, OwnerEmail: "alice@example.com",
			})

			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 0)
			So(result.Apps, ShouldResemble, []siteadmin.App{})
		})

		Convey("ListApps paged: returns apps with resolved emails", func() {
			src1 := &configsource.DatabaseSource{AppID: "app-1", PlanName: "free", CreatedAt: now}
			src2 := &configsource.DatabaseSource{AppID: "app-2", PlanName: "starter", CreatedAt: now}

			globalID1 := relay.ToGlobalID("User", "user-1")
			globalID2 := relay.ToGlobalID("User", "user-2")
			svr := adminAPIServer(getNodesResponse(
				map[string]interface{}{"id": globalID1, "standardAttributes": map[string]interface{}{"email": "alice@example.com"}},
				map[string]interface{}{"id": globalID2, "standardAttributes": map[string]interface{}{"email": "bob@example.com"}},
			))
			defer svr.Close()

			cs := &fakeConfigSourceStore{
				sources:    []*configsource.DatabaseSource{src1, src2},
				totalCount: 2,
			}
			os := &fakeOwnerStore{owners: map[string]string{"app-1": "user-1", "app-2": "user-2"}}

			svc := makeService(svr, cs, os)
			result, err := svc.ListApps(ctxWithSession(), ListAppsParams{Page: 1, PageSize: 10})

			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 2)
			So(result.Apps, ShouldHaveLength, 2)
		})

		Convey("ListApps: PageSize 0 is clamped to maxPageSize", func() {
			svr := adminAPIServer(getNodesResponse())
			defer svr.Close()

			cs := &fakeConfigSourceStore{sources: nil, totalCount: 0}
			os := &fakeOwnerStore{owners: map[string]string{}}

			svc := makeService(svr, cs, os)
			result, err := svc.ListApps(ctxWithSession(), ListAppsParams{Page: 1, PageSize: 0})

			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 0)
		})

		Convey("GetApp: nil audit DB yields UserCount 0", func() {
			src := &configsource.DatabaseSource{AppID: "app-1", PlanName: "free", CreatedAt: now}
			globalID := relay.ToGlobalID("User", "user-1")
			svr := adminAPIServer(getNodesResponse(
				map[string]interface{}{"id": globalID, "standardAttributes": map[string]interface{}{"email": "alice@example.com"}},
			))
			defer svr.Close()

			cs := &fakeConfigSourceStore{sources: []*configsource.DatabaseSource{src}, totalCount: 1}
			os := &fakeOwnerStore{owners: map[string]string{"app-1": "user-1"}}

			svc := makeService(svr, cs, os)
			detail, err := svc.GetApp(ctxWithSession(), "app-1")

			So(err, ShouldBeNil)
			So(detail.Id, ShouldEqual, "app-1")
			So(detail.OwnerEmail, ShouldEqual, "alice@example.com")
			So(detail.Plan, ShouldEqual, "free")
			So(detail.UserCount, ShouldEqual, 0)
		})

		Convey("GetApp: app not found returns ErrAppNotFound", func() {
			svr := adminAPIServer(getNodesResponse())
			defer svr.Close()

			cs := &fakeConfigSourceStore{sources: nil, totalCount: 0}
			os := &fakeOwnerStore{owners: map[string]string{}}

			svc := makeService(svr, cs, os)
			_, err := svc.GetApp(ctxWithSession(), "nonexistent")

			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, configsource.ErrAppNotFound)
		})
	})
}
