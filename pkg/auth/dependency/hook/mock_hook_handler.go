package hook

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

func NewMockHookUpdateMetaHandler(metadata userprofile.Data) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		decoder := json.NewDecoder(req.Body)
		var payload AuthPayload
		err := decoder.Decode(&payload)
		if err != nil {
			panic(err)
		}

		user := payload.Data
		user["metadata"] = metadata

		body, err := json.Marshal(user)
		if err != nil {
			panic(err)
		}

		rw.WriteHeader(http.StatusOK)
		rw.Write(body)
	}))

	return server
}

func NewMockHookErrorHandler(errorCode skyerr.ErrorCode, errorMsg string) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusInternalServerError)

		errResp := ErrorResp{
			Code:    errorCode,
			Message: errorMsg,
		}

		if bytes, err := json.Marshal(errResp); err == nil {
			rw.Write(bytes)
		}
		return
	}))

	return server
}
