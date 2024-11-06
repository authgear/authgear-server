package accountmigration

import (
	"context"
	"net/url"
)

type Hook interface {
	Call(ctx context.Context, u *url.URL, hookReq *HookRequest) (*HookResponse, error)
}
