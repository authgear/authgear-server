package handler

import (
	"context"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

type Factory interface {
	NewHandler(context.Context, config.TenantConfiguration) Handler
}

type ProviderGraph interface {
	Provide(name string, ctx context.Context, configuration config.TenantConfiguration) interface{}
}
