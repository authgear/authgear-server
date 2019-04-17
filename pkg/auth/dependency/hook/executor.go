package hook

import (
	"time"

	"github.com/skygeario/skygear-server/pkg/core/skyerr"

	"github.com/franela/goreq"
	"github.com/skygeario/skygear-server/pkg/auth/response"
)

type ExecutorImpl struct{}

func (m ExecutorImpl) ExecHook(url string, timeOut int, user *response.User) error {
	// TODO: set timeout
	req := goreq.Request{
		Method:      "POST",
		Uri:         url,
		Body:        user,
		Accept:      "application/json",
		ContentType: "application/json",
		Timeout:     time.Duration(timeOut) * time.Second,
	}

	var err error
	var resp *goreq.Response
	if resp, err = req.Do(); err != nil {
		return handleReqErr(err)
	}

	if resp.StatusCode != 200 {
		return handleRespErr(resp)
	}

	return resp.Body.FromJsonTo(user)
}

func handleReqErr(err error) error {
	if goreqerr, ok := err.(*goreq.Error); ok {
		if goreqerr.Timeout() {
			return skyerr.NewError(skyerr.HookTimeOut, "Hook time out")
		}
	}
	return err
}

func handleRespErr(resp *goreq.Response) error {
	var body string
	var bodyErr error
	body, bodyErr = resp.Body.ToString()
	if bodyErr != nil {
		return bodyErr
	}
	return skyerr.NewError(skyerr.UnexpectedError, body)
}
