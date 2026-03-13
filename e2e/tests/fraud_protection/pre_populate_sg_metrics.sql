{{- range $i, $v := until 120 }}
INSERT INTO _audit_metrics (id, app_id, name, key, created_at)
VALUES ('{{ uuidv4 }}', '{{ $.AppID }}', 'sms_otp_verified', 'phone_country:SG', NOW() - INTERVAL '2 days');
{{- end }}
