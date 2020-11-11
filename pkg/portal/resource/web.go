package resource

import (
	"github.com/authgear/authgear-server/pkg/util/resource"
)

var ThemesJSON = PortalRegistry.Register(&resource.SimpleDescriptor{
	Path: "themes.json",
})

var TranslationsJSON = PortalRegistry.Register(&resource.SimpleDescriptor{
	Path: "translations.json",
})
