package loginid

import (
	"strings"

	"github.com/authgear/authgear-server/pkg/util/resource"
)

type ResourceManager interface {
	Read(desc resource.Descriptor, args map[string]interface{}) (*resource.MergedFile, error)
}

// TODO(resource): more specific merging?

var ReservedNameTXT = resource.RegisterResource(resource.SimpleFile{
	Name: "reserved_name.txt",
	ParseFn: func(data []byte) (interface{}, error) {
		reservedWords := strings.Split(string(data), "\n")
		return ReservedNameData(reservedWords), nil
	},
})
