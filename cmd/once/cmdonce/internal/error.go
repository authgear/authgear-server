package internal

import (
	"errors"
)

var (
	ErrNoDocker           = errors.New("no docker")
	ErrDockerVolumeExists = errors.New("docker volume exists")
)
