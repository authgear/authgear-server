package handler

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
)

type Factory interface {
	authz.PolicyProvider
	NewHandler(request *http.Request) http.Handler
}
