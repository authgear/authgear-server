package resource

import (
	"github.com/authgear/authgear-server/pkg/util/blocklist"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

var ReservedAppIDTXT = PortalRegistry.Register(resource.JoinedFile{
	Name:      "reserved_app_id.txt",
	Separator: []byte("\n"),
	ParseFn: func(data []byte) (interface{}, error) {
		return blocklist.New(string(data))
	},
})
