package preprocessor

import (
	"net/http"

	log "github.com/Sirupsen/logrus"

	"github.com/oursky/skygear/plugin"
	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skyerr"
)

type HookRegistryPreprocessor struct {
	PluginInitContext *plugin.InitContext
}

func (p HookRegistryPreprocessor) Preprocess(payload *router.Payload, response *router.Response) int {
	if !p.PluginInitContext.IsReady() {
		log.Errorf("Request cannot be handled because plugins are unavailable at the moment.")
		response.Err = skyerr.NewError(skyerr.PluginUnavailable, "plugins are unavailable at the moment")
		return http.StatusServiceUnavailable
	}
	payload.HookRegistry = p.PluginInitContext.HookRegistry
	return http.StatusOK
}
