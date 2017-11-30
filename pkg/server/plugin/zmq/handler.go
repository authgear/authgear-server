// Copyright 2017-present Oursky Ltd.
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

package zmq

import (
	"net/http"
)

type zmqResponseWriter struct {
	header   int
	response []byte
}

func (w *zmqResponseWriter) Header() http.Header {
	return map[string][]string{}
}
func (w *zmqResponseWriter) Write(body []byte) (int, error) {
	w.response = append(w.response, body...)
	return len(body), nil
}
func (w *zmqResponseWriter) WriteHeader(status int) {
	w.header = status
}
