package oauth

// Only for e2e use. Do not use it in other packages.
func E2EEncodeRefreshToken(grantID string, token string) string {
	return EncodeRefreshToken(grantID, token)
}

// Only for e2e use. Do not use it in other packages.
func E2EHashToken(token string) string {
	return HashToken(token)
}
