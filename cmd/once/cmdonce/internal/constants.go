package internal

const (
	Version = "1.0.0"
)

const (
	// ProgramName is the name of this program.
	ProgramName = "authgear-once"
	// BinDocker is the name of the binary of docker.
	BinDocker = "docker"

	// NOTE(once): Possibly breaking change
	// The name of the volume cannot be changed without any consideration.
	// NameDockerVolume is the name of the Docker volume.
	NameDockerVolume = "authgearonce_data"
	// NOTE(once): Possibly breaking change
	// The name of the container cannot be changed without any consideration.
	// NameDockerContainer is the name of the Docker container.
	NameDockerContainer = "authgearonce"

	DefaultDockerName_NoTag = "quay.io/theauthgear/authgear-once"
)
