package model

import (
	"time"
)

type DeploymentHooks struct {
	ID                string
	CreatedAt         *time.Time
	DeploymentVersion string
	AppID             string
	Hooks             []DeploymentHook
	IsLastDeployment  bool
}

type DeploymentHook struct {
	Event string
	URL   string
}
