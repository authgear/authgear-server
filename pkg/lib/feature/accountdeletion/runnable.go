package accountdeletion

import (
	"context"
	"log/slog"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

type AppContextResolver interface {
	ResolveContext(ctx context.Context, appID string, fn func(context.Context, *config.AppContext) error) error
}

type UserService interface {
	DeleteFromScheduledDeletion(ctx context.Context, userID string) error
}

type UserServiceFactory interface {
	MakeUserService(appID string, appContext *config.AppContext) UserService
}

var RunnableLogger = slogutil.NewLogger("account-deletion-runner")

type Runnable struct {
	Store              *Store
	AppContextResolver AppContextResolver
	UserServiceFactory UserServiceFactory
}

func (r *Runnable) Run(ctx context.Context) error {
	appUsers, err := r.Store.ListAppUsers(ctx)
	if err != nil {
		return err
	}
	for _, appUser := range appUsers {
		err = r.AppContextResolver.ResolveContext(ctx, appUser.AppID, func(ctx context.Context, appCtx *config.AppContext) error {
			userService := r.UserServiceFactory.MakeUserService(appUser.AppID, appCtx)
			err = userService.DeleteFromScheduledDeletion(ctx, appUser.UserID)
			if err != nil {
				return err
			}
			logger := RunnableLogger.GetLogger(ctx)
			logger.Info(ctx, "executed scheduled account deletion",
				slog.String("app_id", appUser.AppID),
				slog.String("user_id", appUser.UserID),
			)
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}
