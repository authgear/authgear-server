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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/skygeario/skygear-server/logging"
	"github.com/skygeario/skygear-server/skyerr"
	"github.com/skygeario/skygear-server/skyversion"
)

var log = logging.LoggerEntry("router")

// pipeline encapsulates a transformation which a request will come throught
// from preprocessors to the actual handler. (and postprocessor later)
type pipeline struct {
	Action        string
	Preprocessors []Processor
	Handler
}

// Router to dispatch HTTP request to respective handler
type Router struct {
	actions struct {
		sync.RWMutex
		m map[string]pipeline
	}
}

// PreprocessorRegistry is holding all preprocessors and their mapping with
// a string name.
type PreprocessorRegistry map[string]Processor

// GetByNames returns a list of registered preprocessors by preprocessor names.
func (r PreprocessorRegistry) GetByNames(names ...string) []Processor {
	preprocessors := make([]Processor, len(names))
	for i, name := range names {
		pp, ok := r[name]
		if !ok {
			log.Fatalf("preprocessor %s is not defined", name)
		}
		preprocessors[i] = pp
	}
	return preprocessors
}

// NewRouter is factory for Router
func NewRouter() *Router {
	return &Router{
		struct {
			sync.RWMutex
			m map[string]pipeline
		}{
			m: map[string]pipeline{},
		},
	}
}

// Map to register action to handle mapping
func (r *Router) Map(action string, handler Handler, preprocessors ...Processor) {
	r.actions.Lock()
	defer r.actions.Unlock()
	if len(preprocessors) == 0 {
		preprocessors = handler.GetPreprocessors()
	}
	r.actions.m[action] = pipeline{
		Action:        action,
		Preprocessors: preprocessors,
		Handler:       handler,
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var (
		httpStatus    = http.StatusOK
		resp          Response
		handler       Handler
		preprocessors []Processor
		payload       *Payload
	)

	version := strings.TrimPrefix(skyversion.Version(), "v")

	resp.writer = w
	defer func() {
		if r := recover(); r != nil {
			resp.Err = errorFromRecoveringPanic(r)
			log.WithField("recovered", r).Errorln("panic occurred while handling request")
		}

		resp.Header().Set("Server", fmt.Sprintf("Skygear Server/%s", version))

		if !resp.written {
			resp.Header().Set("Content-Type", "application/json")
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

	var err error
	payload, err = newPayloadForJSONHandler(req)
	if err != nil {
		httpStatus = http.StatusBadRequest
		resp.Err = skyerr.NewRequestJSONInvalidErr(err)
		return
	}

	handler, preprocessors = r.matchHandler(req, payload)

	if handler == nil {
		httpStatus = http.StatusNotFound
		resp.Err = skyerr.NewError(skyerr.UndefinedOperation, "route unmatched")
	} else {
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
}

func (r *Router) matchHandler(req *http.Request, p *Payload) (h Handler, pp []Processor) {
	r.actions.RLock()
	defer r.actions.RUnlock()

	// matching using URL
	action := req.URL.Path
	if strings.HasPrefix(action, "/") {
		action = action[1:]
	}

	action = strings.Replace(action, "/", ":", -1)
	if len(action) > 0 { // prevent matching HomeHandler
		if pipeline, ok := r.actions.m[action]; ok {
			h = pipeline.Handler
			pp = pipeline.Preprocessors
		}
	}

	// matching using payload if needed
	if h == nil {
		if pipeline, ok := r.actions.m[p.RouteAction()]; ok {
			h = pipeline.Handler
			pp = pipeline.Preprocessors
		}
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

	if apiKey := req.Header.Get("X-Skygear-Api-Key"); apiKey != "" {
		p.Data["api_key"] = apiKey
	}
	if accessToken := req.Header.Get("X-Skygear-Access-Token"); accessToken != "" {
		p.Data["access_token"] = accessToken
	}

	return
}
