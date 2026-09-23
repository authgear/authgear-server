package auditlogstreaming

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/x509"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

// Datadog's documented intake limits: at most 1000 logs in the array, and
// 5MB of uncompressed body. A batch is split into as many requests as
// these require. datadogMaxLogBytes is the largest encoded log that can
// be sent at all: one element alone in the array, minus the two brackets
// and the separator fits() charges unconditionally. Deriving it from the
// body limit, rather than writing a second number, is what guarantees
// that any log the sender keeps fits an empty chunk.
//
// All three are var, not const, so tests can shrink them rather than
// constructing megabyte- or thousand-entry-sized batches.
var (
	datadogMaxEntriesPerRequest = 1000
	datadogMaxRequestBytes      = 5 * 1000 * 1000
	datadogMaxLogBytes          = datadogMaxRequestBytes - 3
)

// datadogRequestTimeout bounds one chunk's request; datadogBatchTimeout
// bounds every chunk of one stream's batch together, so that a destination
// that is slow rather than down cannot hold the drain lock (runnable.go,
// WithMutexExpiry) for chunks * datadogRequestTimeout. Both are var so
// tests can shrink them.
var (
	datadogRequestTimeout = 30 * time.Second
	datadogBatchTimeout   = 2 * time.Minute
)

// DatadogClientFactory builds the client for one datadog stream. It is an
// interface so that a test can hand the sender a client pointed at an
// httptest server without minting a certificate authority.
type DatadogClientFactory interface {
	MakeClient(rootCAs *x509.CertPool) *http.Client
}

type SSRFSafeDatadogClientFactory struct {
	AllowNonPublicAddresses bool
	AllowedHosts            []string
}

func (f *SSRFSafeDatadogClientFactory) MakeClient(rootCAs *x509.CertPool) *http.Client {
	return httputil.NewSSRFSafeExternalClient(datadogRequestTimeout, httputil.SSRFSafeExternalClientOptions{
		AllowNonPublicAddresses: f.AllowNonPublicAddresses,
		AllowedHosts:            f.AllowedHosts,
		Sink:                    "telemetry.audit_logs.streams",
		RootCAs:                 rootCAs,
	})
}

// datadogChunkWriter accumulates encoded logs into one gzipped JSON array,
// tracking the uncompressed size so that Datadog's limit -- which is on
// the uncompressed body -- can be respected while only the compressed
// bytes are held.
type datadogChunkWriter struct {
	buf          bytes.Buffer
	gz           *gzip.Writer
	count        int
	uncompressed int
}

func newDatadogChunkWriter() *datadogChunkWriter {
	w := &datadogChunkWriter{}
	w.gz = gzip.NewWriter(&w.buf)
	_, _ = w.gz.Write([]byte("["))
	w.uncompressed = 1
	return w
}

// fits reports whether log can be added without exceeding either limit.
// The two extra bytes are the comma that would separate log from the
// previous element and the closing "]". The comma is charged even for the
// first element, where there is none, so the bound is one byte
// conservative -- cheaper than branching in both fits and add, for a byte
// out of five million.
func (w *datadogChunkWriter) fits(log []byte) bool {
	return w.count < datadogMaxEntriesPerRequest &&
		w.uncompressed+1+len(log)+1 <= datadogMaxRequestBytes
}

func (w *datadogChunkWriter) add(log []byte) error {
	if w.count > 0 {
		if _, err := w.gz.Write([]byte(",")); err != nil {
			return err
		}
		w.uncompressed++
	}
	if _, err := w.gz.Write(log); err != nil {
		return err
	}
	w.uncompressed += len(log)
	w.count++
	return nil
}

func (w *datadogChunkWriter) finish() ([]byte, error) {
	if _, err := w.gz.Write([]byte("]")); err != nil {
		return nil, err
	}
	// Close flushes gzip's own buffer; buf is short of the trailer until
	// it returns, so buf.Bytes() is only valid afterwards.
	if err := w.gz.Close(); err != nil {
		return nil, err
	}
	return w.buf.Bytes(), nil
}

// sendDatadogHTTP encodes and delivers one datadog stream's batch,
// chunked to respect Datadog's per-request limits. Entries are encoded
// one at a time into the chunk under construction rather than all up
// front, so peak memory for a large batch stays bounded to one chunk plus
// one log, the same bound the tcp path takes with maxEntriesPerWrite.
func (s *SenderImpl) sendDatadogHTTP(
	ctx context.Context,
	logger slogutil.NamedLogger,
	resolved *ResolvedStream,
	entries []QueuedEntry,
) {
	ctx, cancel := context.WithTimeout(ctx, datadogBatchTimeout)
	defer cancel()

	client := s.DatadogClientFactory.MakeClient(resolved.HTTP.RootCAs)

	chunks := 0
	w := newDatadogChunkWriter()
	for _, entry := range entries {
		log, err := EncodeDatadogLog(resolved.Datadog, s.Hostname, entry)
		if err != nil {
			logger.WithError(err).Warn(ctx, "failed to encode audit log entry for datadog",
				slog.String("app_id", s.AppID),
				slog.String("stream", resolved.Name),
				slog.String("event_id", entry.Event.ID))
			continue
		}

		// The spec's "a single entry that alone exceeds the body limit is
		// dropped", and it cannot be delivered by any splitting. This
		// threshold must stay derived from the same arithmetic fits uses;
		// a log that passes here and then fails fits on a fresh chunk
		// would be written into an over-limit request anyway, since add
		// below is unconditional.
		if len(log) > datadogMaxLogBytes {
			logger.Warn(ctx, "dropping oversized audit log entry for datadog",
				slog.String("app_id", s.AppID),
				slog.String("stream", resolved.Name),
				slog.String("event_id", entry.Event.ID),
				slog.Int("size", len(log)))
			continue
		}

		if !w.fits(log) {
			// Unreachable while the size check above holds -- fits is
			// always true for a kept log on a fresh writer -- kept anyway
			// so a later change to either threshold degrades into a
			// wasted request rather than a stream of empty arrays.
			if w.count > 0 {
				body, err := w.finish()
				if err != nil {
					logger.WithError(err).Error(ctx, "failed to finish audit log chunk for datadog",
						slog.String("app_id", s.AppID),
						slog.String("stream", resolved.Name))
					return
				}
				chunks++
				if !s.postChunk(ctx, logger, client, resolved, body, w.count) {
					return
				}
			}
			w = newDatadogChunkWriter()
		}

		if err := w.add(log); err != nil {
			logger.WithError(err).Error(ctx, "failed to add audit log entry to datadog chunk",
				slog.String("app_id", s.AppID),
				slog.String("stream", resolved.Name))
			return
		}
	}

	if w.count > 0 {
		body, err := w.finish()
		if err != nil {
			logger.WithError(err).Error(ctx, "failed to finish audit log chunk for datadog",
				slog.String("app_id", s.AppID),
				slog.String("stream", resolved.Name))
			return
		}
		chunks++
		if !s.postChunk(ctx, logger, client, resolved, body, w.count) {
			return
		}
	}

	logger.Debug(ctx, "delivered audit log batch",
		slog.String("app_id", s.AppID),
		slog.String("stream", resolved.Name),
		slog.Int("count", len(entries)),
		slog.Int("chunks", chunks))
}

// postChunk sends one chunk and reports whether the batch should continue.
func (s *SenderImpl) postChunk(
	ctx context.Context,
	logger slogutil.NamedLogger,
	client *http.Client,
	resolved *ResolvedStream,
	body []byte,
	count int,
) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resolved.HTTP.Endpoint, bytes.NewReader(body))
	if err != nil {
		logger.WithError(err).Error(ctx, "failed to build audit log request for datadog",
			slog.String("app_id", s.AppID),
			slog.String("stream", resolved.Name))
		return true
	}
	req.Header.Set("DD-API-KEY", resolved.Datadog.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	resp, err := client.Do(req)
	if err != nil {
		// A refusal by the fetch address policy, a TLS failure, a
		// timeout: this chunk is dropped, the remaining chunks are still
		// sent, per the spec.
		logger.WithError(err).Error(ctx, "failed to deliver audit log chunk to datadog",
			slog.String("app_id", s.AppID),
			slog.String("stream", resolved.Name))
		return true
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	switch {
	case resp.StatusCode == http.StatusAccepted:
		logger.Debug(ctx, "delivered audit log chunk to datadog",
			slog.String("app_id", s.AppID),
			slog.String("stream", resolved.Name),
			slog.Int("count", count))
		return true
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		logger.Error(ctx, "datadog api key is missing, wrong or revoked; abandoning the rest of this batch",
			slog.String("app_id", s.AppID),
			slog.String("stream", resolved.Name),
			slog.Int("status", resp.StatusCode),
			slog.String("body", string(snippet)))
		return false
	default:
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		logger.Error(ctx, "unexpected datadog response status; dropping this chunk",
			slog.String("app_id", s.AppID),
			slog.String("stream", resolved.Name),
			slog.Int("status", resp.StatusCode),
			slog.String("body", string(snippet)))
		return true
	}
}
