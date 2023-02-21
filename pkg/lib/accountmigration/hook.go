package accountmigration

import "net/url"

type Hook interface {
	Call(u *url.URL, hookReq *HookRequest) (*HookResponse, error)
}
