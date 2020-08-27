package model

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type App struct {
	ID      string
	Context *config.AppContext
}
