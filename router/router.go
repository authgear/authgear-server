package router

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/oursky/skygear/skyerr"
)

// pipeline encapsulates a transformation which a request will come throught
// from preprocessors to the actual handler. (and postprocessor later)
type pipeline struct {
	Action        string
	Preprocessors []Processor
	Handler
}

// Router to dispatch HTTP request to respective handler
type Router struct {
	actions map[string]pipeline
}

// Processor specifies the function signature for a Preprocessor
type Processor func(*Payload, *Response) int

// NewRouter is factory for Router
func NewRouter() *Router {
	return &Router{
		map[string]pipeline{},
	}
}

// Map to register action to handle mapping
func (r *Router) Map(action string, handler Handler, preprocessors ...Processor) {
	r.actions[action] = pipeline{
		Action:        action,
		Preprocessors: preprocessors,
		Handler:       handler,
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var (
		httpStatus    = http.StatusOK
		resp          Response
		handler       Handler
		preprocessors []Processor
		payload       *Payload
	)

	resp.writer = w
	defer func() {
		if !resp.written {
			resp.Header().Set("Content-Type", "application/json")
			if resp.Err != nil && httpStatus >= 200 && httpStatus <= 299 {
				resp.writer.WriteHeader(http.StatusBadRequest)
			} else {
				resp.writer.WriteHeader(httpStatus)
			}
			if err := resp.WriteEntity(resp); err != nil {
				panic(err)
			}
		}
	}()

	var err error
	payload, err = newPayloadForJSONHandler(req)
	if err != nil {
		httpStatus = http.StatusBadRequest
		resp.Err = skyerr.NewRequestJSONInvalidErr(err)
		return
	}

	handler, preprocessors = r.matchJSONHandler(payload)

	if handler == nil {
		httpStatus = http.StatusNotFound
		resp.Err = skyerr.NewRequestInvalidErr(errors.New("route unmatched"))
	} else {
		for _, p := range preprocessors {
			httpStatus = p(payload, &resp)
			if resp.Err != nil {
				if httpStatus == 200 {
					httpStatus = 500
				}
				if _, ok := resp.Err.(skyerr.Error); !ok {
					resp.Err = skyerr.NewUnknownErr(resp.Err)
				}
				return
			}
		}
		handler(payload, &resp)
	}
}

func (r *Router) matchJSONHandler(p *Payload) (h Handler, pp []Processor) {
	if pipeline, ok := r.actions[p.RouteAction()]; ok {
		h = pipeline.Handler
		pp = pipeline.Preprocessors
	}
	return
}

func newPayloadForJSONHandler(req *http.Request) (p *Payload, err error) {
	reqBody := req.Body
	if reqBody == nil {
		reqBody = ioutil.NopCloser(bytes.NewReader(nil))
	}

	data := map[string]interface{}{}
	if jsonErr := json.NewDecoder(reqBody).Decode(&data); jsonErr != nil && jsonErr != io.EOF {
		err = jsonErr
		return
	}

	p = &Payload{
		Data: data,
		Meta: map[string]interface{}{},
	}

	if apiKey := req.Header.Get("X-SkygearApi-Key"); apiKey != "" {
		p.Data["api_key"] = apiKey
	}
	if accessToken := req.Header.Get("X-SkygearAccess-Token"); accessToken != "" {
		p.Data["access_token"] = accessToken
	}

	return
}

// CheckAuth will check on the AccessToken, attach DB/RequestID to the response
// This is a no-op if the request action belong to "auth:" group
func CheckAuth(payload *Payload, response *Response) (status int, err error) {
	log.Println("CheckAuth")

	token := payload.AccessToken()

	if token == "validToken" {
		log.Println("CheckAuth -> validToken, ", token)
		return http.StatusOK, nil
	}
	log.Println("CheckAuth -> inValidToken, ", token)
	return http.StatusUnauthorized, errors.New("Unauthorized request")
}
