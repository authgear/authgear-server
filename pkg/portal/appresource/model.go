package appresource

import (
	"github.com/authgear/authgear-server/pkg/util/resource"
)

type Update struct {
	Path string
	Data []byte
}

type DescriptedPath struct {
	Descriptor resource.Descriptor
	Path       string
}
