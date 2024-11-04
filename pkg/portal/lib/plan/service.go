package plan

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/model"
)

type Service struct {
	GlobalDatabase *globaldb.Handle
	PlanStore      *Store
	AppConfig      *portalconfig.AppConfig
}

func (s *Service) GetDefaultPlan(ctx context.Context) (*model.Plan, error) {
	defaultPlanName := s.AppConfig.DefaultPlan
	if defaultPlanName == "" {
		// no default plan is configured
		return nil, nil
	}

	var plan *model.Plan
	var err error
	err = s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		plan, err = s.PlanStore.GetPlan(ctx, defaultPlanName)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return plan, nil
}

func (s *Service) ListPlans(ctx context.Context) ([]*model.Plan, error) {
	var plans []*model.Plan
	var err error
	err = s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		plans, err = s.PlanStore.List(ctx)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return plans, nil
}
