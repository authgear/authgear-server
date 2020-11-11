package loginid

import (
	"github.com/authgear/authgear-server/pkg/util/blocklist"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

type ResourceManager interface {
	Read(desc resource.Descriptor, view resource.View) (interface{}, error)
}

var ReservedNameTXT = resource.RegisterResource(resource.NewlineJoinedDescriptor{
	Path: "reserved_name.txt",
	Parse: func(data []byte) (interface{}, error) {
		return blocklist.New(string(data))
	},
})
