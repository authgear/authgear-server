package router

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"

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

// pathRoute is the path matching version of pipeline. Instead of storing the action
// to match, it stores a regexp to match against request URL.
type pathRoute struct {
	Regexp        *regexp.Regexp
	Preprocessors []Processor
	Handler
}

// Router to dispatch HTTP request to respective handler
type Router struct {
	paths   []pathRoute
	actions map[string]pipeline
}

// Processor specifies the function signature for a Preprocessor
type Processor func(*Payload, *Response) int

// NewRouter is factory for Router
func NewRouter() *Router {
	return &Router{actions: make(map[string]pipeline)}
}

// Handle registers a handler by requests URL's path. Pattern is a regexp
// that defines a match.
func (r *Router) Handle(pattern string, handler Handler, preprocessors ...Processor) {
	r.paths = append(r.paths, pathRoute{
		Regexp:        regexp.MustCompile(`\A/` + pattern + `\z`),
		Preprocessors: preprocessors,
		Handler:       handler,
	})
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
		resp       Response
	)
	resp.writer = w
	defer func() {
		if !resp.written {
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

	var (
		handler       Handler
		preprocessors []Processor
		payload       Payload
	)
	payload.Req = req
	payload.Meta = map[string]interface{}{}
	payload.Data = map[string]interface{}{}

	// match by URL first
	matched := false
	for _, pathRoute := range r.paths {
		indices := pathRoute.Regexp.FindAllStringSubmatchIndex(req.URL.Path, -1)
		if len(indices) > 0 {
			matched = true
			handler = pathRoute.Handler
			preprocessors = pathRoute.Preprocessors

			submatches := submatchesFromIndices(req.URL.Path, indices)
			payload.Params = submatches
			fillPayloadByRequest(&payload, req)
			break
		}
	}

	if !matched {
		// match by JSON body then
		reqBody := req.Body
		if reqBody == nil {
			reqBody = ioutil.NopCloser(bytes.NewReader(nil))
		}
		if err := json.NewDecoder(reqBody).Decode(&payload.Data); err != nil && err != io.EOF {
			httpStatus = http.StatusBadRequest
			resp.Err = oderr.NewRequestJSONInvalidErr(err)
			return
		}

		if pipeline, ok := r.actions[payload.RouteAction()]; ok {
			matched = true
			handler = pipeline.Handler
			preprocessors = pipeline.Preprocessors
		}
	}

	if matched {
		for _, p := range preprocessors {
			httpStatus = p(&payload, &resp)
			if resp.Err != nil {
				if httpStatus == 200 {
					httpStatus = 500
				}
				if _, ok := resp.Err.(oderr.Error); !ok {
					resp.Err = oderr.NewUnknownErr(resp.Err)
				}
				return
			}
		}
		handler(&payload, &resp)
	} else {
		httpStatus = http.StatusNotFound
		resp.Err = oderr.NewRequestInvalidErr(errors.New("route unmatched"))
	}
}

func submatchesFromIndices(s string, indices [][]int) (submatches []string) {
	submatches = make([]string, 0, len(indices))
	for _, pairs := range indices {
		for i := 2; i < len(pairs); i += 2 {
			submatches = append(submatches, s[pairs[i]:pairs[i+1]])
		}
	}
	return
}

func fillPayloadByRequest(payload *Payload, req *http.Request) error {
	if apiKey := req.Header.Get("X-Ourd-API-Key"); apiKey != "" {
		payload.Data["api_key"] = apiKey
	}
	if accessToken := req.Header.Get("X-Ourd-Access-Token"); accessToken != "" {
		payload.Data["access_token"] = accessToken
	}

	return nil
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
