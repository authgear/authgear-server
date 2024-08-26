package samlbinding

import "github.com/authgear/authgear-server/pkg/lib/saml/samlerror"

var ErrNoRequest = &samlerror.ParseRequestFailedError{
	Reason: "no SAMLRequest provided",
}
