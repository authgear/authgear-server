package resource

import (
	"github.com/authgear/authgear-server/pkg/util/blocklist"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

var ReservedAppIDTXT = PortalRegistry.Register(resource.NewlineJoinedDescriptor{
	Path: "reserved_app_id.txt",
	Parse: func(data []byte) (interface{}, error) {
		return blocklist.New(string(data))
	},
})
