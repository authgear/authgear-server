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

func (h hookStoreImpl) ExecBeforeHooksByEvent(event string, user *response.User, accessToken string) error {
	hooks := h.authHookStore[event]
	for _, v := range hooks {
		err := h.execHook(v, user, accessToken)
		if err != nil {
			h.logger.Warnf("Exec %v(%v) hook failed: %v", event, v.URL, err)
			return err
		}
	}
	return nil
}

func (h hookStoreImpl) ExecAfterHooksByEvent(event string, user response.User, accessToken string) error {
	hooks := h.authHookStore[event]
	for _, v := range hooks {
		if err := h.execHook(v, &user, accessToken); err != nil {
			h.logger.Warnf("Exec %v(%v) hook failed: %v", event, v.URL, err)
			return err
		}
	}
	return nil
}

func (h hookStoreImpl) execHookImpl(p ExecHookParam) error {
	err := h.executor.ExecHook(p)
	return err
}

func (h hookStoreImpl) execHook(hook Hook, user *response.User, accessToken string) error {
	p := ExecHookParam{
		URL:         hook.URL,
		TimeOut:     hook.TimeOut,
		User:        user,
		AccessToken: accessToken,
	}
	if hook.Async {
		// for async hook, result is omit
		go h.execHookImpl(p)
		return nil
	}
	return h.execHookImpl(p)
}
