package plan

import (
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/model"
)

type Service struct {
	PlanStore *Store
	AppConfig *portalconfig.AppConfig
}

func (s *Service) GetDefaultPlan() (*model.Plan, error) {
	defaultPlanName := s.AppConfig.DefaultPlan
	if defaultPlanName == "" {
		// no default plan is configured
		return nil, nil
	}
	plan, err := s.PlanStore.GetPlan(defaultPlanName)
	if err != nil {
		return nil, err
	}
	return plan, nil

}
