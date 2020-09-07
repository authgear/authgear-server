package service

import (
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/portal/model"
)

type AppAuthzService interface {
	ListAuthorizedApps(userID string) ([]string, error)
}

type AppService struct {
	ConfigSource *configsource.ConfigSource
	AppAuthz     AppAuthzService
}

func (s *AppService) GetMany(ids []string) (out []*model.App, err error) {
	for _, id := range ids {
		appCtx, err := s.ConfigSource.ContextResolver.ResolveContext(id)
		if err != nil {
			return nil, err
		}
		out = append(out, &model.App{
			ID:      id,
			Context: appCtx,
		})
	}

	return
}

func (s *AppService) List(userID string) ([]*model.App, error) {
	appIDs, err := s.AppAuthz.ListAuthorizedApps(userID)
	if err != nil {
		return nil, err
	}

	return s.GetMany(appIDs)
}

func (s *AppService) UpdateConfig(app *model.App, updateFiles []*model.AppConfigFile, deleteFiles []string) error {
	// TODO(portal): validate & update files
	fmt.Printf("%v %#v %#v\n", app, updateFiles, deleteFiles)
	return errors.New("??e")
}
