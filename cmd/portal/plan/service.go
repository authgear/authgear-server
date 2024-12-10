package plan

import (
	"bytes"
	"context"
	"encoding/json"

	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/config/plan"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type Service struct {
	Handle            *globaldb.Handle
	Store             *plan.Store
	ConfigSourceStore *configsource.Store
	Clock             clock.Clock
}

func (s *Service) CreatePlan(ctx context.Context, name string) error {
	return s.Handle.WithTx(ctx, func(ctx context.Context) (err error) {
		p := plan.NewPlan(name)
		return s.Store.Create(ctx, p)
	})
}

func (s *Service) GetPlan(ctx context.Context, name string) (p *plan.Plan, err error) {
	err = s.Handle.WithTx(ctx, func(ctx context.Context) (e error) {
		p, e = s.Store.GetPlan(ctx, name)
		return
	})
	return
}

// UpdatePlan update the plan feature config and also the app which
// have tha same plan name, returns the updated app IDs
func (s Service) UpdatePlan(ctx context.Context, name string, featureConfigYAML []byte) (appIDs []string, err error) {
	// validation
	_, err = config.ParseFeatureConfig(featureConfigYAML)
	if err != nil {
		return
	}

	rawFeatureConfig, err := parseRawFeatureConfig(featureConfigYAML)
	if err != nil {
		return
	}

	rawFeatureConfigYAML, e := yaml.Marshal(rawFeatureConfig)
	if e != nil {
		err = e
		return
	}

	err = s.Handle.WithTx(ctx, func(ctx context.Context) (err error) {
		p, err := s.Store.GetPlan(ctx, name)
		if err != nil {
			return err
		}

		p.RawFeatureConfig = rawFeatureConfig
		return s.Store.Update(ctx, p)
	})
	if err != nil {
		return
	}

	// update apps feature config
	err = s.Handle.WithTx(ctx, func(ctx context.Context) (err error) {
		consrcs, err := s.ConfigSourceStore.ListByPlan(ctx, name)
		if err != nil {
			return err
		}
		for _, consrc := range consrcs {
			// json.Marshal handled base64 encoded of the YAML file
			// https://golang.org/pkg/encoding/json/#Marshal
			// Array and slice values encode as JSON arrays,
			// except that []byte encodes as a base64-encoded string,
			// and a nil slice encodes as the null JSON value.
			consrc.Data[configsource.AuthgearFeatureYAML] = rawFeatureConfigYAML
			consrc.UpdatedAt = s.Clock.NowUTC()
			err = s.ConfigSourceStore.UpdateDatabaseSource(ctx, consrc)
			if err != nil {
				return err
			}
			appIDs = append(appIDs, consrc.AppID)
		}
		return nil
	})
	return
}

func (s Service) UpdateAppPlan(ctx context.Context, appID string, planName string) error {
	return s.Handle.WithTx(ctx, func(ctx context.Context) (err error) {
		consrc, err := s.ConfigSourceStore.GetDatabaseSourceByAppID(ctx, appID)
		if err != nil {
			return err
		}

		p, err := s.Store.GetPlan(ctx, planName)
		if err != nil {
			return err
		}

		featureConfigYAML, e := yaml.Marshal(p.RawFeatureConfig)
		if e != nil {
			err = e
			return
		}

		consrc.PlanName = p.Name
		// json.Marshal handled base64 encoded of the YAML file
		consrc.Data[configsource.AuthgearFeatureYAML] = featureConfigYAML
		consrc.UpdatedAt = s.Clock.NowUTC()
		err = s.ConfigSourceStore.UpdateDatabaseSource(ctx, consrc)
		if err != nil {
			return err
		}

		return nil
	})
}

func (s Service) GetDatabaseSourceByAppID(ctx context.Context, appID string) (consrc *configsource.DatabaseSource, err error) {
	err = s.Handle.WithTx(ctx, func(ctx context.Context) (e error) {
		consrc, e = s.ConfigSourceStore.GetDatabaseSourceByAppID(ctx, appID)
		return
	})
	return
}

func (s Service) UpdateAppFeatureConfig(ctx context.Context, appID string, featureConfigYAML []byte, planName string) (err error) {
	// validation
	_, err = config.ParseFeatureConfig(featureConfigYAML)
	if err != nil {
		return
	}

	rawFeatureConfig, err := parseRawFeatureConfig(featureConfigYAML)
	if err != nil {
		return
	}

	rawFeatureConfigYAML, e := yaml.Marshal(rawFeatureConfig)
	if e != nil {
		err = e
		return
	}

	return s.Handle.WithTx(ctx, func(ctx context.Context) (err error) {
		consrc, err := s.ConfigSourceStore.GetDatabaseSourceByAppID(ctx, appID)
		if err != nil {
			return err
		}

		consrc.PlanName = planName
		// json.Marshal handled base64 encoded of the YAML file
		consrc.Data[configsource.AuthgearFeatureYAML] = rawFeatureConfigYAML
		consrc.UpdatedAt = s.Clock.NowUTC()
		err = s.ConfigSourceStore.UpdateDatabaseSource(ctx, consrc)
		if err != nil {
			return err
		}

		return nil
	})
}

func parseRawFeatureConfig(inputYAML []byte) (*config.FeatureConfig, error) {
	jsonData, err := yaml.YAMLToJSON(inputYAML)
	if err != nil {
		return nil, err
	}
	var config config.FeatureConfig
	err = json.NewDecoder(bytes.NewReader(jsonData)).Decode(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
