package hook

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/model"
)

type Hook struct {
	Async   bool
	URL     string
	TimeOut int
}

const (
	// BeforeSignup is event name of before signup
	BeforeSignup = "before_signup"
	// AfterSignup is event name of after signup
	AfterSignup = "after_signup"
)

type Store interface {
	WithRequest(request *http.Request) Store
	ExecBeforeHooksByEvent(event string, reqPayload interface{}, user *model.User, accessToken string) error
	ExecAfterHooksByEvent(event string, reqPayload interface{}, user model.User, accessToken string) error
}

type Executor interface {
	ExecHook(p ExecHookParam) error
}
