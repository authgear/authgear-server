//+build wireinject

package handler

import (
	"net/http"

	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

// ProvideTenantConfig provides pointer to TenantConfiguration
// Some existing code requires the struct, so in later refactoring
// we have to convert the dependency to use pointer first.
func ProvideTenantConfig(r *http.Request) *config.TenantConfiguration {
	return config.GetTenantConfig(r.Context())
}

var DefaultSet = wire.NewSet(
	ProvideTenantConfig,
)

// TODO(authui): Delete Foobar

type Foobar struct{}

func NewFoobar(tConfig *config.TenantConfiguration) *Foobar {
	return &Foobar{}
}

func InjectFoobar(r *http.Request) *Foobar {
	wire.Build(DefaultSet, NewFoobar)
	return &Foobar{}
}
