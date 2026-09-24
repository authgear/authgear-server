package auditlogstreaming

import (
	"bytes"
	"context"
	"crypto/tls"
	"log/slog"
	"net"
	"time"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var Logger = slogutil.NewLogger("audit-log-streaming")

// dialTimeout and writeTimeout are var, not const, so tests can shrink them
// rather than waiting out the real duration.
var (
	dialTimeout  = 5 * time.Second
	writeTimeout = 30 * time.Second
)

// maxEntriesPerWrite bounds how many entries are encoded into one buffer
// and one conn.Write call. The batch (up to maxQueueLength entries, see
// queue.go) is still delivered as a whole over one connection -- this
// only splits the encoding/writing into fixed-size chunks, so peak memory
// for a backlogged project's delivery stays bounded to roughly this many
// entries' worth of frames rather than scaling with the whole batch.
const maxEntriesPerWrite = 100

// dial opens the connection for a stream's tcp transport: plaintext TCP,
// or TLS when the stream enables it.
func dial(ctx context.Context, s *ResolvedTCP) (net.Conn, error) {
	dialCtx, cancel := context.WithTimeout(ctx, dialTimeout)
	defer cancel()

	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(dialCtx, "tcp", s.Address)
	if err != nil {
		return nil, err
	}

	if !s.TLSEnabled {
		return conn, nil
	}

	host, _, err := net.SplitHostPort(s.Address)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: host,
		RootCAs:    s.TLSRootCAs,
	}
	if s.TLSClientCertificate != nil {
		tlsConfig.Certificates = []tls.Certificate{*s.TLSClientCertificate}
	}

	tlsConn := tls.Client(conn, tlsConfig)
	if err := tlsConn.HandshakeContext(dialCtx); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return tlsConn, nil
}

// QueuedEntry is one entry drained from the queue, ready for the sender.
//
// Event carries the fields the encoder reads (Type, ID, Context). Its
// Payload is always nil: event.Payload is an interface with methods, which
// encoding/json cannot decode a stored JSON object into, and the encoder
// never reads Payload. Raw is the original bytes the producer enqueued,
// used verbatim as MSG so it is byte-identical to what was marshalled once
// at enqueue time -- reconstructing MSG by re-marshalling Event would
// silently drop the whole payload, since Event.Payload is nil.
type QueuedEntry struct {
	Event *event.Event
	Raw   []byte
}

// Sender delivers one tick's batch of entries for one app to every
// configured stream of that app.
type Sender interface {
	Send(ctx context.Context, entries []QueuedEntry)
}

type SenderImpl struct {
	AppID    string
	Hostname string
	Streams  []*config.TelemetryAuditLogStreamConfig
	TLS      *config.TelemetryAuditLogStreamTLSMaterials

	DatadogCredentials   *config.TelemetryAuditLogStreamDatadogCredentials
	DatadogClientFactory DatadogClientFactory
}

var _ Sender = &SenderImpl{}

// maxConcurrentStreamDeliveries bounds how many of one app's streams are
// delivered to concurrently within one Send call. Delivered sequentially,
// a batch's total delivery time is the SUM of every stream's dial+write
// time, so a single slow or unreachable collector serially delays every
// other stream -- see forEachBounded's doc comment for why that matters
// for the drain lock.
const maxConcurrentStreamDeliveries = 10

// Send delivers entries to every configured stream, concurrently (bounded
// by maxConcurrentStreamDeliveries). Streams are independent: a failure
// of one does not affect the other, per the spec's UC3.
func (s *SenderImpl) Send(ctx context.Context, entries []QueuedEntry) {
	if len(entries) == 0 {
		return
	}

	logger := Logger.GetLogger(ctx)

	forEachBounded(s.Streams, maxConcurrentStreamDeliveries, func(streamConfig *config.TelemetryAuditLogStreamConfig) {
		s.sendToStream(ctx, logger, streamConfig, entries)
	})
}

func (s *SenderImpl) sendToStream(
	ctx context.Context,
	logger slogutil.NamedLogger,
	streamConfig *config.TelemetryAuditLogStreamConfig,
	entries []QueuedEntry,
) {
	resolved, err := resolveStream(s.AppID, streamConfig, s.TLS, s.DatadogCredentials)
	if err != nil {
		logger.WithError(err).Error(ctx, "failed to resolve audit log stream",
			slog.String("app_id", s.AppID),
			slog.String("stream", streamConfig.Name))
		return
	}

	switch streamConfig.Type {
	case config.TelemetryAuditLogStreamTypeSyslog:
		s.sendSyslogTCP(ctx, logger, streamConfig, resolved, entries)
	case config.TelemetryAuditLogStreamTypeDatadog:
		s.sendDatadogHTTP(ctx, logger, resolved, entries)
	default:
		logger.Error(ctx, "unknown audit log stream type",
			slog.String("app_id", s.AppID),
			slog.String("stream", streamConfig.Name),
			slog.String("type", string(streamConfig.Type)))
	}
}

// sendSyslogTCP delivers entries to one syslog/tcp stream.
func (s *SenderImpl) sendSyslogTCP(
	ctx context.Context,
	logger slogutil.NamedLogger,
	streamConfig *config.TelemetryAuditLogStreamConfig,
	resolved *ResolvedStream,
	entries []QueuedEntry,
) {
	conn, err := dial(ctx, resolved.TCP)
	if err != nil {
		logger.WithError(err).Error(ctx, "failed to dial audit log stream",
			slog.String("app_id", s.AppID),
			slog.String("stream", streamConfig.Name))
		return
	}
	defer func() { _ = conn.Close() }()

	// One deadline for the whole batch's transfer, set once before the
	// chunked writes below -- not reset per chunk, which would let a
	// slow connection take far longer than writeTimeout in total for a
	// large batch.
	if err := conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
		logger.WithError(err).Error(ctx, "failed to set write deadline for audit log stream",
			slog.String("app_id", s.AppID),
			slog.String("stream", streamConfig.Name))
		return
	}

	var buf bytes.Buffer
	for start := 0; start < len(entries); start += maxEntriesPerWrite {
		end := min(start+maxEntriesPerWrite, len(entries))

		buf.Reset()
		for _, entry := range entries[start:end] {
			frame := Frame(resolved.Syslog.Framing, EncodeRFC5424(resolved.Syslog, s.Hostname, entry.Event, entry.Raw))
			buf.Write(frame)
		}

		if _, err := conn.Write(buf.Bytes()); err != nil {
			logger.WithError(err).Error(ctx, "failed to write audit log stream",
				slog.String("app_id", s.AppID),
				slog.String("stream", streamConfig.Name))
			return
		}
	}

	logger.Debug(ctx, "delivered audit log batch",
		slog.String("app_id", s.AppID),
		slog.String("stream", streamConfig.Name),
		slog.Int("count", len(entries)))
}
