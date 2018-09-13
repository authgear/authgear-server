package db

import (
	"context"

	"github.com/skygeario/skygear-server/pkg/gateway/model"
	"github.com/skygeario/skygear-server/pkg/gateway/db/pq"
)

// NewGatewayStore create new gateway store by db connection url
func NewGatewayStore(ctx context.Context, connString string) (GatewayStore, error) {
	return pq.Connect(ctx, connString)
}

// GatewayStore provide functions to query application info from config db
type GatewayStore interface {

	// GetAppByDomain fetches the App with domain
	GetAppByDomain(domain string, app *model.App) error

	Close() error
}
