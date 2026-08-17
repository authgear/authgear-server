package oauthclient

import (
	"strings"

	"github.com/authgear/authgear-server/pkg/util/rand"
)

const (
	DCRClientIDPrefix = "dcrc_"

	dcrClientIDAlphabet     = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	dcrClientIDRandomLength = 22 // matches spec: "22 chars URL-safe base64 (16 bytes)"
)

// GenerateDCRClientID returns a new dcrc_-prefixed client_id for a
// DCR-registered client.
func GenerateDCRClientID() string {
	return DCRClientIDPrefix + rand.StringWithAlphabet(dcrClientIDRandomLength, dcrClientIDAlphabet, rand.SecureRand)
}

// IsDCRClientID reports whether clientID has the shape of a DCR-registered
// client_id, as opposed to a statically configured or CIMD client_id.
func IsDCRClientID(clientID string) bool {
	return strings.HasPrefix(clientID, DCRClientIDPrefix)
}
