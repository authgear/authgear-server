package samlprotocol

import (
	"fmt"
	"time"

	"github.com/beevik/etree"
	crewjamsaml "github.com/crewjam/saml"
)

type Response struct {
	crewjamsaml.Response
}

// Copied from https://github.com/crewjam/saml/blob/193e551d9a8420216fae88c2b8f4b46696b7bb63/schema.go#L502
func (r *Response) Element() *etree.Element {
	el := etree.NewElement("samlp:Response")
	el.CreateAttr("xmlns:saml", "urn:oasis:names:tc:SAML:2.0:assertion")
	el.CreateAttr("xmlns:samlp", "urn:oasis:names:tc:SAML:2.0:protocol")

	// Note: This namespace is not used by any element or attribute name, but
	// is required so that the AttributeValue type element can have a value like
	// "xs:string". If we don't declare it here, then it will be stripped by the
	// cannonicalizer. This could be avoided by providing a prefix list to the
	// cannonicalizer, but prefix lists do not appear to be implemented correctly
	// in some libraries, so the safest action is to always produce XML that is
	// (a) in canonical form and (b) does not require prefix lists.
	el.CreateAttr(fmt.Sprintf("xmlns:%s", xmlSchemaNamespace), "http://www.w3.org/2001/XMLSchema")

	el.CreateAttr("ID", r.ID)
	if r.InResponseTo != "" {
		el.CreateAttr("InResponseTo", r.InResponseTo)
	}
	el.CreateAttr("Version", r.Version)
	el.CreateAttr("IssueInstant", r.IssueInstant.Format(timeFormat))
	if r.Destination != "" {
		el.CreateAttr("Destination", r.Destination)
	}
	if r.Consent != "" {
		el.CreateAttr("Consent", r.Consent)
	}
	if r.Issuer != nil {
		el.AddChild(r.Issuer.Element())
	}
	if r.Signature != nil {
		el.AddChild(r.Signature)
	}
	el.AddChild(r.Status.Element())
	if r.EncryptedAssertion != nil {
		el.AddChild(r.EncryptedAssertion)
	}
	if r.Assertion != nil {
		el.AddChild(r.Assertion.Element())
	}
	return el
}

func newResponse(issueInstant time.Time, status Status, issuer string) *Response {
	return &Response{
		crewjamsaml.Response{
			ID:           GenerateResponseID(),
			IssueInstant: issueInstant,
			Version:      SAMLVersion2,
			Status:       status.status,
			Issuer: &crewjamsaml.Issuer{
				Format: SAMLIssertFormatEntity,
				Value:  issuer,
			},
		},
	}
}

func NewRequestDeniedErrorResponse(
	issueInstant time.Time,
	issuer string,
	message string,
	details []*etree.Element) *Response {
	var messageEl *crewjamsaml.StatusMessage
	if message != "" {
		messageEl = &crewjamsaml.StatusMessage{
			Value: message,
		}
	}
	status := crewjamsaml.Status{
		StatusCode: crewjamsaml.StatusCode{
			Value: crewjamsaml.StatusRequester,
			StatusCode: &crewjamsaml.StatusCode{
				Value: crewjamsaml.StatusRequestDenied,
			},
		},
		StatusMessage: messageEl,
	}
	if len(details) > 0 {
		status.StatusDetail = &crewjamsaml.StatusDetail{
			Children: details,
		}
	}
	return newResponse(issueInstant, Status{status: status}, issuer)
}

func NewNoPassiveErrorResponse(
	issueInstant time.Time,
	issuer string,
) *Response {
	status := crewjamsaml.Status{
		StatusCode: crewjamsaml.StatusCode{
			Value: crewjamsaml.StatusRequester,
			StatusCode: &crewjamsaml.StatusCode{
				Value: crewjamsaml.StatusNoPassive,
			},
		},
	}
	return newResponse(issueInstant, Status{status: status}, issuer)
}

func NewServerErrorResponse(
	issueInstant time.Time,
	issuer string,
	message string,
	details []*etree.Element) *Response {
	var messageEl *crewjamsaml.StatusMessage
	if message != "" {
		messageEl = &crewjamsaml.StatusMessage{
			Value: message,
		}
	}
	status := crewjamsaml.Status{
		StatusCode: crewjamsaml.StatusCode{
			Value: crewjamsaml.StatusResponder,
		},
		StatusMessage: messageEl,
	}
	if len(details) > 0 {
		status.StatusDetail = &crewjamsaml.StatusDetail{
			Children: details,
		}
	}
	return newResponse(issueInstant, Status{status: status}, issuer)
}

func NewUnexpectedServerErrorResponse(issueInstant time.Time, issuer string) *Response {
	return NewServerErrorResponse(issueInstant, issuer, "unexpected error", nil)
}

type Status struct {
	status crewjamsaml.Status
}

func NewSuccessResponse(
	issueInstant time.Time,
	issuer string,
	inResponseTo string) *Response {
	status := crewjamsaml.Status{
		StatusCode: crewjamsaml.StatusCode{
			Value: crewjamsaml.StatusSuccess,
		},
	}
	response := newResponse(issueInstant, Status{status: status}, issuer)
	response.InResponseTo = inResponseTo
	return response
}
