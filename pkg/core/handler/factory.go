package handler

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
)

type Factory interface {
	NewHandler(config.TenantConfiguration) Handler
}
