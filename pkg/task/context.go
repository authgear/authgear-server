package task

import (
	"github.com/skygeario/skygear-server/pkg/auth/config"
)

type Context struct {
	Config                *config.Config
	PreferredLanguageTags []string
}
