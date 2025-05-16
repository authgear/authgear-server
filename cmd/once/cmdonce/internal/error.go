package internal

import (
	"errors"
	"fmt"
)

var (
	ErrNoDocker                                = errors.New("no docker")
	ErrDockerVolumeExists                      = errors.New("docker volume exists")
	ErrDockerContainerNotExists                = errors.New("docker container does not exist")
	ErrCommandUpgradeNotImplemented            = errors.New("command upgrade not implemented")
	ErrLicenseServerUnknownResponse            = errors.New("unknown response from license server")
	ErrLicenseServerLicenseKeyNotFound         = errors.New("license server license key not found")
	ErrLicenseServerLicenseKeyAlreadyActivated = errors.New("license server license key already activated")
)

type ErrTCPPortAlreadyListening struct {
	Port int
}

func (e *ErrTCPPortAlreadyListening) Error() string {
	return fmt.Sprintf("TCP port %v already listening", e.Port)
}
