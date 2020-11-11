package deps

import (
	"github.com/authgear/authgear-server/pkg/util/resource"
)

type AppBaseResources *resource.Manager

func ProvideAppBaseResources(root *RootProvider) AppBaseResources {
	return root.AppBaseResources
}
