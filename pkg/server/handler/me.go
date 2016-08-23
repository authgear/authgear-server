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

package handler

import (
	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

// MeHandler handles the me request
type MeHandler struct {
	Authenticator router.Processor `preprocessor:"authenticator"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	InjectUser    router.Processor `preprocessor:"inject_user"`
	preprocessors []router.Processor
}

// Setup adds injected pre-processors to preprocessors array
func (h *MeHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.Authenticator,
		h.DBConn,
		h.InjectUser,
	}
}

// GetPreprocessors returns all pre-processors for the handler
func (h *MeHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

// Handle is the handling method of the me request
func (h *MeHandler) Handle(payload *router.Payload, response *router.Response) {
	userinfo := payload.UserInfo
	if userinfo == nil {
		response.Err = skyerr.NewError(skyerr.NotAuthenticated, "Authentication is needed to get current user")
		return
	}

	response.Result = struct {
		UserID      string   `json:"user_id,omitempty"`
		Username    string   `json:"username,omitempty"`
		Email       string   `json:"email,omitempty"`
		Roles       []string `json:"roles,omitempty"`
		AccessToken string   `json:"access_token,omitempty"`
	}{
		userinfo.ID,
		userinfo.Username,
		userinfo.Email,
		userinfo.Roles,
		payload.AccessTokenString(),
	}
}
