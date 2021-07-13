package model

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/portal/appresource"
)

type App struct {
	ID      string
	Context *config.AppContext
}

type AppResource struct {
	DescriptedPath appresource.DescriptedPath
	Context        *config.AppContext
}
