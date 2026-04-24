package useragentblocklist

import (
	"github.com/authgear/authgear-server/pkg/util/blocklist"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

var UserAgentBlockListTXT = resource.RegisterResource(resource.NewlineJoinedDescriptor{
	Path: "user_agent_blocklist.txt",
	Parse: func(data []byte) (interface{}, error) {
		return blocklist.New(string(data))
	},
})
