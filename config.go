package main

import (
	"code.google.com/p/gcfg"
)

// Configuration is Ourd's configuration
type Configuration struct {
	HTTP struct {
		Host string
	}
	DB struct {
		ImplName string `gcfg:"implementation"`
		AppName  string `gcfg:"app-name"`
		Option   string
	}
	TokenStore struct {
		Path string `gcfg:"path"`
	} `gcfg:"token-store"`
	Subscription struct {
		Enabled bool
	}
	APNS struct {
		Gateway  string
		CertPath string `gcfg:"cert-path"`
		KeyPath  string `gcfg:"key-path"`
	}
}

// ReadFileInto reads a configuration from file specified by path
func ReadFileInto(config *Configuration, path string) error {
	return gcfg.ReadFileInto(config, path)
}
