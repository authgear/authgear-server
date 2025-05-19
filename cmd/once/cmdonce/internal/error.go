package internal

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrNoDocker                                = errors.New("no docker")
	ErrDockerContainerNotExists                = errors.New("docker container does not exist")
	ErrCommandUpgradeNotImplemented            = errors.New("command upgrade not implemented")
	ErrLicenseServerUnknownResponse            = errors.New("unknown response from license server")
	ErrLicenseServerLicenseKeyNotFound         = errors.New("license server license key not found")
	ErrLicenseServerLicenseKeyAlreadyActivated = errors.New("license server license key already activated")
	ErrCertbotExitCode10                       = errors.New("certbot exited with 10")
)

type ErrCertbotFailedToGetCertificates struct {
	LicenseKey string
	Domains    []string
}

func (e *ErrCertbotFailedToGetCertificates) Error() string {
	return fmt.Sprintf("failed to get TLS certificates from Let's Encrypt: %v", strings.Join(e.Domains, ","))
}

type ErrTCPPortAlreadyListening struct {
	Port int
}

func (e *ErrTCPPortAlreadyListening) Error() string {
	return fmt.Sprintf("TCP port %v already listening", e.Port)
}
