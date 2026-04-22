package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/siteadmin"
	relay "github.com/authgear/authgear-server/pkg/graphqlgo/relay"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/config/plan"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

// ---- Fakes -------------------------------------------------------------------

type fakePlanStore struct {
	plans []*plan.Plan
}

func (f *fakePlanStore) GetPlan(_ context.Context, name string) (*plan.Plan, error) {
	for _, p := range f.plans {
		if p.Name == name {
			return p, nil
		}
	}
	return nil, plan.ErrPlanNotFound
}

func (f *fakePlanStore) List(_ context.Context) ([]*plan.Plan, error) {
	return f.plans, nil
}

type fakePlanConfigSourceStore struct {
	sources map[string]*configsource.DatabaseSource
	updated *configsource.DatabaseSource
}

func (f *fakePlanConfigSourceStore) GetDatabaseSourceByAppID(_ context.Context, appID string) (*configsource.DatabaseSource, error) {
	if s, ok := f.sources[appID]; ok {
		// Return a copy so the test can check mutations.
		cp := *s
		return &cp, nil
	}
	return nil, configsource.ErrAppNotFound
}

func (f *fakePlanConfigSourceStore) UpdateDatabaseSource(_ context.Context, dbs *configsource.DatabaseSource) error {
	f.updated = dbs
	return nil
}

type fakePlanOwnerStore struct {
	owners map[string]string // appID → userID
}

func (f *fakePlanOwnerStore) GetOwnerByAppID(_ context.Context, appID string) (string, error) {
	if uid, ok := f.owners[appID]; ok {
		return uid, nil
	}
	return "", ErrOwnerNotFound
}

// ---- Tests -------------------------------------------------------------------

func TestPlanService_ListPlans(t *testing.T) {
	Convey("ListPlans", t, func() {
		svc := &PlanService{
			GlobalDatabase: fakeDatabase{},
			PlanStore: &fakePlanStore{
				plans: []*plan.Plan{
					{Name: "free"},
					{Name: "enterprise"},
				},
			},
		}

		Convey("returns all plans", func() {
			result, err := svc.ListPlans(context.Background())
			So(err, ShouldBeNil)
			So(result, ShouldResemble, []siteadmin.Plan{
				{Name: "free"},
				{Name: "enterprise"},
			})
		})

		Convey("returns empty slice when no plans", func() {
			svc.PlanStore = &fakePlanStore{plans: nil}
			result, err := svc.ListPlans(context.Background())
			So(err, ShouldBeNil)
			So(result, ShouldHaveLength, 0)
		})
	})
}

func TestPlanService_ChangeAppPlan(t *testing.T) {
	createdAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	makeService := func(svr *httptest.Server, planStore *fakePlanStore, csStore *fakePlanConfigSourceStore, ownerStore *fakePlanOwnerStore) *PlanService {
		return &PlanService{
			GlobalDatabase:    fakeDatabase{},
			PlanStore:         planStore,
			ConfigSourceStore: csStore,
			OwnerStore:        ownerStore,
			AdminAPI: &AdminAPIService{
				AdminAPI:   &fakeAdminAPI{serverURL: svr.URL},
				HTTPClient: SiteAdminHTTPClient{Client: &http.Client{}},
			},
			Clock: clock.NewMockClockAtTime(createdAt),
		}
	}

	Convey("ChangeAppPlan", t, func() {
		planStore := &fakePlanStore{plans: []*plan.Plan{{Name: "enterprise"}}}
		csStore := &fakePlanConfigSourceStore{
			sources: map[string]*configsource.DatabaseSource{
				"app1": {AppID: "app1", PlanName: "free", CreatedAt: createdAt},
			},
		}

		Convey("returns NotFound when plan does not exist", func() {
			svr := adminAPIServer(getNodesResponse())
			defer svr.Close()

			svc := makeService(svr, planStore, csStore, &fakePlanOwnerStore{})
			_, err := svc.ChangeAppPlan(ctxWithSession(), "app1", "nonexistent")
			So(err, ShouldNotBeNil)
			So(apierrors.IsAPIErrorWithCondition(err, func(e *apierrors.APIError) bool {
				return e.Kind.Name == apierrors.NotFound
			}), ShouldBeTrue)
		})

		Convey("returns NotFound when app does not exist", func() {
			svr := adminAPIServer(getNodesResponse())
			defer svr.Close()

			svc := makeService(svr, planStore, csStore, &fakePlanOwnerStore{})
			_, err := svc.ChangeAppPlan(ctxWithSession(), "missing-app", "enterprise")
			So(err, ShouldNotBeNil)
			So(apierrors.IsAPIErrorWithCondition(err, func(e *apierrors.APIError) bool {
				return e.Kind.Name == apierrors.NotFound
			}), ShouldBeTrue)
		})

		Convey("updates plan_name on config source", func() {
			svr := adminAPIServer(getNodesResponse())
			defer svr.Close()

			svc := makeService(svr, planStore, csStore, &fakePlanOwnerStore{})
			_, err := svc.ChangeAppPlan(ctxWithSession(), "app1", "enterprise")
			So(err, ShouldBeNil)
			So(csStore.updated, ShouldNotBeNil)
			So(csStore.updated.PlanName, ShouldEqual, "enterprise")
		})

		Convey("returns App with new plan, no owner", func() {
			svr := adminAPIServer(getNodesResponse())
			defer svr.Close()

			svc := makeService(svr, planStore, csStore, &fakePlanOwnerStore{})
			app, err := svc.ChangeAppPlan(ctxWithSession(), "app1", "enterprise")
			So(err, ShouldBeNil)
			So(app, ShouldResemble, &siteadmin.App{
				Id:           "app1",
				Plan:         "enterprise",
				CreatedAt:    createdAt,
				OwnerEmail:   "",
				LastMonthMau: 0,
			})
		})

		Convey("returns App with owner email resolved", func() {
			globalID := relay.ToGlobalID("User", "user1")
			svr := adminAPIServer(getNodesResponse(
				map[string]interface{}{
					"id":                 globalID,
					"standardAttributes": map[string]interface{}{"email": "owner@example.com"},
				},
			))
			defer svr.Close()

			ownerStore := &fakePlanOwnerStore{owners: map[string]string{"app1": "user1"}}
			svc := makeService(svr, planStore, csStore, ownerStore)
			app, err := svc.ChangeAppPlan(ctxWithSession(), "app1", "enterprise")
			So(err, ShouldBeNil)
			So(app.OwnerEmail, ShouldEqual, "owner@example.com")
			So(app.Plan, ShouldEqual, "enterprise")
		})
	})
}
