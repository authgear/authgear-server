package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// Router to dispatch HTTP request to respective handler
type Router struct {
	actions map[string]actionHandler
}

type actionHandler struct {
	Action  string
	Handler http.Handler
}

type bodyReader struct {
	*bytes.Buffer
}

func (m bodyReader) Close() error {
	return nil
}

type requestJSON struct {
	Action   interface{} `json:"action"`
	APIKey   interface{} `json:"api_key"`
	Email    interface{} `json:"email"`
	Password interface{} `json:"password"`
}

// NewRouter is factory for Router
func NewRouter() *Router {
	return &Router{actions: make(map[string]actionHandler)}
}

// Map to register action to handle mapping
func (r *Router) Map(action string, handle func(http.ResponseWriter, *http.Request)) {
	var actionHandler actionHandler
	actionHandler.Action = action
	actionHandler.Handler = http.HandlerFunc(handle)
	r.actions[action] = actionHandler
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var (
		httpStatus = http.StatusOK
		reqJSON    requestJSON
		errString  string
	)
	defer func() {
		if httpStatus != http.StatusOK {
			w.Write([]byte(errString))
		}
	}()
	body, _ := ioutil.ReadAll(req.Body)
	nextBody := bodyReader{bytes.NewBuffer(body)}
	if err := json.Unmarshal(body, &reqJSON); err != nil {
		httpStatus = http.StatusBadRequest
		errString = err.Error()
		return
	}
	actionHandler, ok := r.actions[reqJSON.Action.(string)]
	if ok {
		req.Body = nextBody
		actionHandler.Handler.ServeHTTP(w, req)
	} else {
		w.Write([]byte("Unmatched Route"))
	}
}
