package hook

import (
	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/auth/response"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

type hookStoreImpl struct {
	authHookStore map[string][]Hook
	executor      Executor
	logger        *logrus.Entry
}

func NewHookProvider(
	authHooks []config.AuthHook,
	executor Executor,
	logger *logrus.Entry,
) Store {
	authHookStore := make(map[string][]Hook)
	for _, v := range authHooks {
		hook := Hook{
			Async:   v.Async,
			URL:     v.URL,
			TimeOut: v.TimeOut,
		}

		if hooks, ok := authHookStore[v.Event]; ok {
			authHookStore[v.Event] = append(hooks, hook)
		} else {
			authHookStore[v.Event] = []Hook{hook}
		}
	}
	return &hookStoreImpl{
		authHookStore: authHookStore,
		executor:      executor,
		logger:        logger,
	}
}

func (h hookStoreImpl) ExecBeforeHooksByEvent(event string, user *response.User) error {
	hooks := h.authHookStore[event]
	for _, v := range hooks {
		err := h.execHook(v, user)
		if err != nil {
			h.logger.Warnf("Exec %v(%v) hook failed: %v", event, v.URL, err)
			return err
		}
	}
	return nil
}

func (h hookStoreImpl) ExecAfterHooksByEvent(event string, user response.User) error {
	hooks := h.authHookStore[event]
	for _, v := range hooks {
		if err := h.execHook(v, &user); err != nil {
			h.logger.Warnf("Exec %v(%v) hook failed: %v", event, v.URL, err)
			return err
		}
	}
	return nil
}

func (h hookStoreImpl) execHookImpl(url string, timeOut int, user *response.User) error {
	err := h.executor.ExecHook(url, timeOut, user)
	return err
}

func (h hookStoreImpl) execHook(hook Hook, user *response.User) error {
	if hook.Async {
		// for async hook, result is omit
		go h.execHookImpl(hook.URL, hook.TimeOut, user)
		return nil
	}
	return h.execHookImpl(hook.URL, hook.TimeOut, user)
}
