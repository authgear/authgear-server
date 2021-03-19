package loginid

import (
	"github.com/authgear/authgear-server/pkg/util/blocklist"
	"github.com/authgear/authgear-server/pkg/util/matchlist"
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

var EmailDomainBlockListTXT = resource.RegisterResource(resource.NewlineJoinedDescriptor{
	Path: "email_domain_blocklist.txt",
	Parse: func(data []byte) (interface{}, error) {
		return matchlist.New(string(data), true, false)
	},
})

// FreeEmailProviderDomainsTXT is provided by
// https://gist.github.com/tbrianjones/5992856/93213efb652749e226e69884d6c048e595c1280a
var FreeEmailProviderDomainsTXT = resource.RegisterResource(resource.NewlineJoinedDescriptor{
	Path: "free_email_provider_domain_list.txt",
	Parse: func(data []byte) (interface{}, error) {
		return matchlist.New(string(data), true, false)
	},
})

var EmailDomainAllowListTXT = resource.RegisterResource(resource.NewlineJoinedDescriptor{
	Path: "email_domain_allowlist.txt",
	Parse: func(data []byte) (interface{}, error) {
		return matchlist.New(string(data), true, false)
	},
})
