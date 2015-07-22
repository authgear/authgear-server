package plugin

import (
	"encoding/json"
)

// AuthProvider is implemented by plugin to provider user authentication functionality to Ourd.
type AuthProvider struct {
	Name   string
	plugin *Plugin
}

// AuthDataRequest is sent by Ourd to plugin which contains data for authentication
type AuthDataRequest struct {
	AuthData map[string]interface{} `json:"auth_data"`
}

// AuthDataResponse is sent by plugin to Ourd which contains authenticated data
type AuthDataResponse struct {
	PrincipalID string                 `json:"principal_id"`
	AuthData    map[string]interface{} `json:"auth_data"`
}

// Login calls the AuthProvider implemented by plugin to request for user login authentication
func (p *AuthProvider) Login(authData map[string]interface{}) (principalID string, newAuthData map[string]interface{}, err error) {
	request := AuthDataRequest{authData}
	inbytes, err := json.Marshal(request)
	if err != nil {
		return
	}

	outbytes, err := p.plugin.transport.RunProvider(p.Name, "login", inbytes)
	if err != nil {
		return
	}

	response := AuthDataResponse{}
	err = json.Unmarshal(outbytes, &response)
	if err != nil {
		return
	}

	principalID = p.Name + ":" + response.PrincipalID
	newAuthData = response.AuthData
	return
}

// Logout calls the AuthProvider implemented by plugin to request for user logout.
func (p *AuthProvider) Logout(authData map[string]interface{}) (newAuthData map[string]interface{}, err error) {
	request := AuthDataRequest{authData}
	inbytes, err := json.Marshal(request)
	if err != nil {
		return
	}

	outbytes, err := p.plugin.transport.RunProvider(p.Name, "logout", inbytes)
	if err != nil {
		return
	}

	response := AuthDataResponse{}
	err = json.Unmarshal(outbytes, &response)
	if err != nil {
		return
	}

	newAuthData = response.AuthData
	return
}

// Info calls the AuthProvider implemented by plugin to request for user information.
func (p *AuthProvider) Info(authData map[string]interface{}) (newAuthData map[string]interface{}, err error) {
	request := AuthDataRequest{authData}
	inbytes, err := json.Marshal(request)
	if err != nil {
		return
	}

	outbytes, err := p.plugin.transport.RunProvider(p.Name, "info", inbytes)
	if err != nil {
		return
	}

	response := AuthDataResponse{}
	err = json.Unmarshal(outbytes, &response)
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
