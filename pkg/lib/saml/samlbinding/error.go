package samlbinding

import "github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"

var ErrNoRequest = &samlprotocol.ParseRequestFailedError{
	Reason: "no SAMLRequest provided",
}
