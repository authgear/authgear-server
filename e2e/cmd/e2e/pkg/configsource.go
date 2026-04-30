package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	cp "github.com/otiai10/copy"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
)

var BuiltInConfigSourceDir = "./var"

func (c *End2End) CreateApp(ctx context.Context, appID string, baseConfigSourceDir string, override string, featuresOverride string, extraFilesDir string) error {
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		return err
	}

	configSourceDir, err := c.createTempConfigSource(ctx, appID, baseConfigSourceDir, override, featuresOverride)
	if err != nil {
		return err
	}

	if extraFilesDir != "" {
		err = cp.Copy(extraFilesDir, configSourceDir)
		if err != nil {
			return err
		}
	}

	err = CreatePortalConfigSource(
		cfg.GlobalDatabase.DatabaseURL,
		cfg.GlobalDatabase.DatabaseSchema,
		configSourceDir,
	)
	if err != nil {
		return err
	}

	err = CreatePortalDefaultDomain(
		cfg.GlobalDatabase.DatabaseURL,
		cfg.GlobalDatabase.DatabaseSchema,
		".authgeare2e.localhost",
		appID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *End2End) createTempConfigSource(ctx context.Context, appID string, baseConfigSource string, overrideYAML string, featuresOverrideYAML string) (string, error) {
	tempAppDir, err := os.MkdirTemp("", "e2e-")
	if err != nil {
		return "", err
	}

	err = cp.Copy(BuiltInConfigSourceDir, tempAppDir)
	if err != nil {
		return "", err
	}

	baseConfigSourceInfo, err := os.Stat(baseConfigSource)
	if err != nil {
		return "", err
	}

	dest := tempAppDir
	if !baseConfigSourceInfo.IsDir() {
		dest = filepath.Join(tempAppDir, baseConfigSourceInfo.Name())
	}
	err = cp.Copy(baseConfigSource, dest)
	if err != nil {
		return "", err
	}

	authgearYAMLPath := filepath.Join(tempAppDir, configsource.AuthgearYAML)
	authgearYAML, err := os.ReadFile(authgearYAMLPath)
	if err != nil {
		return "", err
	}

	cfg, err := config.Parse(ctx, []byte(authgearYAML))
	if err != nil {
		return "", err
	}

	mergedYAML, err := mergeYAMLObjects(authgearYAML, []byte(overrideYAML))
	if err != nil {
		return "", err
	}
	cfg, err = config.Parse(ctx, mergedYAML)
	if err != nil {
		return "", err
	}

	cfg.ID = config.AppID(appID)
	cfg.HTTP.PublicOrigin = fmt.Sprintf("http://%s.authgeare2e.localhost:4000", appID)

	newAuthgearYAML, err := exportConfig(cfg)
	if err != nil {
		return "", err
	}

	err = os.WriteFile(authgearYAMLPath, newAuthgearYAML, fs.FileMode(0644))
	if err != nil {
		return "", err
	}

	if featuresOverrideYAML != "" {
		authgearFeaturesYAMLPath := filepath.Join(tempAppDir, configsource.AuthgearFeatureYAML)
		authgearFeaturesYAML, err := os.ReadFile(authgearFeaturesYAMLPath)
		if err != nil {
			return "", err
		}

		newFeaturesYAML, err := mergeYAMLObjects(authgearFeaturesYAML, []byte(featuresOverrideYAML))
		if err != nil {
			return "", err
		}
		_, err = config.ParseFeatureConfigWithoutDefaults(ctx, newFeaturesYAML)
		if err != nil {
			return "", err
		}

		err = os.WriteFile(authgearFeaturesYAMLPath, newFeaturesYAML, fs.FileMode(0644))
		if err != nil {
			return "", err
		}
	}

	return tempAppDir, nil
}

func mergeYAMLObjects(baseYAML []byte, overrideYAML []byte) ([]byte, error) {
	if len(bytes.TrimSpace(overrideYAML)) == 0 {
		return baseYAML, nil
	}

	baseJSON, err := yaml.YAMLToJSON(baseYAML)
	if err != nil {
		return nil, err
	}
	overrideJSON, err := yaml.YAMLToJSON(overrideYAML)
	if err != nil {
		return nil, err
	}

	var baseObj map[string]interface{}
	if err := json.Unmarshal(baseJSON, &baseObj); err != nil {
		return nil, err
	}
	var overrideObj map[string]interface{}
	if err := json.Unmarshal(overrideJSON, &overrideObj); err != nil {
		return nil, err
	}

	mergeJSONObject(baseObj, overrideObj)

	mergedJSON, err := json.Marshal(baseObj)
	if err != nil {
		return nil, err
	}
	return yaml.JSONToYAML(mergedJSON)
}

func mergeJSONObject(dst map[string]interface{}, src map[string]interface{}) {
	for k, srcValue := range src {
		srcMap, srcIsMap := srcValue.(map[string]interface{})
		dstMap, dstIsMap := dst[k].(map[string]interface{})
		if srcIsMap && dstIsMap {
			mergeJSONObject(dstMap, srcMap)
			continue
		}
		dst[k] = srcValue
	}
}

func exportFeaturesConfig(cfg *config.FeatureConfig) ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	err := encoder.Encode(cfg)
	if err != nil {
		return nil, err
	}

	jsonData := buf.Bytes()
	yamlData, err := yaml.JSONToYAML(jsonData)
	if err != nil {
		return nil, err
	}

	return yamlData, nil
}

func exportConfig(config *config.AppConfig) ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	err := encoder.Encode(config)
	if err != nil {
		return nil, err
	}

	jsonData := buf.Bytes()
	yamlData, err := yaml.JSONToYAML(jsonData)
	if err != nil {
		return nil, err
	}

	return yamlData, nil
}
