package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"dario.cat/mergo"
	cp "github.com/otiai10/copy"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
)

var BuiltInConfigSourceDir = "./var"

func (c *End2End) CreateApp(appID string, baseConfigSourceDir string, override string) error {
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		return err
	}

	configSourceDir, err := c.createTempConfigSource(appID, baseConfigSourceDir, override)
	if err != nil {
		return err
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
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *End2End) createTempConfigSource(appID string, baseConfigSource string, overrideYAML string) (string, error) {
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

	cfg, err := config.Parse([]byte(authgearYAML))
	if err != nil {
		return "", err
	}

	var overrideCfg config.AppConfig
	jsonData, err := yaml.YAMLToJSON([]byte(overrideYAML))
	if err != nil {
		return "", err
	}

	decoder := json.NewDecoder(bytes.NewReader(jsonData))
	err = decoder.Decode(&overrideCfg)
	if err != nil {
		return "", err
	}

	err = mergo.Merge(cfg, &overrideCfg, mergo.WithOverride)
	if err != nil {
		return "", err
	}

	cfg.ID = config.AppID(appID)
	cfg.HTTP.PublicOrigin = fmt.Sprintf("http://%s.authgeare2e.localhost:4000", appID)
	// This is a workaround for this bug
	// https://github.com/golang/go/issues/38988
	//
	// http.Client always pass request.URL to http.CookieJar.
	// But a more correct behavior should be passing a net.URL
	// with net.URL.Host = http.Request.Host (if http.Request.Host is non-zero)
	//
	// To work around this problem, we ask Authgear to write the cookie
	// with a buggy domain, that is request.URL.Host = "127.0.0.1"
	cookieDomain := "127.0.0.1"
	cfg.HTTP.CookieDomain = &cookieDomain

	newAuthgearYAML, err := exportConfig(cfg)
	if err != nil {
		return "", err
	}

	err = os.WriteFile(authgearYAMLPath, newAuthgearYAML, fs.FileMode(0644))
	if err != nil {
		return "", err
	}

	return tempAppDir, nil
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
