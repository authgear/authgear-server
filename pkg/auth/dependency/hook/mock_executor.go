package hook

import (
	"github.com/skygeario/skygear-server/pkg/auth/response"
)

type MockExecutorResult struct {
	User  response.User
	Error error
}

type mockExecutorImpl struct {
	results map[string]MockExecutorResult
}

func NewMockExecutorImpl(results map[string]MockExecutorResult) Executor {
	return mockExecutorImpl{
		results: results,
	}
}

func (m mockExecutorImpl) ExecHook(url string, timeOut int, user *response.User) error {
	result, ok := m.results[url]
	if !ok {
		return nil
	}

	*user = result.User
	return result.Error
}
