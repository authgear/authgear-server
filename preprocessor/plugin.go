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

	"github.com/skygeario/skygear-server/plugin"
	"github.com/skygeario/skygear-server/router"
	"github.com/skygeario/skygear-server/skyerr"
)

type EnsurePluginReadyPreprocessor struct {
	PluginInitContext *plugin.InitContext
}

func (p *EnsurePluginReadyPreprocessor) Preprocess(payload *router.Payload, response *router.Response) int {
	if !p.PluginInitContext.IsReady() {
		log.Errorf("Request cannot be handled because plugins are unavailable at the moment.")
		response.Err = skyerr.NewError(skyerr.PluginUnavailable, "plugins are unavailable at the moment")
		return http.StatusServiceUnavailable
	}
	return http.StatusOK
}
