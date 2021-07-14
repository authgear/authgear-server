package plan

import (
	"bytes"
	"encoding/json"

	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/portal/lib/plan"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type Service struct {
	Handle            *globaldb.Handle
	Store             *plan.Store
	ConfigSourceStore *configsource.Store
	Clock             clock.Clock
}

func (s *Service) CreatePlan(name string) error {
	return s.Handle.WithTx(func() (err error) {
		p := model.NewPlan(name)
		return s.Store.Create(p)
	})
}

func (s *Service) GetPlan(name string) (p *model.Plan, err error) {
	err = s.Handle.WithTx(func() (e error) {
		p, e = s.Store.GetPlan(name)
		return
	})
	return
}

// UpdatePlan update the plan feature config and also the app which
// have tha same plan name, returns the updated app IDs
func (s Service) UpdatePlan(name string, featureConfigYAML []byte) (appIDs []string, err error) {
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

	err = s.Handle.WithTx(func() (err error) {
		p, err := s.Store.GetPlan(name)
		if err != nil {
			return err
		}

		p.RawFeatureConfig = rawFeatureConfig
		return s.Store.Update(p)
	})
	if err != nil {
		return
	}

	// update apps feature config
	err = s.Handle.WithTx(func() (err error) {
		consrcs, err := s.ConfigSourceStore.ListByPlan(name)
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
			err = s.ConfigSourceStore.UpdateDatabaseSource(consrc)
			if err != nil {
				return err
			}
			appIDs = append(appIDs, consrc.AppID)
		}
		return nil
	})
	return
}

func (s Service) UpdateAppPlan(appID string, planName string) error {
	return s.Handle.WithTx(func() (err error) {
		consrc, err := s.ConfigSourceStore.GetDatabaseSourceByAppID(appID)
		if err != nil {
			return err
		}

		p, err := s.Store.GetPlan(planName)
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
		err = s.ConfigSourceStore.UpdateDatabaseSource(consrc)
		if err != nil {
			return err
		}

		return nil
	})
}

func (s Service) GetDatabaseSourceByAppID(appID string) (consrc *configsource.DatabaseSource, err error) {
	err = s.Handle.WithTx(func() (e error) {
		consrc, e = s.ConfigSourceStore.GetDatabaseSourceByAppID(appID)
		return
	})
	return
}

func (s Service) UpdateAppFeatureConfig(appID string, featureConfigYAML []byte, planName string) (err error) {
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

	return s.Handle.WithTx(func() (err error) {
		consrc, err := s.ConfigSourceStore.GetDatabaseSourceByAppID(appID)
		if err != nil {
			return err
		}

		consrc.PlanName = planName
		// json.Marshal handled base64 encoded of the YAML file
		consrc.Data[configsource.AuthgearFeatureYAML] = rawFeatureConfigYAML
		consrc.UpdatedAt = s.Clock.NowUTC()
		err = s.ConfigSourceStore.UpdateDatabaseSource(consrc)
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
