package plugin

import (
	"encoding/json"

	log "github.com/Sirupsen/logrus"

	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skyerr"
)

type LambdaHandler struct {
	Plugin            *Plugin
	Name              string
	AccessKeyRequired bool
	UserRequired      bool
	PreprocessorList  router.PreprocessorRegistry
	preprocessors     []router.Processor
}

func NewLambdaHandler(info map[string]interface{}, ppreg router.PreprocessorRegistry, p *Plugin) *LambdaHandler {
	handler := &LambdaHandler{
		Plugin:           p,
		Name:             info["name"].(string),
		PreprocessorList: ppreg,
	}
	handler.AccessKeyRequired, _ = info["key_required"].(bool)
	handler.UserRequired, _ = info["user_required"].(bool)
	return handler
}

func (h *LambdaHandler) Setup() {
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

func (h *LambdaHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

// Handle executes lambda function implemented by the plugin.
func (h *LambdaHandler) Handle(payload *router.Payload, response *router.Response) {
	inbytes, err := json.Marshal(payload.Data)
	if err != nil {
		response.Err = skyerr.NewUnknownErr(err)
		return
	}

	outbytes, err := h.Plugin.transport.RunLambda(payload.Context, h.Name, inbytes)
	if err != nil {
		response.Err = skyerr.NewUnknownErr(err)
		return
	}

	result := map[string]interface{}{}
	err = json.Unmarshal(outbytes, &result)
	if err != nil {
		response.Err = skyerr.NewUnknownErr(err)
		return
	}
	log.WithFields(log.Fields{
		"name":   h.Name,
		"input":  payload.Data,
		"result": result,
		"err":    err,
	}).Debugf("Executed a lambda with result")

	response.Result = result
}
