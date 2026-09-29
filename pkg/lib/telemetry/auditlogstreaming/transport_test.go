package auditlogstreaming

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func newTestEntry(marker string) QueuedEntry {
	raw := fmt.Appendf(nil, `{"marker":%q}`, marker)
	return QueuedEntry{
		Event: &event.Event{
			ID:   marker,
			Type: "user.authenticated",
			Context: event.Context{
				Timestamp: 1700000000,
				AppID:     "app",
			},
		},
		Raw: raw,
	}
}

// readLines accepts one connection on l, reads newline-framed lines until
// the peer closes, and sends them on the returned channel.
func readLines(t *testing.T, l net.Listener) <-chan []string {
	t.Helper()
	out := make(chan []string, 1)
	go func() {
		conn, err := l.Accept()
		if err != nil {
			out <- nil
			return
		}
		defer conn.Close()
		var lines []string
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		out <- lines
	}()
	return out
}

func TestSenderImplSend(t *testing.T) {
	Convey("SenderImpl.Send", t, func() {
		Convey("a batch of three entries arrives as three frames on one connection, in order", func() {
			l, err := net.Listen("tcp", "127.0.0.1:0")
			So(err, ShouldBeNil)
			defer l.Close()

			lines := readLines(t, l)

			streamConfig := newTestStreamConfig("collector", false)
			streamConfig.TCP.Address = l.Addr().String()

			sender := &SenderImpl{
				AppID:    "app",
				Hostname: "app.example.com",
				Streams:  []*config.TelemetryAuditLogStreamConfig{streamConfig},
			}
			sender.Send(t.Context(), []QueuedEntry{
				newTestEntry("one"), newTestEntry("two"), newTestEntry("three"),
			})

			got := <-lines
			So(got, ShouldHaveLength, 3)
			So(got[0], ShouldContainSubstring, `"marker":"one"`)
			So(got[1], ShouldContainSubstring, `"marker":"two"`)
			So(got[2], ShouldContainSubstring, `"marker":"three"`)
		})

		Convey("a batch spanning multiple maxEntriesPerWrite chunks still arrives whole and in order", func() {
			l, err := net.Listen("tcp", "127.0.0.1:0")
			So(err, ShouldBeNil)
			defer l.Close()

			lines := readLines(t, l)

			streamConfig := newTestStreamConfig("collector", false)
			streamConfig.TCP.Address = l.Addr().String()

			sender := &SenderImpl{
				AppID:    "app",
				Hostname: "app.example.com",
				Streams:  []*config.TelemetryAuditLogStreamConfig{streamConfig},
			}

			// 2.5x maxEntriesPerWrite, so this batch is split across
			// three separate conn.Write calls, not one.
			n := maxEntriesPerWrite*2 + maxEntriesPerWrite/2
			var entries []QueuedEntry
			for i := range n {
				entries = append(entries, newTestEntry(fmt.Sprintf("entry-%d", i)))
			}
			sender.Send(t.Context(), entries)

			got := <-lines
			So(got, ShouldHaveLength, n)
			for i, line := range got {
				So(line, ShouldContainSubstring, fmt.Sprintf(`"marker":"entry-%d"`, i))
			}
		})

		Convey("two streams both receive the whole batch; a stream whose dial fails does not prevent the other", func() {
			l, err := net.Listen("tcp", "127.0.0.1:0")
			So(err, ShouldBeNil)
			defer l.Close()
			lines := readLines(t, l)

			good := newTestStreamConfig("good", false)
			good.TCP.Address = l.Addr().String()

			bad := newTestStreamConfig("bad", false)
			bad.TCP.Address = "127.0.0.1:1" // nothing listens on port 1

			sender := &SenderImpl{
				AppID:    "app",
				Hostname: "app.example.com",
				Streams:  []*config.TelemetryAuditLogStreamConfig{bad, good},
			}
			sender.Send(t.Context(), []QueuedEntry{newTestEntry("one")})

			got := <-lines
			So(got, ShouldHaveLength, 1)
			So(got[0], ShouldContainSubstring, `"marker":"one"`)
		})

		Convey("TLS with a matching certificate_authority is delivered", func() {
			ca := newTestCA(t, "ca")
			serverLeaf := ca.issueLeaf(t, "127.0.0.1", x509.ExtKeyUsageServerAuth, []net.IP{net.ParseIP("127.0.0.1")}, nil)

			l, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{
				Certificates: []tls.Certificate{serverLeaf.tlsCertificate()},
			})
			So(err, ShouldBeNil)
			defer l.Close()
			lines := readLines(t, l)

			streamConfig := newTestStreamConfig("collector", true)
			streamConfig.TCP.Address = l.Addr().String()
			materials := config.TelemetryAuditLogStreamTLSMaterials{
				{StreamName: "collector", CertificateAuthority: &config.X509Certificate{Pem: ca.pem()}},
			}

			sender := &SenderImpl{
				AppID:    "app",
				Hostname: "app.example.com",
				Streams:  []*config.TelemetryAuditLogStreamConfig{streamConfig},
				TLS:      &materials,
			}
			sender.Send(t.Context(), []QueuedEntry{newTestEntry("one")})

			got := <-lines
			So(got, ShouldHaveLength, 1)
		})

		Convey("TLS with a non-matching certificate_authority is dropped", func() {
			ca := newTestCA(t, "ca")
			serverLeaf := ca.issueLeaf(t, "127.0.0.1", x509.ExtKeyUsageServerAuth, []net.IP{net.ParseIP("127.0.0.1")}, nil)
			wrongCA := newTestCA(t, "wrong-ca")

			l, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{
				Certificates: []tls.Certificate{serverLeaf.tlsCertificate()},
			})
			So(err, ShouldBeNil)
			defer l.Close()
			lines := readLines(t, l)

			streamConfig := newTestStreamConfig("collector", true)
			streamConfig.TCP.Address = l.Addr().String()
			materials := config.TelemetryAuditLogStreamTLSMaterials{
				{StreamName: "collector", CertificateAuthority: &config.X509Certificate{Pem: wrongCA.pem()}},
			}

			sender := &SenderImpl{
				AppID:    "app",
				Hostname: "app.example.com",
				Streams:  []*config.TelemetryAuditLogStreamConfig{streamConfig},
				TLS:      &materials,
			}
			sender.Send(t.Context(), []QueuedEntry{newTestEntry("one")})

			// The handshake fails, so nothing is ever accepted; give the
			// dropped connection attempt a moment to actually fail before
			// closing the listener out from under it.
			time.Sleep(100 * time.Millisecond)
			l.Close()
			got := <-lines
			So(got, ShouldBeEmpty)
		})

		Convey("mTLS: delivered when the client certificate is signed by the listener's client CA", func() {
			serverCA := newTestCA(t, "server-ca")
			serverLeaf := serverCA.issueLeaf(t, "127.0.0.1", x509.ExtKeyUsageServerAuth, []net.IP{net.ParseIP("127.0.0.1")}, nil)
			clientCA := newTestCA(t, "client-ca")
			clientLeaf := clientCA.issueLeaf(t, "client", x509.ExtKeyUsageClientAuth, nil, nil)

			l, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{
				Certificates: []tls.Certificate{serverLeaf.tlsCertificate()},
				ClientAuth:   tls.RequireAndVerifyClientCert,
				ClientCAs:    clientCA.pool(),
			})
			So(err, ShouldBeNil)
			defer l.Close()
			lines := readLines(t, l)

			streamConfig := newTestStreamConfig("collector", true)
			streamConfig.TCP.Address = l.Addr().String()
			materials := config.TelemetryAuditLogStreamTLSMaterials{
				{
					StreamName:           "collector",
					CertificateAuthority: &config.X509Certificate{Pem: serverCA.pem()},
					ClientCertificate: &config.TelemetryAuditLogStreamClientCertificate{
						Certificate: &config.X509Certificate{Pem: clientLeaf.certPEM},
						Key:         clientLeaf.jwk,
					},
				},
			}

			sender := &SenderImpl{
				AppID:    "app",
				Hostname: "app.example.com",
				Streams:  []*config.TelemetryAuditLogStreamConfig{streamConfig},
				TLS:      &materials,
			}
			sender.Send(t.Context(), []QueuedEntry{newTestEntry("one")})

			got := <-lines
			So(got, ShouldHaveLength, 1)
		})

		Convey("mTLS: refused when the client certificate is not signed by the listener's client CA", func() {
			serverCA := newTestCA(t, "server-ca")
			serverLeaf := serverCA.issueLeaf(t, "127.0.0.1", x509.ExtKeyUsageServerAuth, []net.IP{net.ParseIP("127.0.0.1")}, nil)
			clientCA := newTestCA(t, "client-ca")
			untrustedCA := newTestCA(t, "untrusted-ca")
			clientLeaf := untrustedCA.issueLeaf(t, "client", x509.ExtKeyUsageClientAuth, nil, nil)

			l, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{
				Certificates: []tls.Certificate{serverLeaf.tlsCertificate()},
				ClientAuth:   tls.RequireAndVerifyClientCert,
				ClientCAs:    clientCA.pool(),
			})
			So(err, ShouldBeNil)
			defer l.Close()
			lines := readLines(t, l)

			streamConfig := newTestStreamConfig("collector", true)
			streamConfig.TCP.Address = l.Addr().String()
			materials := config.TelemetryAuditLogStreamTLSMaterials{
				{
					StreamName:           "collector",
					CertificateAuthority: &config.X509Certificate{Pem: serverCA.pem()},
					ClientCertificate: &config.TelemetryAuditLogStreamClientCertificate{
						Certificate: &config.X509Certificate{Pem: clientLeaf.certPEM},
						Key:         clientLeaf.jwk,
					},
				},
			}

			sender := &SenderImpl{
				AppID:    "app",
				Hostname: "app.example.com",
				Streams:  []*config.TelemetryAuditLogStreamConfig{streamConfig},
				TLS:      &materials,
			}
			sender.Send(t.Context(), []QueuedEntry{newTestEntry("one")})

			time.Sleep(100 * time.Millisecond)
			l.Close()
			got := <-lines
			So(got, ShouldBeEmpty)
		})

		Convey("a listener that accepts and then stalls: the write deadline fires and the batch is dropped", func() {
			oldWriteTimeout := writeTimeout
			writeTimeout = 200 * time.Millisecond
			defer func() { writeTimeout = oldWriteTimeout }()

			l, err := net.Listen("tcp", "127.0.0.1:0")
			So(err, ShouldBeNil)
			defer l.Close()

			stopAccept := make(chan struct{})
			defer close(stopAccept)
			go func() {
				conn, err := l.Accept()
				if err != nil {
					return
				}
				if tcpConn, ok := conn.(*net.TCPConn); ok {
					_ = tcpConn.SetReadBuffer(1)
				}
				// Never read: the peer's writes must eventually block.
				<-stopAccept
				_ = conn.Close()
			}()

			streamConfig := newTestStreamConfig("collector", false)
			streamConfig.TCP.Address = l.Addr().String()

			sender := &SenderImpl{
				AppID:    "app",
				Hostname: "app.example.com",
				Streams:  []*config.TelemetryAuditLogStreamConfig{streamConfig},
			}

			// A large batch, to exceed the shrunk receive window quickly.
			var entries []QueuedEntry
			padding := make([]byte, 4096)
			for i := range 1024 {
				e := newTestEntry(fmt.Sprintf("entry-%d-%s", i, padding))
				entries = append(entries, e)
			}

			done := make(chan struct{})
			go func() {
				sender.Send(t.Context(), entries)
				close(done)
			}()

			select {
			case <-done:
				// Send returned, which is what must happen -- it must not
				// hang the tick waiting for a connection that never reads.
			case <-time.After(5 * time.Second):
				t.Fatal("Send did not return after the write deadline should have fired")
			}
		})
	})
}
