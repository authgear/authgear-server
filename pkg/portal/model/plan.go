package model

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type Plan struct {
	ID               string
	Name             string
	RawFeatureConfig *config.FeatureConfig
}

func NewPlan(name string) *Plan {
	return &Plan{
		ID:               uuid.New(),
		Name:             name,
		RawFeatureConfig: &config.FeatureConfig{},
	}
}
