package hook

import (
	"io"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/auth/response"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

type hookStoreImpl struct {
	hookStore map[string][]Hook
	executor  Executor
	logger    *logrus.Entry
	requestID string
	path      string
	payload   io.ReadCloser
}

func NewHookProvider(
	hooks []config.Hook,
	executor Executor,
	logger *logrus.Entry,
	requestID string,
) Store {
	hookStore := make(map[string][]Hook)
	for _, v := range hooks {
		hook := Hook{
			Async:   v.Async,
			URL:     v.URL,
			TimeOut: v.TimeOut,
		}

		if hooks, ok := hookStore[v.Event]; ok {
			hookStore[v.Event] = append(hooks, hook)
		} else {
			hookStore[v.Event] = []Hook{hook}
		}
	}
	return &hookStoreImpl{
		hookStore: hookStore,
		executor:  executor,
		logger:    logger,
		requestID: requestID,
	}
}

func (h hookStoreImpl) WithRequest(request *http.Request) Store {
	h.path = request.URL.Path
	return h
}

func (h hookStoreImpl) ExecBeforeHooksByEvent(event string, reqPayload interface{}, user *response.User, accessToken string) error {
	respDecoder := AuthRespPayload{
		User: user,
	}
	return h.execHooks(event, reqPayload, user, accessToken, &respDecoder)
}

func (h hookStoreImpl) ExecAfterHooksByEvent(event string, reqPayload interface{}, user response.User, accessToken string) error {
	return h.execHooks(event, reqPayload, &user, accessToken, nil)
}

func (h hookStoreImpl) execHooks(
	event string,
	reqPayload interface{},
	user *response.User,
	accessToken string,
	respDecoder *AuthRespPayload,
) error {
	hooks := h.hookStore[event]
	for _, v := range hooks {
		payload, err := NewDefaultAuthPayload(event, *user, h.requestID, h.path, reqPayload)
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
		if respDecoder != nil {
			p.RespDecoder = respDecoder
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
