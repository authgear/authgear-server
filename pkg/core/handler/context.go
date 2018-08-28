package handler

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

type Context struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request

	TenantConfiguration config.TenantConfiguration
}
