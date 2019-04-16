package hook

import (
	"github.com/skygeario/skygear-server/pkg/auth/response"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

type hookStoreImpl struct {
	authHooks []config.AuthHook
	executor  Executor
}

func NewHookProvider(authHooks []config.AuthHook, executor Executor) Store {
	return &hookStoreImpl{
		authHooks: authHooks,
		executor:  executor,
	}
}

func (h hookStoreImpl) ExecBeforeHooksByEvent(event string, user *response.User) error {
	hooks := h.getHooksByEvent(event)
	for _, v := range hooks {
		err := h.execHook(v, user)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h hookStoreImpl) ExecAfterHooksByEvent(event string, user response.User) error {
	hooks := h.getHooksByEvent(event)
	for _, v := range hooks {
		if err := h.execHook(v, &user); err != nil {
			return err
		}
	}
	return nil
}

func (h hookStoreImpl) getHooksByEvent(event string) []Hook {
	hooks := make([]Hook, 0)
	if len(h.authHooks) == 0 {
		return hooks
	}
	for _, v := range h.authHooks {
		if v.Event == event {
			hooks = append(hooks, Hook{
				Async:   v.Async,
				URL:     v.URL,
				TimeOut: v.TimeOut,
			})
		}
	}
	return hooks
}

func (h hookStoreImpl) execHookImpl(url string, timeOut int, user *response.User) error {
	err := h.executor.ExecHook(url, timeOut, user)
	return err
}

func (h hookStoreImpl) execHook(hook Hook, user *response.User) error {
	if hook.Async {
		go h.execHookImpl(hook.URL, hook.TimeOut, user)
		return nil
	}
	return h.execHookImpl(hook.URL, hook.TimeOut, user)
}
