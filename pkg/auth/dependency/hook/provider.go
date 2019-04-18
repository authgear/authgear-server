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
	requestID     string
}

func NewHookProvider(
	authHooks []config.AuthHook,
	executor Executor,
	logger *logrus.Entry,
	requestID string,
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
		requestID:     requestID,
	}
}

func (h hookStoreImpl) ExecBeforeHooksByEvent(event string, user *response.User, accessToken string) error {
	hooks := h.authHookStore[event]
	for _, v := range hooks {
		payload, err := NewDefaultAuthPayload(event, *user, h.requestID)
		if err != nil {
			h.logger.Warnf("Fail to generate auth hook payload")
			return err
		}
		respDecoder := AuthRespPayload{
			User: user,
		}
		p := ExecHookParam{
			URL:         v.URL,
			TimeOut:     v.TimeOut,
			AccessToken: accessToken,
			BodyEncoder: payload,
			RespDecoder: &respDecoder,
		}
		err = h.execHook(p, v.Async)
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
		payload, err := NewDefaultAuthPayload(event, user, h.requestID)
		if err != nil {
			h.logger.Warnf("Fail to generate auth hook payload")
			return err
		}
		p := ExecHookParam{
			URL:         v.URL,
			TimeOut:     v.TimeOut,
			AccessToken: accessToken,
			BodyEncoder: payload,
		}
		if err := h.execHook(p, v.Async); err != nil {
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

func (h hookStoreImpl) execHook(p ExecHookParam, async bool) error {
	if async {
		// for async hook, omit result from hook
		go h.execHookImpl(p)
		return nil
	}
	return h.execHookImpl(p)
}
