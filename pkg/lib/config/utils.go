package config

import (
	"github.com/sirupsen/logrus"

	"github.com/authgear/authgear-server/pkg/util/log"
)

func newBool(v bool) *bool { return &v }

func newInt(v int) *int { return &v }

func NewSecretMaskLogHook(cfg *SecretConfig) logrus.Hook {
	var patterns []log.MaskPattern
	for _, item := range cfg.Secrets {
		for _, s := range item.Data.SensitiveStrings() {
			if len(s) == 0 {
				continue
			}
			patterns = append(patterns, log.NewPlainMaskPattern(s))
		}
	}

	return &log.FormatHook{
		MaskPatterns: patterns,
		Mask:         "********",
	}
}
