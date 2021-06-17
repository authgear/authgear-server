package service

import (
	"encoding/json"
	"errors"

	"github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/model"
	portalresource "github.com/authgear/authgear-server/pkg/portal/resource"
	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

type ResourceManager interface {
	Read(desc resource.Descriptor, view resource.View) (interface{}, error)
}

type SystemConfigProvider struct {
	AuthgearConfig *config.AuthgearConfig
	AppConfig      *config.AppConfig
	SearchConfig   *config.SearchConfig
	AuditLogConfig *config.AuditLogConfig
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
		AuthgearClientID:   p.AuthgearConfig.ClientID,
		AuthgearEndpoint:   p.AuthgearConfig.Endpoint,
		AppHostSuffix:      p.AppConfig.HostSuffix,
		AvailableLanguages: intl.AvailableLanguages,
		BuiltinLanguages:   intl.BuiltinLanguages,
		Themes:             themes,
		Translations:       translations,
		SearchEnabled:      p.SearchConfig.Enabled,
		AuditLogEnabled:    p.AuditLogConfig.Enabled,
	}, nil
}

func (p *SystemConfigProvider) loadJSON(desc resource.Descriptor) (interface{}, error) {
	result, err := p.Resources.Read(desc, resource.EffectiveResource{})
	if errors.Is(err, resource.ErrResourceNotFound) {
		// Omit the JSON if resource not configured.
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	bytes := result.([]byte)

	var data interface{}

	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
