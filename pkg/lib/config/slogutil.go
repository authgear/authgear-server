package config

import (
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

func NewMaskPatternFromSecretConfig(cfg *SecretConfig) []slogutil.MaskPattern {
	var patterns []slogutil.MaskPattern
	if cfg != nil {
		for _, item := range cfg.Secrets {
			for _, s := range item.Data.SensitiveStrings() {
				if s != "" {
					patterns = append(patterns, slogutil.NewPlainMaskPattern(s))
				}
			}
		}
	}
	return patterns
}
