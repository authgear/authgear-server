package workflow

import (
	"context"
	"net/http"
)

type CookieGetter interface {
	GetCookies(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]*http.Cookie, error)
}
