package hook

import (
	"time"

	"github.com/skygeario/skygear-server/pkg/core/skyerr"

	"github.com/franela/goreq"
)

type ExecutorImpl struct{}

type ExecHookParam struct {
	URL         string
	TimeOut     int
	AccessToken string
	BodyEncoder ReqBodyEncoder
	RespDecoder RespDecoder
}

func (m ExecutorImpl) ExecHook(p ExecHookParam) error {
	req := goreq.Request{
		Method:      "POST",
		Uri:         p.URL,
		Body:        p.BodyEncoder.Encode(),
		Accept:      "application/json",
		ContentType: "application/json",
		Timeout:     time.Duration(p.TimeOut) * time.Second,
	}

	if p.AccessToken != "" {
		req.AddHeader("X-Skygear-Access-Token", p.AccessToken)
	}

	var resp *goreq.Response
	var err error
	if resp, err = req.Do(); err != nil {
		return handleReqErr(err)
	}

	if resp.StatusCode != 200 {
		return handleRespErr(resp)
	}

	if p.RespDecoder != nil {
		return p.RespDecoder.Decode(resp.Body)
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
