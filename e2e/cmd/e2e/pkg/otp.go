package e2e

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/util/secretcode"
)

func (c *End2End) GetLinkOTPCode(ctx context.Context, appID string, claimName string, claimValue string) (string, error) {
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		return "", err
	}
	cfg.ConfigSource = &configsource.Config{
		Type:  configsource.TypeDatabase,
		Watch: false,
	}

	p, err := deps.NewRootProvider(
		ctx,
		cfg.EnvironmentConfig,
		cfg.ConfigSource,
		cfg.CustomResourceDirectory,
	)
	if err != nil {
		return "", err
	}

	configSrcController := newConfigSourceController(p)
	err = configSrcController.Open(ctx)
	if err != nil {
		return "", err
	}
	defer configSrcController.Close()

	appCtx, err := configSrcController.ResolveContext(ctx, appID)
	if err != nil {
		return "", err
	}

	appProvider := p.NewAppProvider(ctx, appCtx)

	loginIDService := newLoginIDSerivce(appProvider)

	var loginIDs []*identity.LoginID
	err = appProvider.AppDatabase.ReadOnly(ctx, func(ctx context.Context) (err error) {
		loginIDs, err = loginIDService.ListByClaim(ctx, claimName, claimValue)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	if len(loginIDs) != 1 {
		return "", fmt.Errorf("claim not found")
	}

	otpCode := secretcode.LinkOTPSecretCode.GenerateDeterministic(loginIDs[0].UserID)

	return otpCode, nil
}
