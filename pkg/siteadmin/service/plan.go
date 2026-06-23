package service

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/siteadmin"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/config/plan"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

// Narrow interfaces

type PlanServiceAuditService interface {
	LogEvent(ctx context.Context, appID string, payload event.NonBlockingPayload) error
}

type PlanServiceGlobalDatabase interface {
	WithTx(ctx context.Context, do func(ctx context.Context) error) error
	ReadOnly(ctx context.Context, do func(ctx context.Context) error) error
}

type PlanServicePlanStore interface {
	GetPlan(ctx context.Context, name string) (*plan.Plan, error)
	List(ctx context.Context) ([]*plan.Plan, error)
}

type PlanServiceConfigSourceStore interface {
	GetDatabaseSourceByAppID(ctx context.Context, appID string) (*configsource.DatabaseSource, error)
	UpdateDatabaseSource(ctx context.Context, dbs *configsource.DatabaseSource) error
}

type PlanServiceOwnerStore interface {
	GetOwnerByAppID(ctx context.Context, appID string) (string, error)
}

// PlanService

type PlanService struct {
	GlobalDatabase    PlanServiceGlobalDatabase
	PlanStore         PlanServicePlanStore
	ConfigSourceStore PlanServiceConfigSourceStore
	OwnerStore        PlanServiceOwnerStore
	AdminAPI          *AdminAPIService
	AuditService      PlanServiceAuditService
	Clock             clock.Clock
}

func (s *PlanService) ListPlans(ctx context.Context) ([]siteadmin.Plan, error) {
	var plans []*plan.Plan
	err := s.GlobalDatabase.ReadOnly(ctx, func(ctx context.Context) error {
		var e error
		plans, e = s.PlanStore.List(ctx)
		return e
	})
	if err != nil {
		return nil, err
	}
	result := make([]siteadmin.Plan, len(plans))
	for i, p := range plans {
		result[i] = siteadmin.Plan{Name: p.Name}
	}
	return result, nil
}

func (s *PlanService) ChangeAppPlan(ctx context.Context, appID string, planName string) (*siteadmin.App, error) {
	// Verify plan exists, update config source, and look up owner — all in one transaction.
	var dbs *configsource.DatabaseSource
	var oldPlanName string
	var ownerUserID string
	err := s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		_, e := s.PlanStore.GetPlan(ctx, planName)
		if errors.Is(e, plan.ErrPlanNotFound) {
			return apierrors.NotFound.WithReason("PlanNotFound").New("plan not found")
		}
		if e != nil {
			return e
		}

		dbs, e = s.ConfigSourceStore.GetDatabaseSourceByAppID(ctx, appID)
		if errors.Is(e, configsource.ErrAppNotFound) {
			return apierrors.NotFound.WithReason("AppNotFound").New("app not found")
		}
		if e != nil {
			return e
		}
		oldPlanName = dbs.PlanName
		dbs.PlanName = planName
		dbs.UpdatedAt = s.Clock.NowUTC()
		if e = s.ConfigSourceStore.UpdateDatabaseSource(ctx, dbs); e != nil {
			return e
		}

		ownerUserID, e = s.OwnerStore.GetOwnerByAppID(ctx, appID)
		if errors.Is(e, ErrOwnerNotFound) {
			return nil
		}
		return e
	})
	if err != nil {
		return nil, err
	}

	// Resolve owner email — outside the DB transaction (Admin API call).
	ownerEmail := ""
	if ownerUserID != "" {
		emailMap, err := s.AdminAPI.ResolveUserEmails(ctx, []string{ownerUserID})
		if err != nil {
			return nil, err
		}
		ownerEmail = emailMap[ownerUserID]
	}

	if s.AuditService != nil {
		if err := s.AuditService.LogEvent(ctx, appID, &nonblocking.SiteAdminAppPlanUpdatedEventPayload{
			AppID:   appID,
			OldPlan: oldPlanName,
			NewPlan: planName,
		}); err != nil {
			AuditServiceLogger.GetLogger(ctx).WithError(err).Error(ctx, "failed to emit site admin audit log")
		}
	}

	return &siteadmin.App{
		Id:           appID,
		Plan:         planName,
		CreatedAt:    dbs.CreatedAt,
		OwnerEmail:   ownerEmail,
		LastMonthMau: 0,
	}, nil
}
