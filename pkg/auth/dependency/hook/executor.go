package hook

import (
	"io"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/skyerr"

	"github.com/franela/goreq"
)

type ExecutorImpl struct{}

type ReqBodyEncoder interface {
	Encode() interface{}
}

type RespDecoder interface {
	Decode(r io.Reader) error
}

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

type ErrorResp struct {
	Code    skyerr.ErrorCode       `json:"code,omitempty"`
	Message string                 `json:"message"`
	Info    map[string]interface{} `json:"info,omitempty"`
}

func handleRespErr(resp *goreq.Response) error {
	var errResp ErrorResp
	err := resp.Body.FromJsonTo(&errResp)
	if err != nil {
		return err
	}
	if errResp.Code == 0 {
		errResp.Code = skyerr.UnexpectedError
	}
	return skyerr.NewErrorWithInfo(errResp.Code, errResp.Message, errResp.Info)
}
