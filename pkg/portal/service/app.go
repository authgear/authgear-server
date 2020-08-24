package service

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"

	libconfig "github.com/authgear/authgear-server/pkg/lib/config"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type AppService struct {
	Config *libconfig.Config
}

func NewLibConfig(serverConfig *portalconfig.ServerConfig) (c *libconfig.Config, err error) {
	appConfigYAML, err := ioutil.ReadFile(serverConfig.ConfigSource.AppConfigPath)
	if err != nil {
		err = fmt.Errorf("cannot read app config file: %w", err)
		return
	}
	appConfig, err := libconfig.Parse(appConfigYAML)
	if err != nil {
		err = fmt.Errorf("cannot parse app config: %w", err)
		return
	}

	secretConfigYAML, err := ioutil.ReadFile(serverConfig.ConfigSource.SecretConfigPath)
	if err != nil {
		err = fmt.Errorf("cannot read secret config file: %w", err)
		return
	}

	secretConfig, err := libconfig.ParseSecret(secretConfigYAML)
	if err != nil {
		err = fmt.Errorf("cannot parse secret config: %w", err)
		return
	}

	if err = secretConfig.Validate(appConfig); err != nil {
		err = fmt.Errorf("invalid secret config: %w", err)
		return
	}

	c = &libconfig.Config{
		AppConfig:    appConfig,
		SecretConfig: secretConfig,
	}
	return
}

func (s *AppService) GetMany(ids []string) (out []*model.App, err error) {
	for _, id := range ids {
		if id == string(s.Config.AppConfig.ID) {
			out = append(out, &model.App{
				ID:           id,
				AppConfig:    s.Config.AppConfig,
				SecretConfig: s.Config.SecretConfig,
			})
		}
	}

	return
}

func (s *AppService) Count() (uint64, error) {
	return 1, nil
}

func (s *AppService) QueryPage(after, before graphqlutil.Cursor, first, last *uint64) (out []graphqlutil.PageItem, err error) {
	// FIXME(portal): we ignore the arguments here and always return 1 item.
	val := &model.App{
		ID:           string(s.Config.AppConfig.ID),
		AppConfig:    s.Config.AppConfig,
		SecretConfig: s.Config.SecretConfig,
	}
	out = append(out, graphqlutil.PageItem{
		Value:  val,
		Cursor: graphqlutil.Cursor(base64.RawURLEncoding.EncodeToString([]byte(val.ID))),
	})
	return
}
