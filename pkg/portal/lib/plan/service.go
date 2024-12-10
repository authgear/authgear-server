package plan

import (
	"context"

	configplan "github.com/authgear/authgear-server/pkg/lib/config/plan"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
)

type Service struct {
	GlobalDatabase *globaldb.Handle
	PlanStore      *configplan.Store
	AppConfig      *portalconfig.AppConfig
}

func (s *Service) GetDefaultPlan(ctx context.Context) (*configplan.Plan, error) {
	defaultPlanName := s.AppConfig.DefaultPlan
	if defaultPlanName == "" {
		// no default plan is configured
		return nil, nil
	}

	var plan *configplan.Plan
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

func (s *Service) ListPlans(ctx context.Context) ([]*configplan.Plan, error) {
	var plans []*configplan.Plan
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
