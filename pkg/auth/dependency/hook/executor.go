package hook

import (
	"time"

	"github.com/skygeario/skygear-server/pkg/core/skyerr"

	"github.com/franela/goreq"
	"github.com/skygeario/skygear-server/pkg/auth/response"
)

type ExecutorImpl struct{}

type ExecHookParam struct {
	URL         string
	TimeOut     int
	User        *response.User
	AccessToken string
}

func (m ExecutorImpl) ExecHook(p ExecHookParam) error {
	// TODO: set timeout
	req := goreq.Request{
		Method:      "POST",
		Uri:         p.URL,
		Body:        p.User,
		Accept:      "application/json",
		ContentType: "application/json",
		Timeout:     time.Duration(p.TimeOut) * time.Second,
	}

	if p.AccessToken != "" {
		req.AddHeader("X-Skygear-Access-Token", p.AccessToken)
	}

	var err error
	var resp *goreq.Response
	if resp, err = req.Do(); err != nil {
		return handleReqErr(err)
	}

	if resp.StatusCode != 200 {
		return handleRespErr(resp)
	}

	return resp.Body.FromJsonTo(p.User)
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
