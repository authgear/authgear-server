package accountanonymization

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type AppContextResolver interface {
	ResolveContext(appID string) (*config.AppContext, error)
}

type UserService interface {
	AnonymizeFromScheduledAnonymization(userID string) error
}

type UserServiceFactory interface {
	MakeUserService(ctx context.Context, appID string, appContext *config.AppContext) UserService
}

type RunnableLogger struct{ *log.Logger }

func NewRunnableLogger(lf *log.Factory) RunnableLogger {
	return RunnableLogger{lf.New("account-anonymization-runner")}
}

type Runnable struct {
	Context            context.Context
	Store              *Store
	AppContextResolver AppContextResolver
	UserServiceFactory UserServiceFactory
	Logger             RunnableLogger
}

func (r *Runnable) Run() error {
	appUsers, err := r.Store.ListAppUsers()
	if err != nil {
		return err
	}
	for _, appUser := range appUsers {
		appContext, err := r.AppContextResolver.ResolveContext(appUser.AppID)
		if err != nil {
			return err
		}
		userService := r.UserServiceFactory.MakeUserService(r.Context, appUser.AppID, appContext)
		err = userService.AnonymizeFromScheduledAnonymization(appUser.UserID)
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
