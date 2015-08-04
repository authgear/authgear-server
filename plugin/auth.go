package plugin

import ()

// AuthProvider is implemented by plugin to provider user authentication functionality to Ourd.
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
