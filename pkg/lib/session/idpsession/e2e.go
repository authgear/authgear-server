package idpsession

// Only for e2e use. Do not use it in other packages.
func E2EEncodeToken(idpSessionID string, token string) string {
	return encodeToken(idpSessionID, token)
}

// Only for e2e use. Do not use it in other packages.
func E2EHashToken(token string) string {
	return hashToken(token)
}
