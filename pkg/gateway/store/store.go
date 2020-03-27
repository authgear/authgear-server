package store

import (
	"github.com/skygeario/skygear-server/pkg/gateway/model"
)

// GatewayStore provide functions to query application info from config db
type GatewayStore interface {
	// GetDomain fetch domain record
	GetDomain(domain string) (*model.Domain, error)

	// GetDefaultDomain fetch default domain record
	GetDefaultDomain(domain string) (*model.Domain, error)

	// GetDomainByAppIDAndAssignment return app default domain
	GetDomainByAppIDAndAssignment(appID string, assignment model.AssignmentType) (*model.Domain, error)

	// GetApp fetch app by id
	GetApp(id string) (*model.App, error)

	// GetLastDeploymentRoutes return all routes of last deployment
	GetLastDeploymentRoutes(app model.App) ([]*model.DeploymentRoute, error)

	// GetLastDeploymentHooks return all hooks of last deployment
	GetLastDeploymentHooks(app model.App) (*model.DeploymentHooks, error)

	Close() error
}
