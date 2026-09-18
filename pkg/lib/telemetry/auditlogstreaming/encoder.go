package auditlogstreaming

import (
	"bytes"
	"strconv"
	"strings"
	"time"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

// ResolvedSyslog is the syslog-encoding-relevant part of a resolved stream.
// It is separate from ResolvedStream (resolve.go) because the encoder does
// not need the stream's transport or TLS material.
type ResolvedSyslog struct {
	Format           config.SyslogFormat
	Framing          config.SyslogFraming
	Facility         int
	AppName          string
	StructuredDataID string
}

// msgID is the constant RFC 5424 MSGID. It does not carry the activity
// type: RFC 5424 caps MSGID at 32 characters, and the longest activity
// type in the repo (admin_api.mutation.unschedule_account_anonymization.executed)
// is 60.
const msgID = "authgear-audit-log"

// severityByType lists the activity types delivered at severity 4
// (warning). Every other activity type is severity 6 (info). The list is
// explicit and a new activity type is info until it is added here --
// severity is a convenience for receiver-side filtering, activity_type in
// the structured data is the authoritative classification of an entry.
var severityByType = map[event.Type]int{
	nonblocking.EmailError:    4,
	nonblocking.SMSError:      4,
	nonblocking.WhatsappError: 4,
}

func severity(t event.Type) int {
	if s, ok := severityByType[t]; ok {
		return s
	}
	return 6
}

// EncodeRFC5424 builds a single RFC 5424 syslog message:
//
//	<PRI>1 TIMESTAMP HOSTNAME APP-NAME - authgear-audit-log [SD] MSG
//
// hostname is the project's RFC 5424 HOSTNAME (see ResolveHostname), msg is
// the JSON-encoded event, byte-identical to what the producer enqueued.
func EncodeRFC5424(s *ResolvedSyslog, hostname string, e *event.Event, msg []byte) []byte {
	var buf bytes.Buffer

	pri := s.Facility*8 + severity(e.Type)
	buf.WriteByte('<')
	buf.WriteString(strconv.Itoa(pri))
	buf.WriteByte('>')
	buf.WriteString("1 ")
	buf.WriteString(time.Unix(e.Context.Timestamp, 0).UTC().Format("2006-01-02T15:04:05Z"))
	buf.WriteByte(' ')
	buf.WriteString(hostname)
	buf.WriteByte(' ')
	buf.WriteString(s.AppName)
	buf.WriteByte(' ')
	buf.WriteString("-") // PROCID
	buf.WriteByte(' ')
	buf.WriteString(msgID)
	buf.WriteByte(' ')

	buf.WriteByte('[')
	buf.WriteString(s.StructuredDataID)
	writeSDParam(&buf, "app_id", e.Context.AppID)
	writeSDParam(&buf, "id", e.ID)
	writeSDParam(&buf, "activity_type", string(e.Type))
	if e.Context.UserID != nil {
		writeSDParam(&buf, "user_id", *e.Context.UserID)
	}
	writeSDParam(&buf, "client_id", e.Context.ClientID)
	writeSDParam(&buf, "ip_address", e.Context.IPAddress)
	buf.WriteByte(']')
	buf.WriteByte(' ')

	buf.Write(msg)

	return buf.Bytes()
}

// writeSDParam writes ` name="value"` to buf, omitting the parameter
// entirely when value is empty.
func writeSDParam(buf *bytes.Buffer, name string, value string) {
	if value == "" {
		return
	}
	buf.WriteByte(' ')
	buf.WriteString(name)
	buf.WriteString(`="`)
	buf.WriteString(escapeSDParam(value))
	buf.WriteByte('"')
}

// escapeSDParam escapes a structured-data parameter value per RFC 5424
// section 6.3.3. Backslash must be escaped first, or the escapes
// introduced by the other two replacements get double-escaped.
func escapeSDParam(v string) string {
	v = strings.ReplaceAll(v, `\`, `\\`)
	v = strings.ReplaceAll(v, `"`, `\"`)
	v = strings.ReplaceAll(v, `]`, `\]`)
	return v
}

// Frame delimits syslogMsg for the wire, per the configured framing.
func Frame(framing config.SyslogFraming, syslogMsg []byte) []byte {
	switch framing {
	case config.SyslogFramingOctetCounting:
		prefix := strconv.Itoa(len(syslogMsg)) + " "
		out := make([]byte, 0, len(prefix)+len(syslogMsg))
		out = append(out, prefix...)
		out = append(out, syslogMsg...)
		return out
	case config.SyslogFramingNewline:
		out := make([]byte, 0, len(syslogMsg)+1)
		out = append(out, syslogMsg...)
		out = append(out, '\n')
		return out
	default:
		panic("auditlogstreaming: unknown framing: " + string(framing))
	}
}
