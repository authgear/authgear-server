package loginid

import (
	"context"

	"github.com/authgear/authgear-server/pkg/util/blocklist"
	"github.com/authgear/authgear-server/pkg/util/matchlist"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

type ResourceManager interface {
	Read(ctx context.Context, desc resource.Descriptor, view resource.View) (interface{}, error)
}

var ReservedNameTXT = resource.RegisterResource(resource.NewlineJoinedDescriptor{
	Path: "reserved_name.txt",
	Parse: func(data []byte) (interface{}, error) {
		return blocklist.New(string(data))
	},
})

var UsernameExcludedKeywordsTXT = resource.RegisterResource(resource.NewlineJoinedDescriptor{
	Path: "username_excluded_keywords.txt",
	Parse: func(data []byte) (interface{}, error) {
		return matchlist.New(string(data), true, true)
	},
})

var EmailDomainBlockListTXT = resource.RegisterResource(resource.NewlineJoinedDescriptor{
	Path: "email_domain_blocklist.txt",
	Parse: func(data []byte) (interface{}, error) {
		return matchlist.New(string(data), true, false)
	},
})

// FreeEmailProviderDomainsTXT is provided by
// https://github.com/Kikobeats/free-email-domains/blob/master/domains.json
var FreeEmailProviderDomainsTXT = resource.RegisterResource(resource.NewlineJoinedDescriptor{
	Path: "free_email_provider_domain_list.txt",
	Parse: func(data []byte) (interface{}, error) {
		return matchlist.New(string(data), true, false)
	},
})

// DisposableEmailDomainsTXT is provided by
// https://github.com/disposable-email-domains/disposable-email-domains/blob/main/disposable_email_blocklist.conf
var DisposableEmailDomainsTXT = resource.RegisterResource(resource.NewlineJoinedDescriptor{
	Path: "disposable_email_domain_list.txt",
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
