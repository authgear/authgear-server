package router

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
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
	Skips []string
	Next  http.Handler
}

func (l *LoggingMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Debugf("%v %v", r.Method, r.RequestURI)

	log.Debugln("------ Header: ------")
	for key, value := range r.Header {
		log.Debugf("%s: %v", key, value)
	}

	skipBody := false
	for _, s := range l.Skips {
		if strings.HasPrefix(r.URL.Path, s) {
			skipBody = true
			break
		}
	}
	body, _ := ioutil.ReadAll(r.Body)
	r.Body = ioutil.NopCloser(bytes.NewReader(body))

	log.Debugln("------ Request: ------")
	if !skipBody && (r.Header.Get("Content-Type") == "" || r.Header.Get("Content-Type") == "application/json") {
		log.Debugln(string(body))
	} else {
		log.Debugf("%d bytes of request body", len(body))
	}

	rlogger := &responseLogger{w: w}
	l.Next.ServeHTTP(rlogger, r)

	log.Debugln("------ Response: ------")
	if !skipBody && (w.Header().Get("Content-Type") == "" || w.Header().Get("Content-Type") == "application/json") {
		log.Debugln(rlogger.String())
	} else {
		log.Debugf("%d bytes of response body", len(rlogger.String()))
	}
}
