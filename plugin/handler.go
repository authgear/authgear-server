package plugin

import (
	log "github.com/Sirupsen/logrus"
	"io/ioutil"

	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skyerr"
)

type PluginHandler struct {
	Plugin            *Plugin
	Name              string
	AccessKeyRequired bool
	UserRequired      bool
	PreprocessorList  router.PreprocessorRegistry
	preprocessors     []router.Processor
}

func NewPluginHandler(info pluginHandlerInfo, ppreg router.PreprocessorRegistry, p *Plugin) *PluginHandler {
	handler := &PluginHandler{
		Plugin:            p,
		Name:              info.Name,
		AccessKeyRequired: info.KeyRequired,
		UserRequired:      info.UserRequired,
		PreprocessorList:  ppreg,
	}
	return handler
}

func (h *PluginHandler) Setup() {
	if h.UserRequired {
		h.preprocessors = h.PreprocessorList.GetByNames(
			"plugin", "authenticator", "dbconn", "inject_user", "require_user")
	} else if h.AccessKeyRequired {
		h.preprocessors = h.PreprocessorList.GetByNames(
			"plugin", "authenticator")
	} else {
		h.preprocessors = h.PreprocessorList.GetByNames("plugin")
	}
}

func (h *PluginHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

// Handle executes lambda function implemented by the plugin.
func (h *PluginHandler) Handle(payload *router.Payload, response *router.Response) {
	inbytes, err := ioutil.ReadAll(payload.Req.Body)
	if err != nil {
		panic(err)
	}

	outbytes, err := h.Plugin.transport.RunHandler(h.Name, inbytes)
	log.Debugf("outbytes %s", outbytes)
	if err != nil {
		switch e := err.(type) {
		case skyerr.Error:
			response.Err = e
		case error:
			response.Err = skyerr.NewUnknownErr(err)
		}
		return
	}

	log.WithFields(log.Fields{
		"name": h.Name,
		"err":  err,
	}).Debugf("Executed a handler with result")
	response.Write(outbytes)
}
