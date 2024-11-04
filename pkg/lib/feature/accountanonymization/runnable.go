package accountanonymization

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type AppContextResolver interface {
	ResolveContext(ctx context.Context, appID string) (*config.AppContext, error)
}

type UserService interface {
	AnonymizeFromScheduledAnonymization(ctx context.Context, userID string) error
}

type UserServiceFactory interface {
	MakeUserService(appID string, appContext *config.AppContext) UserService
}

type RunnableLogger struct{ *log.Logger }

func NewRunnableLogger(lf *log.Factory) RunnableLogger {
	return RunnableLogger{lf.New("account-anonymization-runner")}
}

type Runnable struct {
	Store              *Store
	AppContextResolver AppContextResolver
	UserServiceFactory UserServiceFactory
	Logger             RunnableLogger
}

func (r *Runnable) Run(ctx context.Context) error {
	appUsers, err := r.Store.ListAppUsers(ctx)
	if err != nil {
		return err
	}
	for _, appUser := range appUsers {
		appContext, err := r.AppContextResolver.ResolveContext(ctx, appUser.AppID)
		if err != nil {
			return err
		}
		userService := r.UserServiceFactory.MakeUserService(appUser.AppID, appContext)
		err = userService.AnonymizeFromScheduledAnonymization(ctx, appUser.UserID)
		if err != nil {
			return err
		}
		r.Logger.WithFields(map[string]interface{}{
			"app_id":  appUser.AppID,
			"user_id": appUser.UserID,
		}).Infof("executed scheduled account anonymization")
	}
	return nil
}
