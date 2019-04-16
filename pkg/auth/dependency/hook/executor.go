package hook

import (
	"errors"
	"time"

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
		return err
	}

	if resp.StatusCode == 200 {
		err = resp.Body.FromJsonTo(user)
	} else {
		var body string
		var bodyErr error
		body, bodyErr = resp.Body.ToString()
		if bodyErr != nil {
			return bodyErr
		}
		err = errors.New(body)
	}

	return err
}
