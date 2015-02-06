package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Router struct {
	actions map[string]ActionHandler
}

type ActionHandler struct {
	Action  string
	Handler http.Handler
}

type bodyReader struct {
    *bytes.Buffer
}

func (m bodyReader) Close() error {
	return nil
}

func NewRouter() *Router {
	return &Router{actions: make(map[string]ActionHandler)}
}

func (r *Router) HandleFunc(action string, handle func(http.ResponseWriter, *http.Request)) {
	var actionHandler ActionHandler
	actionHandler.Action = action
	actionHandler.Handler = http.HandlerFunc(handle)
	r.actions[action] = actionHandler
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var (
		httpStatus = http.StatusOK
		reqJson    RequestJson
		errString  string
	)
	defer func() {
		if httpStatus != http.StatusOK {
			w.Write([]byte(errString))
		}
	}()
	body, _ := ioutil.ReadAll(req.Body)
	nextBody := bodyReader{bytes.NewBuffer(body)}
	if err := json.Unmarshal(body, &reqJson); err != nil {
		httpStatus = http.StatusBadRequest
		errString = err.Error()
		return
	}
	actionHandler, ok := r.actions[reqJson.Action.(string)]
	if ok {
		req.Body = nextBody
		actionHandler.Handler.ServeHTTP(w, req)
	} else {
		w.Write([]byte("Unmatched Route"))
	}
}
