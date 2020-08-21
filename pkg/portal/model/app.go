package model

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type App struct {
	ID           string               `json:"id"`
	AppConfig    *config.AppConfig    `json:"appConfig"`
	SecretConfig *config.SecretConfig `json:"secretConfig"`
}
