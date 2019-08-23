package gateway

import (
	"context"
	"net/http"

	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

type DependencyMap struct{}

// nolint: golint
func (m DependencyMap) Provide(
	dependencyName string,
	request *http.Request,
	ctx context.Context,
	requestID string,
	tConfig config.TenantConfiguration,
) interface{} {
	switch dependencyName {
	case "SessionProvider":
		return coreAuth.NewDefaultSessionProvider(ctx, tConfig)
	case "AuthInfoStore":
		return coreAuth.NewDefaultAuthInfoStore(ctx, tConfig)
	case "TxContext":
		return db.NewTxContextWithContext(ctx, tConfig)
	default:
		return nil
	}
}
