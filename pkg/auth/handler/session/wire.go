package session

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/deps"
)

func newResolveHandler(p *deps.RequestProvider) http.Handler {
	return (*ResolveHandler)(nil)
}
