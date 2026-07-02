package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
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
	end := min(offset+limit, uint64(len(f.sources)))
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
	// collaborators maps appID → all collaborator user IDs (all roles).
	// When set for an app, overrides OwnerUserID for filter matching and
	// relevance ranking. When not set, falls back to OwnerUserID only.
	collaborators map[string][]string
}

func (f *fakeOwnerStore) GetOwnerByAppID(_ context.Context, appID string) (string, error) {
	if uid, ok := f.owners[appID]; ok {
		return uid, nil
	}
	return "", ErrOwnerNotFound
}

//nolint:gocognit
func (f *fakeOwnerStore) ListAppsWithStats(_ context.Context, params ListAppsStoreParams) ([]AppStoreRow, int, error) {
	// Filter
	var filtered []AppStoreRow
	for _, r := range f.rows {
		if params.AppID != "" && !strings.HasPrefix(r.AppID, params.AppID) {
			continue
		}
		if params.PlanName != "" && r.PlanName != params.PlanName {
			continue
		}
		if len(params.CollaboratorUserIDs) > 0 {
			// Match any collaborator (any role) — mirrors EXISTS subquery in real SQL.
			matched := false
			if collabs, ok := f.collaborators[r.AppID]; ok {
				for _, uid := range collabs {
					if slices.Contains(params.CollaboratorUserIDs, uid) {
						matched = true
						break
					}
				}
			} else {
				matched = slices.Contains(params.CollaboratorUserIDs, r.OwnerUserID)
			}
			if !matched {
				continue
			}
		}
		filtered = append(filtered, r)
	}

	// Sort — mirrors SQL: primary field with configurable direction, app_id ASC secondary.
	// When sort=Relevance, best collaborator rank is prepended as the primary key.
	sort.SliceStable(filtered, func(i, j int) bool {
		a, b := filtered[i], filtered[j]
		if params.Sort == siteadmin.Relevance {
			ra := f.bestRankFor(a, params.CollaboratorUserIDs)
			rb := f.bestRankFor(b, params.CollaboratorUserIDs)
			if ra != rb {
				return ra < rb
			}
		}
		if params.Sort == siteadmin.Mau {
			if a.LastMonthMAU != b.LastMonthMAU {
				if params.Order == siteadmin.Asc {
					return a.LastMonthMAU < b.LastMonthMAU
				}
				return a.LastMonthMAU > b.LastMonthMAU
			}
		} else {
			if !a.CreatedAt.Equal(b.CreatedAt) {
				if params.Order == siteadmin.Asc {
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
	end := min(offset+int(params.PageSize), total)
	return filtered[offset:end], total, nil
}

// bestRankFor returns the best (lowest) position of any of the app's collaborators
// within collaboratorUserIDs, mirroring MIN(array_position(...)) in real SQL.
// Falls back to checking OwnerUserID when no collaborators map entry exists.
func (f *fakeOwnerStore) bestRankFor(row AppStoreRow, collaboratorUserIDs []string) int {
	sentinel := len(collaboratorUserIDs)
	if collabs, ok := f.collaborators[row.AppID]; ok {
		best := sentinel
		for _, uid := range collabs {
			if pos := slices.Index(collaboratorUserIDs, uid); pos >= 0 && pos < best {
				best = pos
			}
		}
		return best
	}
	pos := slices.Index(collaboratorUserIDs, row.OwnerUserID)
	if pos < 0 {
		return sentinel
	}
	return pos
}

// fakeDatabase satisfies AppServiceDatabase by directly executing the callback
// without a real DB transaction — suitable for unit tests that use in-memory fakes.
type fakeDatabase struct{}

func (fakeDatabase) WithTx(ctx context.Context, do func(context.Context) error) error {
	return do(ctx)
}

func (fakeDatabase) ReadOnly(ctx context.Context, do func(context.Context) error) error {
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
func adminAPIServer(response any) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
}

// multiResponseAdminAPIServer starts a test HTTP server that serves each response
// in order. After the slice is exhausted, the last response is repeated.
func multiResponseAdminAPIServer(responses ...any) *httptest.Server {
	var mu sync.Mutex
	idx := 0
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		var resp any
		if idx < len(responses) {
			resp = responses[idx]
			idx++
		} else {
			resp = responses[len(responses)-1]
		}
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
}

// getUsersByEmailResponse builds a GraphQL data envelope for the
// getUsersByStandardAttribute query.
func getUsersByEmailResponse(globalIDs ...string) map[string]any {
	users := make([]any, len(globalIDs))
	for i, id := range globalIDs {
		users[i] = map[string]any{"id": id}
	}
	return map[string]any{
		"data": map[string]any{"users": users},
	}
}

// getNodesResponse builds a GraphQL data envelope for the getUserNodes query.
func getNodesResponse(nodes ...any) map[string]any {
	if nodes == nil {
		nodes = []any{}
	}
	return map[string]any{
		"data": map[string]any{"nodes": nodes},
	}
}

// searchUsersByKeywordResponse builds a GraphQL data envelope for the searchOwnersByKeyword query.
// hasNextPage controls the truncated flag.
func searchUsersByKeywordResponse(hasNextPage bool, globalIDs ...string) map[string]any {
	edges := make([]any, len(globalIDs))
	for i, id := range globalIDs {
		edges[i] = map[string]any{"node": map[string]any{"id": id}}
	}
	return map[string]any{
		"data": map[string]any{
			"users": map[string]any{
				"edges":    edges,
				"pageInfo": map[string]any{"hasNextPage": hasNextPage},
			},
		},
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

		Convey("ListApps with collaborator_search: no matching users returns empty", func() {
			svr := adminAPIServer(searchUsersByKeywordResponse(false))
			defer svr.Close()

			svc := makeService(svr, &fakeConfigSourceStore{}, &fakeOwnerStore{})
			result, err := svc.ListApps(ctxWithSession(), ListAppsParams{
				Page: 1, PageSize: 10, CollaboratorSearch: "nobody@example.com",
			})

			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 0)
			So(result.Apps, ShouldResemble, []siteadmin.App{})
			So(result.CollaboratorSearchTruncated, ShouldBeFalse)
		})

		Convey("ListApps with collaborator_search: found user but no apps returns empty", func() {
			globalID := relay.ToGlobalID("User", "user-1")
			svr := adminAPIServer(searchUsersByKeywordResponse(false, globalID))
			defer svr.Close()

			svc := makeService(svr, &fakeConfigSourceStore{}, &fakeOwnerStore{rows: []AppStoreRow{}})
			result, err := svc.ListApps(ctxWithSession(), ListAppsParams{
				Page: 1, PageSize: 10, CollaboratorSearch: "alice@example.com",
			})

			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 0)
			So(result.Apps, ShouldResemble, []siteadmin.App{})
		})

		Convey("ListApps with collaborator_search: truncated flag propagated when hasNextPage=true", func() {
			globalID := relay.ToGlobalID("User", "user-1")
			svr := multiResponseAdminAPIServer(
				searchUsersByKeywordResponse(true, globalID),
				getNodesResponse(),
			)
			defer svr.Close()

			os := &fakeOwnerStore{
				rows: []AppStoreRow{
					{AppID: "app-1", PlanName: "free", CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), OwnerUserID: "user-1"},
				},
			}
			svc := makeService(svr, &fakeConfigSourceStore{}, os)
			result, err := svc.ListApps(ctxWithSession(), ListAppsParams{
				Page: 1, PageSize: 10, CollaboratorSearch: "alice",
			})

			So(err, ShouldBeNil)
			So(result.CollaboratorSearchTruncated, ShouldBeTrue)
		})

		Convey("ListApps paged: returns apps with resolved emails and last_month_mau", func() {
			t1 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC) // newer
			t2 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC) // older

			globalID1 := relay.ToGlobalID("User", "user-1")
			globalID2 := relay.ToGlobalID("User", "user-2")
			svr := adminAPIServer(getNodesResponse(
				map[string]any{"id": globalID1, "standardAttributes": map[string]any{"email": "alice@example.com"}},
				map[string]any{"id": globalID2, "standardAttributes": map[string]any{"email": "bob@example.com"}},
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

		Convey("ListApps: plan + collaborator_search combined filter", func() {
			globalID := relay.ToGlobalID("User", "user-1")
			svr := multiResponseAdminAPIServer(
				searchUsersByKeywordResponse(false, globalID),
				getNodesResponse(),
			)
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
				Page: 1, PageSize: 10, Plan: "starter", CollaboratorSearch: "alice@example.com",
			})

			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 1)
			So(result.Apps[0].Id, ShouldEqual, "app-1")
		})

		Convey("ListApps: sort=relevance without collaborator_search returns error", func() {
			svr := adminAPIServer(getNodesResponse())
			defer svr.Close()

			svc := makeService(svr, &fakeConfigSourceStore{}, &fakeOwnerStore{})
			_, err := svc.ListApps(ctxWithSession(), ListAppsParams{
				Page: 1, PageSize: 10, Sort: siteadmin.Relevance,
			})

			So(err, ShouldNotBeNil)
			So(apierrors.IsAPIError(err), ShouldBeTrue)
		})

		Convey("ListApps: collaborator_search with sort=relevance orders by owner rank", func() {
			t1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

			// user-2 ranks higher (first in search result) than user-1.
			globalID1 := relay.ToGlobalID("User", "user-1")
			globalID2 := relay.ToGlobalID("User", "user-2")
			svr := multiResponseAdminAPIServer(
				searchUsersByKeywordResponse(false, globalID2, globalID1),
				getNodesResponse(),
			)
			defer svr.Close()

			os := &fakeOwnerStore{
				rows: []AppStoreRow{
					{AppID: "app-a", PlanName: "free", CreatedAt: t1, OwnerUserID: "user-1"},
					{AppID: "app-b", PlanName: "free", CreatedAt: t1, OwnerUserID: "user-2"},
				},
			}

			svc := makeService(svr, &fakeConfigSourceStore{}, os)
			result, err := svc.ListApps(ctxWithSession(), ListAppsParams{
				Page: 1, PageSize: 10, CollaboratorSearch: "alice", Sort: siteadmin.Relevance,
			})

			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 2)
			// app-b (user-2, rank 0) must come before app-a (user-1, rank 1)
			So(result.Apps[0].Id, ShouldEqual, "app-b")
			So(result.Apps[1].Id, ShouldEqual, "app-a")
		})

		Convey("ListApps: collaborator_search with sort=mau sorts by MAU not owner rank", func() {
			t1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

			// user-2 ranks higher in search but app-a (user-1) has higher MAU.
			globalID1 := relay.ToGlobalID("User", "user-1")
			globalID2 := relay.ToGlobalID("User", "user-2")
			svr := multiResponseAdminAPIServer(
				searchUsersByKeywordResponse(false, globalID2, globalID1),
				getNodesResponse(),
			)
			defer svr.Close()

			os := &fakeOwnerStore{
				rows: []AppStoreRow{
					{AppID: "app-a", PlanName: "free", CreatedAt: t1, OwnerUserID: "user-1", LastMonthMAU: 200},
					{AppID: "app-b", PlanName: "free", CreatedAt: t1, OwnerUserID: "user-2", LastMonthMAU: 50},
				},
			}

			svc := makeService(svr, &fakeConfigSourceStore{}, os)
			result, err := svc.ListApps(ctxWithSession(), ListAppsParams{
				Page: 1, PageSize: 10, CollaboratorSearch: "alice", Sort: siteadmin.Mau, Order: siteadmin.Desc,
			})

			So(err, ShouldBeNil)
			// MAU sort: app-a (200) before app-b (50), regardless of owner rank
			So(result.Apps[0].Id, ShouldEqual, "app-a")
			So(result.Apps[1].Id, ShouldEqual, "app-b")
		})

		Convey("ListApps: app_id prefix filter returns matching apps only", func() {
			t1 := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
			t2 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
			t3 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

			svr := adminAPIServer(getNodesResponse())
			defer svr.Close()

			os := &fakeOwnerStore{
				rows: []AppStoreRow{
					{AppID: "myapp-prod", PlanName: "free", CreatedAt: t1},
					{AppID: "myapp-staging", PlanName: "free", CreatedAt: t2},
					{AppID: "other-app", PlanName: "free", CreatedAt: t3},
				},
			}

			svc := makeService(svr, &fakeConfigSourceStore{}, os)
			result, err := svc.ListApps(ctxWithSession(), ListAppsParams{Page: 1, PageSize: 10, AppID: "myapp"})

			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 2)
			So(result.Apps, ShouldHaveLength, 2)
			So(result.Apps[0].Id, ShouldEqual, "myapp-prod")
			So(result.Apps[1].Id, ShouldEqual, "myapp-staging")
		})

		Convey("ListApps: app_id prefix filter with exact ID still matches", func() {
			t1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

			svr := adminAPIServer(getNodesResponse())
			defer svr.Close()

			os := &fakeOwnerStore{
				rows: []AppStoreRow{
					{AppID: "myapp", PlanName: "free", CreatedAt: t1},
					{AppID: "other-app", PlanName: "free", CreatedAt: t1},
				},
			}

			svc := makeService(svr, &fakeConfigSourceStore{}, os)
			result, err := svc.ListApps(ctxWithSession(), ListAppsParams{Page: 1, PageSize: 10, AppID: "myapp"})

			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 1)
			So(result.Apps[0].Id, ShouldEqual, "myapp")
		})

		Convey("ListApps: collaborator_search matches app by editor role", func() {
			// app-1 has owner="user-owner" and editor="user-editor".
			// Searching by "user-editor" (an editor, not owner) must return app-1.
			t1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
			globalEditorID := relay.ToGlobalID("User", "user-editor")
			svr := multiResponseAdminAPIServer(
				searchUsersByKeywordResponse(false, globalEditorID),
				getNodesResponse(), // email resolution (empty — not checking emails here)
			)
			defer svr.Close()

			os := &fakeOwnerStore{
				rows: []AppStoreRow{
					{AppID: "app-1", PlanName: "free", CreatedAt: t1, OwnerUserID: "user-owner"},
					{AppID: "app-2", PlanName: "free", CreatedAt: t1, OwnerUserID: "user-other"},
				},
				collaborators: map[string][]string{
					"app-1": {"user-owner", "user-editor"},
					"app-2": {"user-other"},
				},
			}
			svc := makeService(svr, &fakeConfigSourceStore{}, os)
			result, err := svc.ListApps(ctxWithSession(), ListAppsParams{
				Page: 1, PageSize: 10, CollaboratorSearch: "editor@example.com",
			})

			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 1)
			So(result.Apps[0].Id, ShouldEqual, "app-1")
		})

		Convey("ListApps: collaborator_search relevance ranks by best collaborator not just owner", func() {
			// user-2 ranks higher in search but is an editor of app-a, not the owner.
			// With collaborator_search, app-a should still rank first (best rank = 0 for user-2).
			t1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
			globalID1 := relay.ToGlobalID("User", "user-1")
			globalID2 := relay.ToGlobalID("User", "user-2")
			// user-2 (rank 0) is editor of app-a; user-1 (rank 1) is owner of app-b.
			svr := multiResponseAdminAPIServer(
				searchUsersByKeywordResponse(false, globalID2, globalID1),
				getNodesResponse(),
			)
			defer svr.Close()

			os := &fakeOwnerStore{
				rows: []AppStoreRow{
					{AppID: "app-a", PlanName: "free", CreatedAt: t1, OwnerUserID: "user-owner-a"},
					{AppID: "app-b", PlanName: "free", CreatedAt: t1, OwnerUserID: "user-1"},
				},
				collaborators: map[string][]string{
					"app-a": {"user-owner-a", "user-2"},
					"app-b": {"user-1"},
				},
			}
			svc := makeService(svr, &fakeConfigSourceStore{}, os)
			result, err := svc.ListApps(ctxWithSession(), ListAppsParams{
				Page: 1, PageSize: 10, CollaboratorSearch: "alice", Sort: siteadmin.Relevance,
			})

			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 2)
			// app-a: best rank = 0 (user-2 is rank 0); app-b: best rank = 1 (user-1 is rank 1)
			So(result.Apps[0].Id, ShouldEqual, "app-a")
			So(result.Apps[1].Id, ShouldEqual, "app-b")
		})

		Convey("GetApp: nil audit DB yields UserCount 0", func() {
			src := &configsource.DatabaseSource{AppID: "app-1", PlanName: "free", CreatedAt: now}
			globalID := relay.ToGlobalID("User", "user-1")
			svr := adminAPIServer(getNodesResponse(
				map[string]any{"id": globalID, "standardAttributes": map[string]any{"email": "alice@example.com"}},
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
