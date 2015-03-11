package router

import (
	"encoding/json"
	"errors"
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
type Processor func(*Payload, *Response) int

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
		resp       Response
	)
	defer func() {
		w.WriteHeader(httpStatus)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			panic(err)
		}
	}()

	if err := json.NewDecoder(req.Body).Decode(&reqJSON); err != nil {
		httpStatus = http.StatusBadRequest
		resp.Err = oderr.NewFmt(oderr.RequestInvalidErr, err.Error())
		return
	}

	payload := Payload{
		Meta: map[string]interface{}{},
		Data: reqJSON,
	}

	if pipeline, ok := r.actions[payload.RouteAction()]; ok {
		for _, p := range pipeline.Preprocessors {
			httpStatus = p(&payload, &resp)
			if resp.Err != nil {
				if httpStatus == 200 {
					httpStatus = 500
				}
				if _, ok := resp.Err.(oderr.Error); !ok {
					resp.Err = oderr.New(oderr.UnknownErr, resp.Err.Error())
				}
				return
			}
		}
		pipeline.Handler(&payload, &resp)
	} else {
		httpStatus = http.StatusNotFound
		resp.Err = oderr.New(oderr.RequestInvalidErr, "Unmatched Route")
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
