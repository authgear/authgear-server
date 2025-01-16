package runtimeresource

import (
	"embed"
)

//go:embed all:resources/authgear
var EmbedFS_resources_authgear embed.FS

const RelativePath_resources_authgear = "resources/authgear"

//go:embed all:resources/portal
var EmbedFS_resources_portal embed.FS

const RelativePath_resources_portal = "resources/portal"
