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

package handlertest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/oursky/skygear/router"
)

// SingleRouteRouter is a router that only serves with a single handler,
// regardless of the requested action.
type SingleRouteRouter router.Router

// NewSingleRouteRouter creates a SingleRouteRouter, mapping the specified
// handler as the only route.
func NewSingleRouteRouter(handler router.Handler, prepareFunc func(*router.Payload)) *SingleRouteRouter {
	r := router.NewRouter()
	r.Map("", handler, &FuncProcessor{prepareFunc})
	return (*SingleRouteRouter)(r)
}

type FuncProcessor struct {
	Mockfunc func(*router.Payload)
}

func (p FuncProcessor) Preprocess(payload *router.Payload, _ *router.Response) int {
	p.Mockfunc(payload)
	return http.StatusOK
}

// POST invoke the only route mapped on the SingleRouteRouter.
func (r *SingleRouteRouter) POST(body string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("POST", "", strings.NewReader(body))
	resp := httptest.NewRecorder()

	(*router.Router)(r).ServeHTTP(resp, req)
	return resp
}

// SingleUserAuthProvider is an AuthProvider that only authenticates
// a single user if the auth data provided contains the required
// principal name.
type SingleUserAuthProvider struct {
	providerName  string
	principalName string
}

// NewSingleUserAuthProvider creates a new instance of SingleUserAuthProvider.
func NewSingleUserAuthProvider(providerName string, principalName string) *SingleUserAuthProvider {
	return &SingleUserAuthProvider{providerName, principalName}
}

// Login implements the AuthProvider's Login interface.
func (p *SingleUserAuthProvider) Login(authData map[string]interface{}) (principalID string, newAuthData map[string]interface{}, err error) {
	if authData["name"] == p.principalName {
		principalID = p.providerName + ":" + p.principalName
		newAuthData = authData
	} else {
		err = fmt.Errorf("Incorrect user.")
	}
	return
}

// Logout implements the AuthProvider's Logout interface.
func (p *SingleUserAuthProvider) Logout(authData map[string]interface{}) (newAuthData map[string]interface{}, err error) {
	newAuthData = authData
	return
}

// Info implements the AuthProvider's Info interface.
func (p *SingleUserAuthProvider) Info(authData map[string]interface{}) (newAuthData map[string]interface{}, err error) {
	newAuthData = authData
	return
}
