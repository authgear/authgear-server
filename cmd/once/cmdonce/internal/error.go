package internal

import (
	"errors"
)

var (
	ErrNoDocker                     = errors.New("no docker")
	ErrDockerVolumeExists           = errors.New("docker volume exists")
	ErrDockerContainerNotExists     = errors.New("docker container does not exist")
	ErrCommandUpgradeNotImplemented = errors.New("command upgrade not implemented")
)
