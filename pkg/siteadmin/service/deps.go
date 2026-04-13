package service

import (
	"time"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func NewHTTPClient() AppServiceHTTPClient {
	return AppServiceHTTPClient{Client: httputil.NewExternalClient(5 * time.Second)}
}

var DependencySet = wire.NewSet(
	wire.Struct(new(AppOwnerStore), "*"),
	wire.Bind(new(AppServiceOwnerStore), new(*AppOwnerStore)),
	wire.Struct(new(AppService), "*"),
	NewHTTPClient,
)
