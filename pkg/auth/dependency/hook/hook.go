package hook

import (
	"github.com/skygeario/skygear-server/pkg/auth/response"
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
	ExecBeforeHooksByEvent(event string, user *response.User, accessToken string) error
	ExecAfterHooksByEvent(event string, user response.User, accessToken string) error
}

type Executor interface {
	ExecHook(p ExecHookParam) error
}
