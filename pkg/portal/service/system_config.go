package service

import (
	"github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/model"
)

type SystemConfigProvider struct {
	AuthgearConfig *config.AuthgearConfig
	AppConfig      *config.AppConfig
}

func (p *SystemConfigProvider) SystemConfig() (*model.SystemConfig, error) {
	return &model.SystemConfig{
		AuthgearClientID: p.AuthgearConfig.ClientID,
		AuthgearEndpoint: p.AuthgearConfig.Endpoint,
		AppHostSuffix:    p.AppConfig.HostSuffix,
	}, nil
}
