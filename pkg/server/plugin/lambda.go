// Copyright 2015-present Oursky Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package plugin

import (
	"encoding/json"

	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
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
			"authenticator",
			"dbconn",
			"inject_auth",
			"require_auth",
			"plugin_ready",
		)
	} else if h.AccessKeyRequired {
		h.preprocessors = h.PreprocessorList.GetByNames(
			"authenticator",
			"plugin_ready",
		)
	} else {
		h.preprocessors = h.PreprocessorList.GetByNames(
			"inject_auth_id",
			"plugin_ready",
		)
	}
}

func (h *LambdaHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

// Handle executes lambda function implemented by the plugin.
func (h *LambdaHandler) Handle(payload *router.Payload, response *router.Response) {
	inbytes, err := json.Marshal(payload.Data)
	if err != nil {
		response.Err = skyerr.MakeError(err)
		return
	}

	outbytes, err := h.Plugin.transport.RunLambda(payload.Context, h.Name, inbytes)
	if err != nil {
		switch e := err.(type) {
		case skyerr.Error:
			response.Err = e
		case error:
			response.Err = skyerr.MakeError(err)
		}
		return
	}

	result := map[string]interface{}{}
	err = json.Unmarshal(outbytes, &result)
	if err != nil {
		response.Err = skyerr.MakeError(err)
		return
	}
	log.WithFields(logrus.Fields{
		"name":   h.Name,
		"input":  payload.Data,
		"result": result,
		"err":    err,
	}).Debugf("Executed a lambda with result")

	response.Result = result
}
