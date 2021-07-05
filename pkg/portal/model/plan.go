package model

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type Plan struct {
	ID               string
	Name             string
	RawFeatureConfig *config.FeatureConfig
}
