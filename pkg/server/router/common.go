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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/skygeario/skygear-server/pkg/server/skyerr"
	"github.com/skygeario/skygear-server/pkg/server/skyversion"
)

// commonRouter implements the HandlerFunc interface that is common
// to Router and Gateway.
type commonRouter struct {
	payloadFunc      func(req *http.Request) (p *Payload, err error)
	matchHandlerFunc func(req *http.Request, p *Payload) (h Handler, pp []Processor)
}

func (r *commonRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var (
		httpStatus    = http.StatusOK
		resp          Response
		handler       Handler
		preprocessors []Processor
		payload       *Payload
	)

	version := strings.TrimPrefix(skyversion.Version(), "v")
	w.Header().Set("Server", fmt.Sprintf("Skygear Server/%s", version))

	resp.writer = w

	defer func() {
		if r := recover(); r != nil {
			resp.Err = errorFromRecoveringPanic(r)
			log.WithField("recovered", r).Errorln("panic occurred while handling request")
		}

		writer := resp.Writer()
		if writer == nil {
			// The response is already written.
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		if resp.Err != nil && httpStatus >= 200 && httpStatus <= 299 {
			httpStatus = defaultStatusCode(resp.Err)
		}

		writer.WriteHeader(httpStatus)
		if err := writeEntity(writer, resp); err != nil {
			panic(err)
		}
	}()

	// Create request, response struct and match handler
	var err error
	payload, err = r.payloadFunc(req)
	if err != nil {
		httpStatus = http.StatusBadRequest
		resp.Err = skyerr.NewRequestJSONInvalidErr(err)
		return
	}

	handler, preprocessors = r.matchHandlerFunc(req, payload)
	if handler == nil {
		httpStatus = http.StatusNotFound
		resp.Err = skyerr.NewError(skyerr.UndefinedOperation, "route unmatched")
		return
	}

	httpStatus = r.callHandler(handler, preprocessors, payload, &resp)
}

func (r *commonRouter) callHandler(handler Handler, pp []Processor, payload *Payload, resp *Response) (httpStatus int) {
	httpStatus = http.StatusOK

	defer func() {
		if r := recover(); r != nil {
			log.WithField("recovered", r).Errorln("panic occurred while handling request")

			resp.Err = errorFromRecoveringPanic(r)
			httpStatus = defaultStatusCode(resp.Err)
		}
	}()

	for _, p := range pp {
		httpStatus = p.Preprocess(payload, resp)
		if resp.Err != nil {
			if httpStatus == http.StatusOK {
				httpStatus = defaultStatusCode(resp.Err)
			}
			return
		}
	}

	handler.Handle(payload, resp)
	return httpStatus
}

func writeEntity(w http.ResponseWriter, i interface{}) error {
	if w == nil {
		return errors.New("writer is nil")
	}
	return json.NewEncoder(w).Encode(i)
}
