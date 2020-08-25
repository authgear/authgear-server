package service

import (
	"encoding/base64"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type ConfigGetter interface {
	GetConfig() (*config.Config, error)
}

type AppService struct {
	ConfigGetter ConfigGetter
}

func (s *AppService) GetMany(ids []string) (out []*model.App, err error) {
	cfg, err := s.ConfigGetter.GetConfig()
	if err != nil {
		return
	}

	for _, id := range ids {
		if id == string(cfg.AppConfig.ID) {
			out = append(out, &model.App{
				ID:           id,
				AppConfig:    cfg.AppConfig,
				SecretConfig: cfg.SecretConfig,
			})
		}
	}

	return
}

func (s *AppService) Count() (uint64, error) {
	return 1, nil
}

func (s *AppService) QueryPage(after, before graphqlutil.Cursor, first, last *uint64) (out []graphqlutil.PageItem, err error) {
	cfg, err := s.ConfigGetter.GetConfig()
	if err != nil {
		return
	}

	// FIXME(portal): we ignore the arguments here and always return 1 item.
	val := &model.App{
		ID:           string(cfg.AppConfig.ID),
		AppConfig:    cfg.AppConfig,
		SecretConfig: cfg.SecretConfig,
	}
	out = append(out, graphqlutil.PageItem{
		Value:  val,
		Cursor: graphqlutil.Cursor(base64.RawURLEncoding.EncodeToString([]byte(val.ID))),
	})
	return
}
