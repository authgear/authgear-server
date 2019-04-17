package hook

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/skygeario/skygear-server/pkg/auth/response"
)

type MockExecutorResult struct {
	User     response.User
	ErrorMsg string
}

func NewMockHookHandler(result MockExecutorResult) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if result.ErrorMsg != "" {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(result.ErrorMsg))
			return
		}

		body, err := json.Marshal(result.User)
		if err != nil {
			panic(err)
		}

		rw.WriteHeader(http.StatusOK)
		rw.Write(body)
	}))

	return server
}
