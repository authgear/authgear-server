package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"

	goyaml "gopkg.in/yaml.v3"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/config/plan"
	"github.com/authgear/authgear-server/pkg/portal/deps"
	siteadminauditlog "github.com/authgear/authgear-server/pkg/siteadmin/auditlog"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/filepathutil"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

// ---- Narrow interfaces -------------------------------------------------------

type FeatureConfigServiceGlobalDatabase interface {
	WithTx(ctx context.Context, do func(ctx context.Context) error) error
	ReadOnly(ctx context.Context, do func(ctx context.Context) error) error
}

type FeatureConfigServicePlanStore interface {
	GetPlan(ctx context.Context, name string) (*plan.Plan, error)
}

type FeatureConfigServiceConfigSourceStore interface {
	GetDatabaseSourceByAppID(ctx context.Context, appID string) (*configsource.DatabaseSource, error)
	UpdateDatabaseSource(ctx context.Context, dbs *configsource.DatabaseSource) error
}

type FeatureConfigServiceAuditService interface {
	LogEvent(ctx context.Context, appID string, payload event.NonBlockingPayload) error
}

// ---- Domain types -------------------------------------------------------------

// AppFeatureConfigResult is the internal computation result — pointer-typed
// because config.FeatureConfig sections are pointer fields throughout
// pkg/lib/config. Transport dereferences EffectivePlanFeatureConfig /
// EffectiveAppFeatureConfig into the (value-typed, generated) wire schema.
type AppFeatureConfigResult struct {
	PlanName                   string
	EffectivePlanFeatureConfig *config.FeatureConfig
	AppFeatureConfigYAML       string
	EffectiveAppFeatureConfig  *config.FeatureConfig
}

// ---- FeatureConfigService -------------------------------------------------------

type FeatureConfigService struct {
	GlobalDatabase    FeatureConfigServiceGlobalDatabase
	PlanStore         FeatureConfigServicePlanStore
	ConfigSourceStore FeatureConfigServiceConfigSourceStore
	AuditService      FeatureConfigServiceAuditService
	BaseResources     deps.AppBaseResources
	Clock             clock.Clock
}

func (s *FeatureConfigService) GetAppFeatureConfig(ctx context.Context, appID string) (*AppFeatureConfigResult, error) {
	var planName string
	var overrideYAML []byte
	var planEffective *config.FeatureConfig
	var effective *config.FeatureConfig
	err := s.GlobalDatabase.ReadOnly(ctx, func(ctx context.Context) error {
		dbs, e := s.ConfigSourceStore.GetDatabaseSourceByAppID(ctx, appID)
		if errors.Is(e, configsource.ErrAppNotFound) {
			return apierrors.NotFound.WithReason("AppNotFound").New("app not found")
		}
		if e != nil {
			return e
		}
		planName = dbs.PlanName
		overrideYAML = dbs.Data[featureConfigOverrideKey()]

		planEffective, effective, e = s.computeLayers(ctx, planName, overrideYAML)
		return e
	})
	if err != nil {
		return nil, err
	}

	return &AppFeatureConfigResult{
		PlanName:                   planName,
		EffectivePlanFeatureConfig: planEffective,
		AppFeatureConfigYAML:       string(overrideYAML),
		EffectiveAppFeatureConfig:  effective,
	}, nil
}

func (s *FeatureConfigService) UpdateAppFeatureConfig(ctx context.Context, appID string, rawYAML string) (*AppFeatureConfigResult, error) {
	overrideBytes, err := parseAppFeatureConfigOverride(rawYAML)
	if err != nil {
		return nil, err
	}

	var planName string
	var oldYAML string
	var planEffective *config.FeatureConfig
	var effective *config.FeatureConfig
	err = s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		dbs, e := s.ConfigSourceStore.GetDatabaseSourceByAppID(ctx, appID)
		if errors.Is(e, configsource.ErrAppNotFound) {
			return apierrors.NotFound.WithReason("AppNotFound").New("app not found")
		}
		if e != nil {
			return e
		}
		planName = dbs.PlanName

		key := featureConfigOverrideKey()
		oldYAML = string(dbs.Data[key])

		planEffective, effective, e = s.computeLayers(ctx, dbs.PlanName, overrideBytes)
		if e != nil {
			return e
		}

		if dbs.Data == nil {
			dbs.Data = map[string][]byte{}
		}
		if overrideBytes == nil {
			delete(dbs.Data, key)
		} else {
			dbs.Data[key] = overrideBytes
		}
		dbs.UpdatedAt = s.Clock.NowUTC()

		return s.ConfigSourceStore.UpdateDatabaseSource(ctx, dbs)
	})
	if err != nil {
		return nil, err
	}

	if s.AuditService != nil {
		if e := s.AuditService.LogEvent(ctx, appID, &siteadminauditlog.AppFeatureConfigUpdatedPayload{
			AppID:                   appID,
			OldAppFeatureConfigYAML: oldYAML,
			NewAppFeatureConfigYAML: string(overrideBytes),
		}); e != nil {
			AuditServiceLogger.GetLogger(ctx).WithError(e).Error(ctx, "failed to emit site admin audit log")
		}
	}

	return &AppFeatureConfigResult{
		PlanName:                   planName,
		EffectivePlanFeatureConfig: planEffective,
		AppFeatureConfigYAML:       string(overrideBytes),
		EffectiveAppFeatureConfig:  effective,
	}, nil
}

func (s *FeatureConfigService) PreviewAppFeatureConfig(ctx context.Context, appID string, rawYAML string) (*AppFeatureConfigResult, error) {
	overrideBytes, err := parseAppFeatureConfigOverride(rawYAML)
	if err != nil {
		return nil, err
	}

	var planName string
	var planEffective *config.FeatureConfig
	var effective *config.FeatureConfig
	err = s.GlobalDatabase.ReadOnly(ctx, func(ctx context.Context) error {
		dbs, e := s.ConfigSourceStore.GetDatabaseSourceByAppID(ctx, appID)
		if errors.Is(e, configsource.ErrAppNotFound) {
			return apierrors.NotFound.WithReason("AppNotFound").New("app not found")
		}
		if e != nil {
			return e
		}
		planName = dbs.PlanName

		planEffective, effective, e = s.computeLayers(ctx, dbs.PlanName, overrideBytes)
		return e
	})
	if err != nil {
		return nil, err
	}

	return &AppFeatureConfigResult{
		PlanName:                   planName,
		EffectivePlanFeatureConfig: planEffective,
		AppFeatureConfigYAML:       string(overrideBytes),
		EffectiveAppFeatureConfig:  effective,
	}, nil
}

// ---- Merge computation --------------------------------------------------------

func featureConfigOverrideKey() string {
	return filepathutil.EscapePath(configsource.AuthgearFeatureYAML)
}

// computeLayers folds BaseResources, the plan named planName, and a
// candidate app-override document (nil/empty overrideYAML = no override)
// exactly the way dbApp.doLoad computes an app's effective feature config —
// same layer stack, same order, same FS-construction helpers. No merge
// logic is reimplemented.
func (s *FeatureConfigService) computeLayers(
	ctx context.Context,
	planName string,
	overrideYAML []byte,
) (planEffective *config.FeatureConfig, effective *config.FeatureConfig, err error) {
	p, err := s.PlanStore.GetPlan(ctx, planName)
	if err != nil && !errors.Is(err, plan.ErrPlanNotFound) {
		return nil, nil, err
	}
	// On ErrPlanNotFound, p stays nil. MakePlanFSFromPlan(nil) produces an
	// empty plan layer — the same tolerance dbApp.doLoad has for an app
	// referencing a missing plan.

	planFs, err := configsource.MakePlanFSFromPlan(p)
	if err != nil {
		return nil, nil, err
	}

	appData := map[string][]byte{}
	if len(overrideYAML) > 0 {
		appData[featureConfigOverrideKey()] = overrideYAML
	}
	appFs, err := configsource.MakeAppFSFromDatabaseSource(&configsource.DatabaseSource{Data: appData})
	if err != nil {
		return nil, nil, err
	}

	base := (*resource.Manager)(s.BaseResources)
	planEffective, err = readEffectiveFeatureConfig(ctx, base.Overlay(planFs))
	if err != nil {
		return nil, nil, err
	}
	effective, err = readEffectiveFeatureConfig(ctx, base.Overlay(planFs).Overlay(appFs))
	if err != nil {
		return nil, nil, err
	}
	return planEffective, effective, nil
}

// readEffectiveFeatureConfig mirrors configsource.LoadConfig's feature config
// resolution: read the FeatureConfig resource from the given manager,
// falling back to server defaults when no layer has an authgear.features.yaml
// file.
func readEffectiveFeatureConfig(ctx context.Context, mgr *resource.Manager) (*config.FeatureConfig, error) {
	result, err := mgr.Read(ctx, configsource.FeatureConfig, resource.EffectiveResource{})
	if errors.Is(err, resource.ErrResourceNotFound) {
		return config.NewEffectiveDefaultFeatureConfig(), nil
	}
	if err != nil {
		return nil, err
	}
	return result.(*config.FeatureConfig), nil
}

// ---- Validation (PUT and POST share this) --------------------------------------

// parseAppFeatureConfigOverride handles the two things that must be checked
// before a candidate override can even be considered for merging. Returns
// (nil, nil) when raw is empty or whitespace-only, meaning "clear the
// override" — this is the one case callers must treat specially (not an
// error, and not "store an empty file"). Does NOT run schema validation —
// that happens inside computeLayers, which validates this data as one of the
// layers it merges, plus the merged result. It DOES reject YAML that fails
// to even parse syntactically, for the same reason as the multi-document
// check below: computeLayers's underlying error for a syntax failure
// (configsource.AuthgearFeatureYAMLDescriptor.viewEffectiveResource's
// "malformed feature config: %w") is a plain wrapped error, not a
// *validation.AggregatedError, so pkg/api/apierrors's asAPIError can't
// recognize it as ValidationFailed — it falls through to an unclassified
// 500 UnexpectedError instead. A syntax error is a whole-document problem
// like the multi-document case, not a per-field one, so it gets the same
// hand-built ValidationFailed treatment (no causes).
func parseAppFeatureConfigOverride(raw string) ([]byte, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	data := []byte(raw)

	n, err := countYAMLDocuments(data)
	if err != nil {
		return nil, apierrors.ValidationFailed.New("app_feature_config_yaml is not valid YAML: " + err.Error())
	}
	if n > 1 {
		return nil, apierrors.ValidationFailed.New("app_feature_config_yaml must contain at most one YAML document")
	}

	return data, nil
}

// countYAMLDocuments counts YAML documents separated by "---". A genuine
// syntax error is returned to the caller (see parseAppFeatureConfigOverride
// above) — only a clean end-of-input is treated as "done counting".
func countYAMLDocuments(data []byte) (int, error) {
	dec := goyaml.NewDecoder(bytes.NewReader(data))
	count := 0
	for {
		var doc any
		if err := dec.Decode(&doc); err != nil {
			if errors.Is(err, io.EOF) {
				return count, nil
			}
			return count, err
		}
		count++
	}
}
