package router

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/oursky/ourd/oderr"
)

// Handler specifies the function signature of a request handler function
type Handler func(*Payload, *Response)

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
type Processor func(*Payload, *Response) (int, error)

// NewRouter is factory for Router
func NewRouter() *Router {
	return &Router{actions: make(map[string]pipeline)}
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
		httpStatus = http.StatusOK
		reqJSON    map[string]interface{}
		errString  string
	)
	defer func() {
		if httpStatus != http.StatusOK {
			w.WriteHeader(httpStatus)
			w.Write([]byte(errString))
		}
	}()
	body, _ := ioutil.ReadAll(req.Body)
	if err := json.Unmarshal(body, &reqJSON); err != nil {
		httpStatus = http.StatusBadRequest
		errString = err.Error()
		return
	}
	payload := Payload{
		Meta: map[string]interface{}{},
		Data: reqJSON,
	}
	pipeline, ok := r.actions[payload.RouteAction()]
	if ok {
		var resp Response
		for _, p := range pipeline.Preprocessors {
			if s, err := p(&payload, &resp); err != nil {
				httpStatus = s

				if mwErr, ok := err.(oderr.Error); ok {
					b, err := json.Marshal(struct {
						Code    oderr.ErrCode `json:"code"`
						Message string        `json:"message"`
					}{mwErr.Code(), mwErr.Message()})

					if err == nil {
						errString = string(b)
						return
					}
				}

				errString = err.Error()
				return
			}
		}
		pipeline.Handler(&payload, &resp)
		b, err := json.Marshal(resp)
		if err != nil {
			panic("Response Error: " + err.Error())
		}
		w.Write(b)
	} else {
		httpStatus = http.StatusNotFound
		errString = "Unmatched Route"
		return
	}
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
