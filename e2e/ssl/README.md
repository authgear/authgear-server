# e2e TLS fixtures

`ca.crt` / `ca.key` are used by `cmd/proxy` and predate this file.

`syslog-ca.crt`, `syslog-server.crt` and `syslog-server.key` are used by
`cmd/syslogserver` for the TLS listener in
`e2e/tests/telemetry/audit_log_streaming/tls.test.yaml`. The existing
`ca.crt` cannot serve this purpose: it carries no `subjectAltName`, so Go's
TLS client refuses to verify it against `127.0.0.1`.

`syslog-ca.key` is deliberately **not** committed. It is used only to sign
`syslog-server.crt` below and is discarded afterwards; nothing in the suite
needs to re-sign anything at runtime.

Both certificates are valid for 100 years, so the suite does not rot.
Regenerate with:

```sh
openssl req -x509 -newkey rsa:2048 -nodes -keyout syslog-ca.key -out syslog-ca.crt \
  -days 36500 -subj "/CN=authgear-e2e-syslog-ca"

openssl req -newkey rsa:2048 -nodes -keyout syslog-server.key -out syslog-server.csr \
  -subj "/CN=authgear-e2e-syslog"

openssl x509 -req -in syslog-server.csr -CA syslog-ca.crt -CAkey syslog-ca.key \
  -CAcreateserial -out syslog-server.crt -days 36500 \
  -extfile <(printf "subjectAltName=IP:127.0.0.1,DNS:localhost")

rm syslog-server.csr syslog-ca.srl syslog-ca.key
```

No client certificate is issued here: mTLS is covered by unit tests in
`pkg/lib/telemetry/auditlogstreaming` (`transport_test.go`), not by e2e.
