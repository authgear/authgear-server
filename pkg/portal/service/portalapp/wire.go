//go:build wireinject
// +build wireinject

package portalapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/google/wire"
)

func NewPortalAppService(
	appProvider *deps.AppProvider,
	request *http.Request) PortalAppService {
	panic(wire.Build(
		DependencySet,
		wire.Struct(new(PortalAppService), "*"),
	))
}
