package hook

import (
	"encoding/json"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/skyerr"

	"github.com/franela/goreq"
	"github.com/skygeario/skygear-server/pkg/auth/response"
)

type ExecutorImpl struct{}

type ExecHookParam struct {
	URL            string
	TimeOut        int
	Event          string
	User           *response.User
	AccessToken    string
	DecodeRespUser bool
}

func (m ExecutorImpl) ExecHook(p ExecHookParam) error {
	var payload Payload
	var err error
	err = constructPayload(p, &payload)
	if err != nil {
		return err
	}

	req := goreq.Request{
		Method:      "POST",
		Uri:         p.URL,
		Body:        payload,
		Accept:      "application/json",
		ContentType: "application/json",
		Timeout:     time.Duration(p.TimeOut) * time.Second,
	}

	if p.AccessToken != "" {
		req.AddHeader("X-Skygear-Access-Token", p.AccessToken)
	}

	var resp *goreq.Response
	if resp, err = req.Do(); err != nil {
		return handleReqErr(err)
	}

	if resp.StatusCode != 200 {
		return handleRespErr(resp)
	}

	if p.DecodeRespUser {
		return resp.Body.FromJsonTo(p.User)
	}

	return nil
}

func constructPayload(p ExecHookParam, payload *Payload) error {
	// convert user to map[string]interface{}
	var data map[string]interface{}
	var userBytes []byte
	var err error
	userBytes, err = json.Marshal(p.User)
	if err != nil {
		return err
	}
	err = json.Unmarshal(userBytes, &data)
	if err != nil {
		return err
	}

	*payload = Payload{
		Event: p.Event,
		Data:  data,
	}

	return nil
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
