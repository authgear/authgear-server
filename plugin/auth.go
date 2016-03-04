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

import ()

// AuthProvider is implemented by plugin to provider user authentication functionality to Skygear.
type AuthProvider struct {
	Name   string
	plugin *Plugin
}

// Login calls the AuthProvider implemented by plugin to request for user login authentication
func (p *AuthProvider) Login(authData map[string]interface{}) (principalID string, newAuthData map[string]interface{}, err error) {
	request := AuthRequest{p.Name, "login", authData}

	response, err := p.plugin.transport.RunProvider(&request)
	if err != nil {
		return
	}

	principalID = p.Name + ":" + response.PrincipalID
	newAuthData = response.AuthData
	return
}

// Logout calls the AuthProvider implemented by plugin to request for user logout.
func (p *AuthProvider) Logout(authData map[string]interface{}) (newAuthData map[string]interface{}, err error) {
	request := AuthRequest{p.Name, "logout", authData}

	response, err := p.plugin.transport.RunProvider(&request)
	if err != nil {
		return
	}

	newAuthData = response.AuthData
	return
}

// Info calls the AuthProvider implemented by plugin to request for user information.
func (p *AuthProvider) Info(authData map[string]interface{}) (newAuthData map[string]interface{}, err error) {
	request := AuthRequest{p.Name, "info", authData}

	response, err := p.plugin.transport.RunProvider(&request)
	if err != nil {
		return
	}

	newAuthData = response.AuthData
	return
}

// NewAuthProvider creates a new AuthProvider.
func NewAuthProvider(providerName string, plugin *Plugin) *AuthProvider {
	return &AuthProvider{
		Name:   providerName,
		plugin: plugin,
	}
}
