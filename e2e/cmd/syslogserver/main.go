// Command syslogserver is the e2e recorder for audit log streaming.
//
// It runs a plaintext TCP listener (127.0.0.1:5140), a TLS listener
// (127.0.0.1:6514), and an HTTP query API (127.0.0.1:5141) that tests use
// to assert on what arrived. See
// docs/plans/audit-log-streaming/2026-09-18-03-e2e.md.
package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// message is one decoded syslog frame, in the shape the e2e YAML tests
// assert against.
type message struct {
	Framing          string            `json:"framing,omitempty"`
	Port             int               `json:"port"`
	PRI              int               `json:"pri"`
	Facility         int               `json:"facility"`
	Severity         int               `json:"severity"`
	Version          int               `json:"version"`
	Timestamp        string            `json:"timestamp"`
	Hostname         string            `json:"hostname"`
	AppName          string            `json:"app_name"`
	ProcID           string            `json:"procid"`
	MsgID            string            `json:"msgid"`
	StructuredDataID string            `json:"structured_data_id,omitempty"`
	StructuredData   map[string]string `json:"structured_data,omitempty"`
	Message          string            `json:"message,omitempty"`
	// ParseError is set, and every other field left at its zero value,
	// when the frame could not be decoded. A test then fails loudly on
	// the parse_error rather than seeing what looks like a missing
	// message.
	ParseError string `json:"parse_error,omitempty"`
}

type recorder struct {
	mu       sync.Mutex
	messages []message
}

func (r *recorder) append(m message) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.messages = append(r.messages, m)
}

// matching returns every recorded message for port, and for appID when
// appID is non-empty, in the order they arrived, save that entries are
// grouped by framing (stably, so arrival order survives within a group).
// Two streams configured on the same port race to deliver the same
// underlying event, so their relative order is not meaningful; grouping
// by framing makes that race deterministic for tests without hiding a
// real per-stream ordering regression, since arrival order is preserved
// inside each group.
func (r *recorder) matching(appID string, port int) []message {
	r.mu.Lock()
	defer r.mu.Unlock()
	// A nil slice marshals to JSON null, which the e2e matcher cannot
	// compare against a "[]" schema, so start non-nil.
	out := []message{}
	for _, m := range r.messages {
		if m.Port != port {
			continue
		}
		if appID != "" && m.StructuredData["app_id"] != appID {
			continue
		}
		out = append(out, m)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Framing < out[j].Framing })
	return out
}

// query polls matching every 50ms until minCount messages have arrived or
// timeout elapses, then returns whatever matched. Delivery is
// asynchronous and batched, so without this every test would need a sleep
// step tuned by guesswork.
func (r *recorder) query(appID string, port int, minCount int, timeout time.Duration) []message {
	deadline := time.Now().Add(timeout)
	for {
		matched := r.matching(appID, port)
		if len(matched) >= minCount {
			return matched
		}
		if time.Now().After(deadline) {
			return matched
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// readFrame reads one RFC 6587 frame from r, auto-detecting the framing:
// a leading ASCII digit means octet counting, anything else means
// non-transparent (newline) framing.
func readFrame(r *bufio.Reader) (frame []byte, framing string, err error) {
	b, err := r.Peek(1)
	if err != nil {
		return nil, "", err
	}

	if b[0] >= '0' && b[0] <= '9' {
		lenStr, err := r.ReadString(' ')
		if err != nil {
			return nil, "", err
		}
		lenStr = strings.TrimSuffix(lenStr, " ")
		n, err := strconv.Atoi(lenStr)
		if err != nil {
			return nil, "", fmt.Errorf("invalid octet count %q: %w", lenStr, err)
		}
		buf := make([]byte, n)
		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, "", err
		}
		return buf, "octet_counting", nil
	}

	line, err := r.ReadBytes('\n')
	if err != nil {
		return nil, "", err
	}
	return bytes.TrimSuffix(line, []byte("\n")), "newline", nil
}

// parseMessage decodes one RFC 5424 message:
//
//	<PRI>VERSION TIMESTAMP HOSTNAME APP-NAME PROCID MSGID [SD] MSG
func parseMessage(frame []byte, framing string, port int) message {
	m := message{Framing: framing, Port: port}
	s := string(frame)

	if !strings.HasPrefix(s, "<") {
		m.ParseError = "missing PRI: no leading '<'"
		return m
	}
	priEnd := strings.IndexByte(s, '>')
	if priEnd < 0 {
		m.ParseError = "unterminated PRI: no '>'"
		return m
	}
	pri, err := strconv.Atoi(s[1:priEnd])
	if err != nil {
		m.ParseError = "invalid PRI: " + err.Error()
		return m
	}
	m.PRI = pri
	m.Facility = pri / 8
	m.Severity = pri % 8

	rest := s[priEnd+1:]

	var fields [6]string
	for i := range fields {
		idx := strings.IndexByte(rest, ' ')
		if idx < 0 {
			m.ParseError = "truncated header"
			return m
		}
		fields[i] = rest[:idx]
		rest = rest[idx+1:]
	}
	version, err := strconv.Atoi(fields[0])
	if err != nil {
		m.ParseError = "invalid VERSION: " + err.Error()
		return m
	}
	m.Version = version
	m.Timestamp = fields[1]
	m.Hostname = fields[2]
	m.AppName = fields[3]
	m.ProcID = fields[4]
	m.MsgID = fields[5]

	if !strings.HasPrefix(rest, "[") {
		m.ParseError = "missing structured data: no leading '['"
		return m
	}
	sdEnd, err := findStructuredDataEnd(rest)
	if err != nil {
		m.ParseError = err.Error()
		return m
	}
	id, params, err := parseStructuredData(rest[1:sdEnd])
	if err != nil {
		m.ParseError = err.Error()
		return m
	}
	m.StructuredDataID = id
	m.StructuredData = params

	m.Message = strings.TrimPrefix(rest[sdEnd+1:], " ")

	return m
}

// findStructuredDataEnd returns the index in s (which starts with '[') of
// the ']' that closes the structured data element, skipping any '\]'
// escaped inside a quoted parameter value.
func findStructuredDataEnd(s string) (int, error) {
	inQuotes := false
	for i := 1; i < len(s); i++ {
		switch {
		case s[i] == '\\' && inQuotes:
			i++
		case s[i] == '"':
			inQuotes = !inQuotes
		case s[i] == ']' && !inQuotes:
			return i, nil
		}
	}
	return 0, fmt.Errorf("unterminated structured data: no closing ']'")
}

// parseStructuredData parses `SD-ID name="value" name="value" ...`
// (without the enclosing brackets), unescaping \\, \" and \] in values.
func parseStructuredData(s string) (string, map[string]string, error) {
	id, rest, _ := strings.Cut(s, " ")
	params := map[string]string{}

	for {
		rest = strings.TrimPrefix(rest, " ")
		if rest == "" {
			break
		}
		name, tail, ok := strings.Cut(rest, "=")
		if !ok {
			return "", nil, fmt.Errorf("invalid structured data parameter: missing '='")
		}
		if !strings.HasPrefix(tail, `"`) {
			return "", nil, fmt.Errorf("invalid structured data parameter %q: expected quoted value", name)
		}
		tail = tail[1:]

		var value strings.Builder
		i := 0
		closed := false
		for i < len(tail) {
			switch {
			case tail[i] == '\\' && i+1 < len(tail):
				value.WriteByte(tail[i+1])
				i += 2
			case tail[i] == '"':
				i++
				closed = true
			default:
				value.WriteByte(tail[i])
				i++
			}
			if closed {
				break
			}
		}
		if !closed {
			return "", nil, fmt.Errorf("invalid structured data parameter %q: unterminated value", name)
		}
		params[name] = value.String()
		rest = tail[i:]
	}

	return id, params, nil
}

func acceptLoop(l net.Listener, port int, rec *recorder) {
	for {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		go handleConn(conn, port, rec)
	}
}

func handleConn(conn net.Conn, port int, rec *recorder) {
	defer func() { _ = conn.Close() }()
	reader := bufio.NewReader(conn)
	for {
		frame, framing, err := readFrame(reader)
		if err != nil {
			return
		}
		rec.append(parseMessage(frame, framing, port))
	}
}

func trapSIGQUIT() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGQUIT)
	go func() {
		for range c {
			buf := make([]byte, 1024)
			for {
				n := runtime.Stack(buf, true)
				if n < len(buf) {
					_, _ = os.Stderr.Write(buf[:n])
					break
				}
				buf = make([]byte, 2*len(buf))
			}
		}
	}()
}

func main() {
	trapSIGQUIT()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	rec := &recorder{}

	plainListener, err := net.Listen("tcp", "127.0.0.1:5140")
	if err != nil {
		log.Fatalf("failed to listen on 127.0.0.1:5140: %v", err)
	}
	go acceptLoop(plainListener, 5140, rec)

	cert, err := tls.LoadX509KeyPair("./ssl/syslog-server.crt", "./ssl/syslog-server.key")
	if err != nil {
		log.Fatalf("failed to load syslog server certificate: %v", err)
	}
	tlsListener, err := tls.Listen("tcp", "127.0.0.1:6514", &tls.Config{
		Certificates: []tls.Certificate{cert},
	})
	if err != nil {
		log.Fatalf("failed to listen on 127.0.0.1:6514: %v", err)
	}
	go acceptLoop(tlsListener, 6514, rec)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		appID := q.Get("app_id")
		port, _ := strconv.Atoi(q.Get("port"))
		minCount, _ := strconv.Atoi(q.Get("min_count"))

		messages := rec.query(appID, port, minCount, 10*time.Second)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"messages": messages})
	})

	server := &http.Server{
		Addr:              "127.0.0.1:5141",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		_ = server.Shutdown(context.Background())
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("failed to start syslog server query API: %v", err)
	}
}
