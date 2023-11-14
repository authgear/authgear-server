package authflowclienthandlers

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type TutorialCookie interface {
	Pop(r *http.Request, rw http.ResponseWriter, name httputil.TutorialCookieName) bool
}
