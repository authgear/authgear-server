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
}

// ReadFileInto reads a configuration from file specified by path
func ReadFileInto(config *Configuration, path string) {
	if err := gcfg.ReadFileInto(config, path); err != nil {
		// it is certainly an fatal error if we cannot read the config file
		// into Configuration, so panic here
		panic(err)
	}
}
