package provider

import ()

// AuthProvider is an interface to provide user authentication.
type AuthProvider interface {
	Login(authData map[string]interface{}) (string, map[string]interface{}, error)
	Logout(authData map[string]interface{}) (map[string]interface{}, error)
	Info(authData map[string]interface{}) (map[string]interface{}, error)
}
