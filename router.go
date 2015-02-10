package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/oursky/ourd/handlers"
)

type requestPayload struct {
	Action   string `json:"action"`
	APIKey   string `json:"api_key"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (p *requestPayload) RouteAction() string {
	return p.Action
}

// Router to dispatch HTTP request to respective handler
type Router struct {
	actions map[string]actionHandler
}

type actionHandler struct {
	Action  string
	Handler func(response handlers.Responser, payload handlers.Payloader)
}

// NewRouter is factory for Router
func NewRouter() *Router {
	return &Router{actions: make(map[string]actionHandler)}
}

// Map to register action to handle mapping
func (r *Router) Map(action string, handle func(handlers.Responser, handlers.Payloader)) {
	var actionHandler actionHandler
	actionHandler.Action = action
	actionHandler.Handler = handle
	r.actions[action] = actionHandler
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var (
		httpStatus = http.StatusOK
		reqJSON    requestPayload
		errString  string
	)
	defer func() {
		if httpStatus != http.StatusOK {
			w.Write([]byte(errString))
		}
	}()
	body, _ := ioutil.ReadAll(req.Body)
	if err := json.Unmarshal(body, &reqJSON); err != nil {
		httpStatus = http.StatusBadRequest
		errString = err.Error()
		return
	}
	actionHandler, ok := r.actions[reqJSON.Action]
	if ok {
		actionHandler.Handler(w, &reqJSON)
	} else {
		w.Write([]byte("Unmatched Route"))
	}
}
