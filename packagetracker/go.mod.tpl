module github.com/authgear/authgear-server/packagetracker

// go1.21 supports toolchain
// See https://go.dev/doc/toolchain
go 1.26.10

require (
	// We copy some files from crewjam/saml into pkg/lib/saml.
	// It is not necessary to keep them in sync with upstream.
	// But we must keep watching new releases for security fixes that
	// authgear needs, and apply them to our copies by hand.
	// This version is the latest release we have reviewed.
	github.com/crewjam/saml v0.6.3
)
