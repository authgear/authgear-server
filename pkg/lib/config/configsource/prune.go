package configsource

import (
	"context"
	"fmt"

	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/filepathutil"
)

// secretKeysKeptOnPrune are the only secrets kept when an app is pruned.
// oauth (the OAuth signing/encryption JWK set), admin-api.auth, and csrf
// are kept so the app id can never be re-registered while every other
// credential is dropped.
var secretKeysKeptOnPrune = map[config.SecretKey]bool{
	config.OAuthKeyMaterialsKey: true,
	config.AdminAPIAuthKeyKey:   true,
	config.CSRFKeyMaterialsKey:  true,
}

// TrimDataForPrune reduces a pruned app's config source Data down to the
// minimum needed to permanently retire its app id: authgear.yaml keeps only
// id and http.public_origin (the smallest config config.Parse accepts), and
// authgear.secrets.yaml keeps only the secrets named in secretKeysKeptOnPrune.
// Every other resource (templates, translations, feature overrides, etc.)
// is dropped.
func TrimDataForPrune(ctx context.Context, data map[string][]byte) (map[string][]byte, error) {
	trimmed := make(map[string][]byte)

	if raw, ok := data[filepathutil.EscapePath(AuthgearYAML)]; ok {
		out, err := trimAppConfigYAMLForPrune(ctx, raw)
		if err != nil {
			return nil, fmt.Errorf("prune: trim %v: %w", AuthgearYAML, err)
		}
		trimmed[filepathutil.EscapePath(AuthgearYAML)] = out
	}

	if raw, ok := data[filepathutil.EscapePath(AuthgearSecretYAML)]; ok {
		out, err := trimSecretConfigYAMLForPrune(ctx, raw)
		if err != nil {
			return nil, fmt.Errorf("prune: trim %v: %w", AuthgearSecretYAML, err)
		}
		trimmed[filepathutil.EscapePath(AuthgearSecretYAML)] = out
	}

	return trimmed, nil
}

func trimAppConfigYAMLForPrune(ctx context.Context, raw []byte) ([]byte, error) {
	var cfg config.AppConfig
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, err
	}
	if cfg.HTTP == nil {
		return nil, fmt.Errorf("http is required")
	}

	minimal := &config.AppConfig{
		ID: cfg.ID,
		HTTP: &config.HTTPConfig{
			PublicOrigin: cfg.HTTP.PublicOrigin,
		},
	}

	out, err := yaml.Marshal(minimal)
	if err != nil {
		return nil, err
	}

	// Sanity check: the trimmed config must still be a config.Parse-able
	// AppConfig, so it is safe to ever be loaded again.
	if _, err := config.Parse(ctx, out); err != nil {
		return nil, fmt.Errorf("trimmed config is invalid: %w", err)
	}

	return out, nil
}

func trimSecretConfigYAMLForPrune(ctx context.Context, raw []byte) ([]byte, error) {
	var cfg config.SecretConfig
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, err
	}

	minimal := &config.SecretConfig{}
	for _, item := range cfg.Secrets {
		if secretKeysKeptOnPrune[item.Key] {
			minimal.Secrets = append(minimal.Secrets, item)
		}
	}

	out, err := yaml.Marshal(minimal)
	if err != nil {
		return nil, err
	}

	// Sanity check: the trimmed secrets must still be a config.ParseSecret-able
	// SecretConfig, so it is safe to ever be loaded again.
	if _, err := config.ParseSecret(ctx, out); err != nil {
		return nil, fmt.Errorf("trimmed secrets are invalid: %w", err)
	}

	return out, nil
}
