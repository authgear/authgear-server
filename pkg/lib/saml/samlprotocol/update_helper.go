//go:build updatecrewjamsaml

package samlprotocol

// The intended usage of this file is to replace the file in the repository with the original file.
// So that you can use `git diff` to view the changes.
// This file is guarded by a build tag so it does not run normally.
//
// It is intended you run it with
//  go generate -tags updatecrewjamsaml ./pkg/lib/saml/samlprotocol

//go:generate curl -sSL https://raw.githubusercontent.com/crewjam/saml/refs/tags/v0.5.0/schema.go -o schema.go
//go:generate curl -sSL https://raw.githubusercontent.com/crewjam/saml/refs/tags/v0.5.0/duration.go -o duration.go
//go:generate curl -sSL https://raw.githubusercontent.com/crewjam/saml/refs/tags/v0.5.0/time.go -o time.go
//go:generate curl -sSL https://raw.githubusercontent.com/crewjam/saml/refs/tags/v0.5.0/metadata.go -o metadata.go
