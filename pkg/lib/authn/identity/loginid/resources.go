package loginid

import (
	"github.com/authgear/authgear-server/pkg/util/blocklist"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

type ResourceManager interface {
	Read(desc resource.Descriptor, args map[string]interface{}) (*resource.MergedFile, error)
}

var ReservedNameTXT = resource.RegisterResource(resource.JoinedFile{
	Name:      "reserved_name.txt",
	Separator: []byte("\n"),
	ParseFn: func(data []byte) (interface{}, error) {
		return blocklist.New(string(data))
	},
})
