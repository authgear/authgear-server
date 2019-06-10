package db

import (
	"github.com/skygeario/skygear-server/pkg/gateway/model"
)

// GatewayStore provide functions to query application info from config db
type GatewayStore interface {

	// GetAppByDomain fetches the App with domain
	GetAppByDomain(domain string, app *model.App) error

	// FindLongestMatchedCloudCode find the longest matched cloud code by the
	// given path
	FindLongestMatchedCloudCode(path string, app model.App, cloudCode *model.CloudCode) error

	// GetLastDeploymentRoutes return all routes of last deployment
	GetLastDeploymentRoutes(app model.App) ([]*model.DeploymentRoute, error)

	Close() error
}
