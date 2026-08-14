package httputil

import "net/url"

// SensitiveQueryParams lists query parameter names whose values must never
// appear verbatim in logs (e.g. request URLs), because they can carry
// identity tokens, PII, or one-time credentials.
var SensitiveQueryParams = []string{
	// OIDC end_session / authorization request hints.
	"id_token_hint",
	"login_hint",
	// Verification / password reset / tester callback links.
	"code",
	"token",
}

const RedactedQueryParamValue = "redacted"

// RedactedRawQuery returns rawQuery with the values of SensitiveQueryParams
// replaced, so it is safe to write to logs. It returns rawQuery unchanged if
// none of SensitiveQueryParams are present.
func RedactedRawQuery(rawQuery string) string {
	query, err := url.ParseQuery(rawQuery)
	if err != nil {
		return rawQuery
	}

	redacted := false
	for _, key := range SensitiveQueryParams {
		if query.Has(key) {
			query.Set(key, RedactedQueryParamValue)
			redacted = true
		}
	}

	if !redacted {
		return rawQuery
	}

	return query.Encode()
}
