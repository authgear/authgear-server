package service

import (
	"context"
	"errors"
	"maps"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/config/plan"
	"github.com/authgear/authgear-server/pkg/portal/deps"
	siteadminauditlog "github.com/authgear/authgear-server/pkg/siteadmin/auditlog"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

// ---- Fakes -------------------------------------------------------------------

type fakeFeatureConfigPlanStore struct {
	plans map[string]*plan.Plan
}

func (f *fakeFeatureConfigPlanStore) GetPlan(_ context.Context, name string) (*plan.Plan, error) {
	if p, ok := f.plans[name]; ok {
		return p, nil
	}
	return nil, plan.ErrPlanNotFound
}

type fakeFeatureConfigConfigSourceStore struct {
	sources map[string]*configsource.DatabaseSource
	updated *configsource.DatabaseSource
}

func (f *fakeFeatureConfigConfigSourceStore) GetDatabaseSourceByAppID(_ context.Context, appID string) (*configsource.DatabaseSource, error) {
	if s, ok := f.sources[appID]; ok {
		cp := *s
		cp.Data = maps.Clone(s.Data) // deep-enough copy so mutation in the service doesn't corrupt the fixture between test cases
		return &cp, nil
	}
	return nil, configsource.ErrAppNotFound
}

func (f *fakeFeatureConfigConfigSourceStore) UpdateDatabaseSource(_ context.Context, dbs *configsource.DatabaseSource) error {
	f.updated = dbs
	return nil
}

// ---- Test helpers --------------------------------------------------------------

func emptyBaseResources() deps.AppBaseResources {
	return deps.AppBaseResources(resource.NewManager(resource.DefaultRegistry, nil))
}

// ---- Tests -------------------------------------------------------------------

func TestFeatureConfigService_GetAppFeatureConfig(t *testing.T) {
	updatedAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	makeService := func(planStore *fakeFeatureConfigPlanStore, csStore *fakeFeatureConfigConfigSourceStore) *FeatureConfigService {
		return &FeatureConfigService{
			GlobalDatabase:    fakeDatabase{},
			PlanStore:         planStore,
			ConfigSourceStore: csStore,
			BaseResources:     emptyBaseResources(),
			Clock:             clock.NewMockClockAtTime(updatedAt),
		}
	}

	Convey("GetAppFeatureConfig", t, func() {
		planStore := &fakeFeatureConfigPlanStore{
			plans: map[string]*plan.Plan{
				"test-plan": {
					Name: "test-plan",
					RawFeatureConfig: &config.FeatureConfig{
						Collaborator: &config.CollaboratorFeatureConfig{Maximum: new(3)},
						OAuth: &config.OAuthFeatureConfig{
							Client: &config.OAuthClientFeatureConfig{Maximum: new(10)},
						},
					},
				},
			},
		}

		Convey("app has no override", func() {
			csStore := &fakeFeatureConfigConfigSourceStore{
				sources: map[string]*configsource.DatabaseSource{
					"app1": {AppID: "app1", PlanName: "test-plan", Data: map[string][]byte{}},
				},
			}
			svc := makeService(planStore, csStore)

			result, err := svc.GetAppFeatureConfig(context.Background(), "app1")
			So(err, ShouldBeNil)
			So(result.AppFeatureConfigYAML, ShouldEqual, "")
			So(*result.EffectiveAppFeatureConfig.Collaborator.Maximum, ShouldEqual, 3)
			So(result.EffectiveAppFeatureConfig, ShouldResemble, result.EffectivePlanFeatureConfig)
		})

		Convey("app has an override with a comment", func() {
			overrideYAML := "# a comment\ncollaborator:\n  soft_maximum: 1\n"
			csStore := &fakeFeatureConfigConfigSourceStore{
				sources: map[string]*configsource.DatabaseSource{
					"app1": {AppID: "app1", PlanName: "test-plan", Data: map[string][]byte{
						featureConfigOverrideKey(): []byte(overrideYAML),
					}},
				},
			}
			svc := makeService(planStore, csStore)

			result, err := svc.GetAppFeatureConfig(context.Background(), "app1")
			So(err, ShouldBeNil)
			So(result.AppFeatureConfigYAML, ShouldEqual, overrideYAML)
		})

		Convey("collaborator fixture: effective maximum inherited from plan, soft_maximum from override", func() {
			csStore := &fakeFeatureConfigConfigSourceStore{
				sources: map[string]*configsource.DatabaseSource{
					"app1": {AppID: "app1", PlanName: "test-plan", Data: map[string][]byte{
						featureConfigOverrideKey(): []byte("collaborator:\n  soft_maximum: 1\n"),
					}},
				},
			}
			svc := makeService(planStore, csStore)

			result, err := svc.GetAppFeatureConfig(context.Background(), "app1")
			So(err, ShouldBeNil)
			So(*result.EffectiveAppFeatureConfig.Collaborator.Maximum, ShouldEqual, 3)
			So(*result.EffectiveAppFeatureConfig.Collaborator.SoftMaximum, ShouldEqual, 1)
		})

		Convey("oauth.client fixture: effective maximum inherited from plan, custom_ui_enabled from override", func() {
			csStore := &fakeFeatureConfigConfigSourceStore{
				sources: map[string]*configsource.DatabaseSource{
					"app1": {AppID: "app1", PlanName: "test-plan", Data: map[string][]byte{
						featureConfigOverrideKey(): []byte("oauth:\n  client:\n    custom_ui_enabled: true\n"),
					}},
				},
			}
			svc := makeService(planStore, csStore)

			result, err := svc.GetAppFeatureConfig(context.Background(), "app1")
			So(err, ShouldBeNil)
			So(*result.EffectiveAppFeatureConfig.OAuth.Client.Maximum, ShouldEqual, 10)
			So(*result.EffectiveAppFeatureConfig.OAuth.Client.CustomUIEnabled, ShouldBeTrue)
		})

		Convey("app references a plan that does not exist", func() {
			csStore := &fakeFeatureConfigConfigSourceStore{
				sources: map[string]*configsource.DatabaseSource{
					"app1": {AppID: "app1", PlanName: "nonexistent-plan", Data: map[string][]byte{}},
				},
			}
			svc := makeService(planStore, csStore)

			result, err := svc.GetAppFeatureConfig(context.Background(), "app1")
			So(err, ShouldBeNil)
			So(result.EffectivePlanFeatureConfig, ShouldResemble, config.NewEffectiveDefaultFeatureConfig())
		})

		Convey("app does not exist", func() {
			csStore := &fakeFeatureConfigConfigSourceStore{sources: map[string]*configsource.DatabaseSource{}}
			svc := makeService(planStore, csStore)

			_, err := svc.GetAppFeatureConfig(context.Background(), "missing-app")
			So(err, ShouldNotBeNil)
			So(apierrors.IsAPIErrorWithCondition(err, func(e *apierrors.APIError) bool {
				return e.Kind.Name == apierrors.NotFound
			}), ShouldBeTrue)
		})
	})
}

func TestFeatureConfigService_UpdateAppFeatureConfig(t *testing.T) {
	updatedAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	planStore := &fakeFeatureConfigPlanStore{
		plans: map[string]*plan.Plan{
			"test-plan": {
				Name: "test-plan",
				RawFeatureConfig: &config.FeatureConfig{
					Collaborator: &config.CollaboratorFeatureConfig{Maximum: new(3)},
				},
			},
		},
	}

	makeService := func(csStore *fakeFeatureConfigConfigSourceStore) *FeatureConfigService {
		return &FeatureConfigService{
			GlobalDatabase:    fakeDatabase{},
			PlanStore:         planStore,
			ConfigSourceStore: csStore,
			BaseResources:     emptyBaseResources(),
			Clock:             clock.NewMockClockAtTime(updatedAt),
		}
	}

	Convey("UpdateAppFeatureConfig", t, func() {
		csStore := &fakeFeatureConfigConfigSourceStore{
			sources: map[string]*configsource.DatabaseSource{
				"app1": {AppID: "app1", PlanName: "test-plan", Data: map[string][]byte{
					featureConfigOverrideKey(): []byte("collaborator:\n  soft_maximum: 1\n"),
				}},
			},
		}

		Convey("valid YAML is stored byte-for-byte", func() {
			svc := makeService(csStore)
			newYAML := "# comment\ncollaborator:\n  soft_maximum: 2\n"

			result, err := svc.UpdateAppFeatureConfig(context.Background(), "app1", newYAML)
			So(err, ShouldBeNil)
			So(csStore.updated, ShouldNotBeNil)
			So(string(csStore.updated.Data[featureConfigOverrideKey()]), ShouldEqual, newYAML)
			So(*result.EffectiveAppFeatureConfig.Collaborator.SoftMaximum, ShouldEqual, 2)
			So(*result.EffectiveAppFeatureConfig.Collaborator.Maximum, ShouldEqual, 3)
		})

		Convey("empty string clears an existing override", func() {
			svc := makeService(csStore)

			_, err := svc.UpdateAppFeatureConfig(context.Background(), "app1", "")
			So(err, ShouldBeNil)
			So(csStore.updated, ShouldNotBeNil)
			_, ok := csStore.updated.Data[featureConfigOverrideKey()]
			So(ok, ShouldBeFalse)
		})

		Convey("whitespace-only string clears an existing override", func() {
			svc := makeService(csStore)

			_, err := svc.UpdateAppFeatureConfig(context.Background(), "app1", "  \n\t")
			So(err, ShouldBeNil)
			So(csStore.updated, ShouldNotBeNil)
			_, ok := csStore.updated.Data[featureConfigOverrideKey()]
			So(ok, ShouldBeFalse)
		})

		Convey("invalid YAML - unknown top-level key", func() {
			svc := makeService(csStore)

			_, err := svc.UpdateAppFeatureConfig(context.Background(), "app1", "nonexistent_top_level_key: 1\n")
			So(err, ShouldNotBeNil)
			var aggErr *validation.AggregatedError
			So(errors.As(err, &aggErr), ShouldBeTrue)
			So(aggErr.Errors[0].Location, ShouldEqual, "/nonexistent_top_level_key")
			So(csStore.updated, ShouldBeNil)
		})

		Convey("invalid YAML - field-level type error", func() {
			svc := makeService(csStore)

			_, err := svc.UpdateAppFeatureConfig(context.Background(), "app1", "oauth:\n  client:\n    maximum: \"not-a-number\"\n")
			So(err, ShouldNotBeNil)
			var aggErr *validation.AggregatedError
			So(errors.As(err, &aggErr), ShouldBeTrue)
			So(aggErr.Errors[0].Location, ShouldEqual, "/oauth/client/maximum")
			So(aggErr.Errors[0].Keyword, ShouldEqual, "type")
			So(csStore.updated, ShouldBeNil)
		})

		Convey("multi-document YAML is rejected without causes", func() {
			svc := makeService(csStore)

			_, err := svc.UpdateAppFeatureConfig(context.Background(), "app1", "a: 1\n---\nb: 2\n")
			So(err, ShouldNotBeNil)
			So(apierrors.IsAPIErrorWithCondition(err, func(e *apierrors.APIError) bool {
				return e.Kind.Name == apierrors.Invalid && e.Kind.Reason == "ValidationFailed"
			}), ShouldBeTrue)
			var aggErr *validation.AggregatedError
			So(errors.As(err, &aggErr), ShouldBeFalse)
			So(csStore.updated, ShouldBeNil)
		})

		Convey("app does not exist", func() {
			svc := makeService(csStore)

			_, err := svc.UpdateAppFeatureConfig(context.Background(), "missing-app", "collaborator:\n  soft_maximum: 1\n")
			So(err, ShouldNotBeNil)
			So(apierrors.IsAPIErrorWithCondition(err, func(e *apierrors.APIError) bool {
				return e.Kind.Name == apierrors.NotFound
			}), ShouldBeTrue)
			So(csStore.updated, ShouldBeNil)
		})

		Convey("emits audit log with old and new YAML", func() {
			svc := makeService(csStore)
			audit := &fakeAuditService{}
			svc.AuditService = audit

			newYAML := "collaborator:\n  soft_maximum: 2\n"
			_, err := svc.UpdateAppFeatureConfig(context.Background(), "app1", newYAML)
			So(err, ShouldBeNil)
			So(audit.logged, ShouldHaveLength, 1)
			payload, ok := audit.logged[0].(*siteadminauditlog.AppFeatureConfigUpdatedPayload)
			So(ok, ShouldBeTrue)
			So(payload.AppID, ShouldEqual, "app1")
			So(payload.OldAppFeatureConfigYAML, ShouldEqual, "collaborator:\n  soft_maximum: 1\n")
			So(payload.NewAppFeatureConfigYAML, ShouldEqual, newYAML)
		})

		Convey("audit failure does not affect mutation result", func() {
			svc := makeService(csStore)
			// AuditService is nil — mutation must still succeed.
			result, err := svc.UpdateAppFeatureConfig(context.Background(), "app1", "collaborator:\n  soft_maximum: 2\n")
			So(err, ShouldBeNil)
			So(*result.EffectiveAppFeatureConfig.Collaborator.SoftMaximum, ShouldEqual, 2)
		})
	})
}

func TestFeatureConfigService_PreviewAppFeatureConfig(t *testing.T) {
	updatedAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	planStore := &fakeFeatureConfigPlanStore{
		plans: map[string]*plan.Plan{
			"test-plan": {
				Name: "test-plan",
				RawFeatureConfig: &config.FeatureConfig{
					Collaborator: &config.CollaboratorFeatureConfig{Maximum: new(3)},
				},
			},
		},
	}

	makeService := func(csStore *fakeFeatureConfigConfigSourceStore) *FeatureConfigService {
		return &FeatureConfigService{
			GlobalDatabase:    fakeDatabase{},
			PlanStore:         planStore,
			ConfigSourceStore: csStore,
			BaseResources:     emptyBaseResources(),
			Clock:             clock.NewMockClockAtTime(updatedAt),
		}
	}

	Convey("PreviewAppFeatureConfig", t, func() {
		csStore := &fakeFeatureConfigConfigSourceStore{
			sources: map[string]*configsource.DatabaseSource{
				"app1": {AppID: "app1", PlanName: "test-plan", Data: map[string][]byte{}},
			},
		}

		Convey("valid candidate matches what PUT would produce, nothing persisted", func() {
			svc := makeService(csStore)

			result, err := svc.PreviewAppFeatureConfig(context.Background(), "app1", "collaborator:\n  soft_maximum: 1\n")
			So(err, ShouldBeNil)
			So(*result.EffectiveAppFeatureConfig.Collaborator.SoftMaximum, ShouldEqual, 1)
			So(*result.EffectiveAppFeatureConfig.Collaborator.Maximum, ShouldEqual, 3)
			So(csStore.updated, ShouldBeNil)
		})

		Convey("invalid candidate - field-level type error", func() {
			svc := makeService(csStore)

			_, err := svc.PreviewAppFeatureConfig(context.Background(), "app1", "oauth:\n  client:\n    maximum: \"not-a-number\"\n")
			So(err, ShouldNotBeNil)
			var aggErr *validation.AggregatedError
			So(errors.As(err, &aggErr), ShouldBeTrue)
			So(aggErr.Errors[0].Location, ShouldEqual, "/oauth/client/maximum")
			So(aggErr.Errors[0].Keyword, ShouldEqual, "type")
			So(csStore.updated, ShouldBeNil)
		})

		Convey("app does not exist", func() {
			svc := makeService(csStore)

			_, err := svc.PreviewAppFeatureConfig(context.Background(), "missing-app", "collaborator:\n  soft_maximum: 1\n")
			So(err, ShouldNotBeNil)
			So(apierrors.IsAPIErrorWithCondition(err, func(e *apierrors.APIError) bool {
				return e.Kind.Name == apierrors.NotFound
			}), ShouldBeTrue)
		})
	})
}

func TestCountYAMLDocuments(t *testing.T) {
	Convey("countYAMLDocuments", t, func() {
		Convey("0 docs (empty input)", func() {
			n, err := countYAMLDocuments([]byte(""))
			So(err, ShouldBeNil)
			So(n, ShouldEqual, 0)
		})

		Convey("1 doc", func() {
			n, err := countYAMLDocuments([]byte("a: 1\n"))
			So(err, ShouldBeNil)
			So(n, ShouldEqual, 1)
		})

		Convey("2 docs separated by ---", func() {
			n, err := countYAMLDocuments([]byte("a: 1\n---\nb: 2\n"))
			So(err, ShouldBeNil)
			So(n, ShouldEqual, 2)
		})
	})
}
