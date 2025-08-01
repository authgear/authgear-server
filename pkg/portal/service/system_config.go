package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	configlib "github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/model"
	portalresource "github.com/authgear/authgear-server/pkg/portal/resource"
	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/timeutil"
	"github.com/authgear/authgear-server/pkg/version"
)

type ResourceManager interface {
	Read(ctx context.Context, desc resource.Descriptor, view resource.View) (interface{}, error)
}

type SystemConfigProvider struct {
	AuthgearConfig                 *config.AuthgearConfig
	AppConfig                      *config.AppConfig
	SearchConfig                   *config.SearchConfig
	AuditLogConfig                 *config.AuditLogConfig
	AnalyticConfig                 *configlib.AnalyticConfig
	GTMConfig                      *config.GoogleTagManagerConfig
	FrontendSentryConfig           *config.PortalFrontendSentryConfig
	PortalFeaturesConfig           *config.PortalFeaturesConfig
	GlobalUIImplementation         configlib.GlobalUIImplementation
	GlobalUISettingsImplementation configlib.GlobalUISettingsImplementation
	Resources                      ResourceManager
}

func (p *SystemConfigProvider) SystemConfig(ctx context.Context) (*model.SystemConfig, error) {
	themes, err := p.loadJSON(ctx, portalresource.ThemesJSON)
	if err != nil {
		return nil, err
	}

	translations, err := p.loadJSON(ctx, portalresource.TranslationsJSON)
	if err != nil {
		return nil, err
	}

	var analyticEpoch *timeutil.Date
	if !p.AnalyticConfig.Epoch.IsZero() {
		analyticEpoch = &p.AnalyticConfig.Epoch
	}

	return &model.SystemConfig{
		AuthgearAppID:               p.AuthgearConfig.AppID,
		AuthgearClientID:            p.AuthgearConfig.ClientID,
		AuthgearEndpoint:            p.AuthgearConfig.Endpoint,
		AuthgearWebSDKSessionType:   p.AuthgearConfig.WebSDKSessionType,
		IsAuthgearOnce:              AUTHGEARONCE,
		AuthgearOnceLicenseKey:      p.AuthgearConfig.OnceLicenseKey,
		AuthgearOnceLicenseExpireAt: p.AuthgearConfig.OnceLicenseExpireAt,
		AuthgearOnceLicenseeEmail:   p.AuthgearConfig.OnceLicenseeEmail,
		SentryDSN:                   p.FrontendSentryConfig.DSN,
		AppHostSuffix:               p.AppConfig.HostSuffix,
		AvailableLanguages:          intl.AvailableLanguages,
		BuiltinLanguages:            intl.BuiltinLanguages,
		Themes:                      themes,
		Translations:                translations,
		SearchEnabled:               p.SearchConfig.Enabled,
		AuditLogEnabled:             p.AuditLogConfig.Enabled,
		AnalyticEnabled:             p.AnalyticConfig.Enabled,
		AnalyticEpoch:               analyticEpoch,
		GitCommitHash:               strings.TrimPrefix(version.Version, "git-"),
		GTMContainerID:              p.GTMConfig.ContainerID,
		UIImplementation:            string(p.GlobalUIImplementation),
		UISettingsImplementation:    string(p.GlobalUISettingsImplementation),
	}, nil
}

func (p *SystemConfigProvider) loadJSON(ctx context.Context, desc resource.Descriptor) (interface{}, error) {

	result, err := p.Resources.Read(ctx, desc, resource.EffectiveResource{})
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
