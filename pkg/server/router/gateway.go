// Copyright 2015-present Oursky Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package router

import (
	"net/http"
	"regexp"
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
	commonRouter
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
	g.commonRouter.payloadFunc = g.newPayload
	g.commonRouter.matchHandlerFunc = g.matchHandler
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
	g.commonRouter.ServeHTTP(w, req)
}

func (g *Gateway) matchHandler(p *Payload) (h Handler, pp []Processor) {
	method := p.Meta["method"].(string)
	if pathRoute, ok := g.methodPaths[method]; ok {
		h = pathRoute.Handler
		pp = pathRoute.Preprocessors
	}
	return
}

func (g *Gateway) newPayload(req *http.Request) (p *Payload, err error) {
	indices := g.ParamMatch.FindAllStringSubmatchIndex(req.URL.Path, -1)
	params := submatchesFromIndices(req.URL.Path, indices)
	log.Debugf("Matched params: %v", params)
	p = &Payload{
		Req:     req,
		Params:  params,
		Meta:    map[string]interface{}{},
		Data:    map[string]interface{}{},
		Context: req.Context(), // always non-nil
	}

	query := req.URL.Query()
	if apiKey := req.Header.Get("X-Skygear-Api-Key"); apiKey != "" {
		p.Data["api_key"] = apiKey
	} else if apiKey := query.Get("api_key"); apiKey != "" {
		p.Data["api_key"] = apiKey
	}
	if accessToken := req.Header.Get("X-Skygear-Access-Token"); accessToken != "" {
		p.Data["access_token"] = accessToken
	} else if accessToken := query.Get("access_token"); accessToken != "" {
		p.Data["access_token"] = accessToken
	}

	p.Meta["path"] = req.URL.Path
	p.Meta["method"] = req.Method

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
