package router

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"

	log "github.com/Sirupsen/logrus"
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
	methodPaths map[string][]pathRoute
	actions     map[string]pipeline
}

// Processor specifies the function signature for a Preprocessor
type Processor func(*Payload, *Response) int

// NewRouter is factory for Router
func NewRouter() *Router {
	return &Router{
		map[string][]pathRoute{},
		map[string]pipeline{},
	}
}

// Handle registers a handler matched by a request's method and URL's path.
// Pattern is a regexp that defines a matched URL.
func (r *Router) Handle(method string, pattern string, handler Handler, preprocessors ...Processor) {
	r.methodPaths[method] = append(r.methodPaths[method], pathRoute{
		Regexp:        regexp.MustCompile(`\A/` + pattern + `\z`),
		Preprocessors: preprocessors,
		Handler:       handler,
	})
}

// GET register a URL handler by method GET
func (r *Router) GET(pattern string, handler Handler, preprocessors ...Processor) {
	r.Handle("GET", pattern, handler, preprocessors...)
}

// POST register a URL handler by method POST
func (r *Router) POST(pattern string, handler Handler, preprocessors ...Processor) {
	r.Handle("POST", pattern, handler, preprocessors...)
}

// PUT register a URL handler by method PUT
func (r *Router) PUT(pattern string, handler Handler, preprocessors ...Processor) {
	r.Handle("PUT", pattern, handler, preprocessors...)
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
		payload       *Payload
	)

	handler, preprocessors, payload = r.matchRawHandler(req)

	if handler == nil {
		var err error
		payload, err = newPayloadForJSONHandler(req)
		if err != nil {
			httpStatus = http.StatusBadRequest
			resp.Err = oderr.NewRequestJSONInvalidErr(err)
			return
		}

		handler, preprocessors = r.matchJSONHandler(payload)
	}

	if handler == nil {
		httpStatus = http.StatusNotFound
		resp.Err = oderr.NewRequestInvalidErr(errors.New("route unmatched"))
	} else {
		for _, p := range preprocessors {
			httpStatus = p(payload, &resp)
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
		handler(payload, &resp)
	}
}

func (r *Router) matchRawHandler(req *http.Request) (h Handler, pp []Processor, p *Payload) {
	for _, pathRoute := range r.methodPaths[req.Method] {
		indices := pathRoute.Regexp.FindAllStringSubmatchIndex(req.URL.Path, -1)
		if len(indices) > 0 {
			h = pathRoute.Handler
			pp = pathRoute.Preprocessors
			p = newPayloadForRawHandler(req, indices)
			break
		}
	}
	return
}

func (r *Router) matchJSONHandler(p *Payload) (h Handler, pp []Processor) {
	if pipeline, ok := r.actions[p.RouteAction()]; ok {
		h = pipeline.Handler
		pp = pipeline.Preprocessors
	}
	return
}

func newPayloadForRawHandler(req *http.Request, paramIndices [][]int) (p *Payload) {
	p = &Payload{
		Req:    req,
		Params: submatchesFromIndices(req.URL.Path, paramIndices),
		Meta:   map[string]interface{}{},
		Data:   map[string]interface{}{},
	}

	if apiKey := req.Header.Get("X-Ourd-API-Key"); apiKey != "" {
		p.Data["api_key"] = apiKey
	}
	if accessToken := req.Header.Get("X-Ourd-Access-Token"); accessToken != "" {
		p.Data["access_token"] = accessToken
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

	return
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
