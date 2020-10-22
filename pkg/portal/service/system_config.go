package service

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/model"
	portalresource "github.com/authgear/authgear-server/pkg/portal/resource"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

type ResourceManager interface {
	Read(desc resource.Descriptor, args map[string]interface{}) (*resource.MergedFile, error)
}

type SystemConfigProvider struct {
	AuthgearConfig *config.AuthgearConfig
	AppConfig      *config.AppConfig
	Resources      ResourceManager
}

func (p *SystemConfigProvider) SystemConfig() (*model.SystemConfig, error) {
	themes, err := p.loadJSON(portalresource.ThemesJSON)
	if err != nil {
		return nil, err
	}

	translations, err := p.loadJSON(portalresource.TranslationsJSON)
	if err != nil {
		return nil, err
	}

	return &model.SystemConfig{
		AuthgearClientID: p.AuthgearConfig.ClientID,
		AuthgearEndpoint: p.AuthgearConfig.Endpoint,
		AppHostSuffix:    p.AppConfig.HostSuffix,
		Themes:           themes,
		Translations:     translations,
	}, nil
}

func (p *SystemConfigProvider) loadJSON(desc resource.Descriptor) (interface{}, error) {
	json, err := p.Resources.Read(desc, nil)
	if errors.Is(err, resource.ErrResourceNotFound) {
		// Omit the JSON if resource not configured.
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	data, err := desc.Parse(json)
	if err != nil {
		return nil, err
	}

	return data, nil
}
