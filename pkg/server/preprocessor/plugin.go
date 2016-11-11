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

package preprocessor

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/server/plugin"
	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

type EnsurePluginReadyPreprocessor struct {
	PluginContext *plugin.Context
	ClientKey     string
	MasterKey     string
}

func (p *EnsurePluginReadyPreprocessor) Preprocess(
	payload *router.Payload,
	response *router.Response,
) int {

	// allow any request when plugins are ready
	if p.PluginContext.IsReady() {
		return http.StatusOK
	}

	// only allow requests with master key and the "_from_plugin" is set to true
	// when the some plugin are just initialized
	if p.PluginContext.IsInitialized() {
		if err := checkRequestAccessKey(payload, p.ClientKey, p.MasterKey); err != nil {
			response.Err = err
			return http.StatusUnauthorized
		}

		fromPlugin, _ := payload.Data["_from_plugin"].(bool)
		if payload.HasMasterKey() && fromPlugin {
			return http.StatusOK
		}

		response.Err = skyerr.NewError(
			skyerr.PluginInitializing,
			"Plugins are initializing at the moment",
		)
		return http.StatusServiceUnavailable
	}

	response.Err = skyerr.NewError(
		skyerr.PluginUnavailable,
		"plugins are unavailable at the moment",
	)
	return http.StatusServiceUnavailable
}
