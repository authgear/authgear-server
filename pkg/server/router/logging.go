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
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"mime"
	"net"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/server/logging"
)

type responseLogger struct {
	w      http.ResponseWriter
	status int
	size   int
	b      bytes.Buffer
}

func (l *responseLogger) Header() http.Header {
	return l.w.Header()
}

func (l *responseLogger) Write(b []byte) (int, error) {
	if l.status == 0 {
		// The status will be StatusOK if WriteHeader has not been called yet
		l.status = http.StatusOK
	}
	l.b.Write(b)
	size, err := l.w.Write(b)
	l.size += size
	return size, err
}

func (l *responseLogger) WriteHeader(s int) {
	l.w.WriteHeader(s)
	l.status = s
}

func (l *responseLogger) Status() int {
	return l.status
}

func (l *responseLogger) Size() int {
	return l.size
}

func (l *responseLogger) String() string {
	return l.b.String()
}

func (l *responseLogger) Hijack() (c net.Conn, w *bufio.ReadWriter, e error) {
	hijacker := l.w.(http.Hijacker)
	return hijacker.Hijack()
}

type LoggingMiddleware struct {
	Skips       []string
	MimeConcern []string
	Next        http.Handler
	ByteLimit   *int
}

func (l *LoggingMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	skipBody := l.skipBody(r.URL.Path)

	// Log request
	requestFields := logrus.Fields{}

	body, _ := ioutil.ReadAll(r.Body)
	r.Body = ioutil.NopCloser(bytes.NewReader(body))

	requestBodyLength := len(body)
	requestFields["bodyLength"] = requestBodyLength

	var headers []string
	for key, value := range r.Header {
		headers = append(headers, fmt.Sprintf("%s: %v", key, value))
	}
	requestFields["headers"] = headers

	var shouldLogRequestBody = !skipBody &&
		l.isConcernType(r.Header.Get("Content-Type")) &&
		(l.ByteLimit == nil || *l.ByteLimit >= requestBodyLength)
	if shouldLogRequestBody {
		requestFields["body"] = logging.StringValueFormatter(body)
	}

	log.WithFields(requestFields).Debugf("Request %v %v", r.Method, r.RequestURI)

	// Serve request by passing to next middleware or router
	rlogger := &responseLogger{w: w}
	l.Next.ServeHTTP(rlogger, r)

	// Log Response
	responseFields := logrus.Fields{}
	responseBodyLength := rlogger.Size()
	responseFields["bodyLength"] = responseBodyLength
	responseFields["requestID"] = requestID

	var shouldLogResponseBody = !skipBody &&
		l.isConcernType(w.Header().Get("Content-Type")) &&
		(l.ByteLimit == nil || *l.ByteLimit >= responseBodyLength)
	if shouldLogResponseBody {
		responseFields["body"] = logging.StringValueFormatter(rlogger.String())
	}
	log.WithFields(responseFields).Debugf("Response %v %v", r.Method, r.RequestURI)
}

func (l *LoggingMiddleware) skipBody(urlPath string) bool {
	for _, s := range l.Skips {
		if strings.HasPrefix(urlPath, s) {
			return true
		}
	}
	return false
}

func (l *LoggingMiddleware) isConcernType(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	for _, c := range l.MimeConcern {
		if mediaType == c {
			return true
		}
	}
	return false
}
