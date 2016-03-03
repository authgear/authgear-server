package router

import (
	"net/http"
	"regexp"

	log "github.com/Sirupsen/logrus"
)

// pathRoute is the path matching version of pipeline. Instead of storing the action
// to match, it stores a regexp to match against request URL.
type pathRoute struct {
	Preprocessors []Processor
	Handler
}

// Gateway is a man in the middle to inject dependency
// It currently bind to HTTP method, it disregard path.
type Gateway struct {
	ParamMatch  *regexp.Regexp
	methodPaths map[string]pathRoute
}

func NewGateway(pattern string, path string, mux *http.ServeMux) *Gateway {
	match := regexp.MustCompile(`\A/` + pattern + `\z`)
	g := &Gateway{
		ParamMatch:  match,
		methodPaths: map[string]pathRoute{},
	}
	if path != "" && mux != nil {
		mux.Handle(path, g)
	}
	return g
}

// GET register a URL handler by method GET
func (g *Gateway) GET(handler Handler, preprocessors ...Processor) {
	g.Handle("GET", handler, preprocessors...)
}

// POST register a URL handler by method POST
func (g *Gateway) POST(handler Handler, preprocessors ...Processor) {
	g.Handle("POST", handler, preprocessors...)
}

// PUT register a URL handler by method PUT
func (g *Gateway) PUT(handler Handler, preprocessors ...Processor) {
	g.Handle("PUT", handler, preprocessors...)
}

// Handle registers a handler matched by a request's method and URL's path.
// Pattern is a regexp that defines a matched URL.
func (g *Gateway) Handle(method string, handler Handler, preprocessors ...Processor) {
	if len(preprocessors) == 0 {
		preprocessors = handler.GetPreprocessors()
	}
	g.methodPaths[method] = pathRoute{
		Preprocessors: preprocessors,
		Handler:       handler,
	}
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var (
		httpStatus    = http.StatusOK
		resp          Response
		handler       Handler
		preprocessors []Processor
		payload       *Payload
	)
	resp.writer = w
	defer func() {
		if r := recover(); r != nil {
			resp.Err = errorFromRecoveringPanic(r)
			log.WithField("recovered", r).Errorln("panic occurred while handling request")
		}

		if !resp.written && !resp.hijacked {
			if resp.Err != nil && httpStatus >= 200 && httpStatus <= 299 {
				resp.writer.WriteHeader(defaultStatusCode(resp.Err))
			} else {
				resp.writer.WriteHeader(httpStatus)
			}
			if err := resp.WriteEntity(resp); err != nil {
				panic(err)
			}
		}
	}()

	handler, preprocessors, payload = g.matchRawHandler(req)
	if handler == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	for _, p := range preprocessors {
		httpStatus = p.Preprocess(payload, &resp)
		if resp.Err != nil {
			if httpStatus == http.StatusOK {
				httpStatus = defaultStatusCode(resp.Err)
			}
			return
		}
	}
	handler.Handle(payload, &resp)
}

func (g *Gateway) matchRawHandler(req *http.Request) (h Handler, pp []Processor, p *Payload) {
	if pathRoute, ok := g.methodPaths[req.Method]; ok {
		h = pathRoute.Handler
		pp = pathRoute.Preprocessors
		p = g.newPayloadForRawHandler(req)
	}
	return
}

func (g *Gateway) newPayloadForRawHandler(req *http.Request) (p *Payload) {
	indices := g.ParamMatch.FindAllStringSubmatchIndex(req.URL.Path, -1)
	params := submatchesFromIndices(req.URL.Path, indices)
	log.Debugf("Matched params: %v", params)
	p = &Payload{
		Req:    req,
		Params: params,
		Meta:   map[string]interface{}{},
		Data:   map[string]interface{}{},
	}

	if apiKey := req.Header.Get("X-Skygear-Api-Key"); apiKey != "" {
		p.Data["api_key"] = apiKey
	} else if apiKey := req.FormValue("api_key"); apiKey != "" {
		p.Data["api_key"] = apiKey
	}
	if accessToken := req.Header.Get("X-Skygear-Access-Token"); accessToken != "" {
		p.Data["access_token"] = accessToken
	} else if accessToken := req.FormValue("access_token"); accessToken != "" {
		p.Data["access_token"] = accessToken
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
