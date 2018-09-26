package auth

import (
	"context"

	"github.com/skygeario/skygear-server/pkg/auth/dependency"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/server/audit"
)

type DependencyMap struct{}

func NewDependencyMap() DependencyMap {
	return DependencyMap{}
}

func (m DependencyMap) Provide(dependencyName string, ctx context.Context, tConfig config.TenantConfiguration) interface{} {
	switch dependencyName {
	case "TokenStore":
		return coreAuth.NewDefaultTokenStore(ctx, tConfig)
	case "AuthInfoStore":
		return coreAuth.NewDefaultAuthInfoStore(ctx, tConfig)
	case "AuthDataChecker":
		return &dependency.DefaultAuthDataChecker{
			//TODO:
			// from tConfig
			AuthRecordKeys: [][]string{[]string{"email"}, []string{"username"}},
		}
	case "PasswordChecker":
		return &audit.PasswordChecker{
			// TODO:
			// from tConfig
			PwMinLength: 6,
		}
	default:
		return nil
	}
}
