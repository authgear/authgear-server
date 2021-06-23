package model

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type Plan struct {
	ID            string
	Name          string
	FeatureConfig *config.FeatureConfig
}
