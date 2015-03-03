package router

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/oursky/ourd/oddb"
)

// Router to dispatch HTTP request to respective handler
type Router struct {
	actions      map[string]actionHandler
	preprocessor []Processor
}

type Handler func(*Payload, *Response)

type actionHandler struct {
	Action  string
	Handler func(*Payload, *Response)
}

// Processor specifies the function signature for a Preprocessor
type Processor func(*Payload, *Response) (int, error)

// NewRouter is factory for Router
func NewRouter() *Router {
	return &Router{actions: make(map[string]actionHandler)}
}

// Map to register action to handle mapping
func (r *Router) Map(action string, handle func(*Payload, *Response)) {
	var actionHandler actionHandler
	actionHandler.Action = action
	actionHandler.Handler = handle
	r.actions[action] = actionHandler
}

// Preprocess register a Processor func to be called before the actual hanlder
func (r *Router) Preprocess(p Processor) {
	r.preprocessor = append(r.preprocessor, p)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var (
		httpStatus = http.StatusOK
		reqJSON    interface{}
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
		make(map[string]interface{}),
		reqJSON.(map[string]interface{}),
		nil,
	}
	actionHandler, ok := r.actions[payload.RouteAction()]
	if ok {
		var resp Response
		for _, p := range r.preprocessor {
			if s, err := p(&payload, &resp); err != nil {
				httpStatus = s
				errString = err.Error()
				return
			}
		}
		actionHandler.Handler(&payload, &resp)
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
	if payload.IsAuth() {
		log.Println("CheckAuth -> IsAuth")
		return http.StatusOK, nil
	}
	token := payload.AccessToken()
	if token == "validToken" {
		log.Println("CheckAuth -> validToken, ", token)
		return http.StatusOK, nil
	}
	log.Println("CheckAuth -> inValidToken, ", token)
	return http.StatusUnauthorized, errors.New("Unauthorized request")
}

// AssignDBConn will assign the DBConn to the payload
func AssignDBConn(payload *Payload, response *Response) (status int, err error) {
	log.Println("GetDB Conn")
	c, err := oddb.Open("fs", "_", "data")
	if err != nil {
		return http.StatusServiceUnavailable, err
	}
	payload.DBConn = c
	log.Println("GetDB Conn OK")
	return http.StatusOK, nil
}
