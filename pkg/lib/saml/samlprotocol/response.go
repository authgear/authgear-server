package samlprotocol

import (
	"time"

	"github.com/beevik/etree"
	crewjamsaml "github.com/crewjam/saml"
)

func newResponse(issueInstant time.Time, status Status, issuer string) *Response {
	return &Response{
		ID:           GenerateResponseID(),
		IssueInstant: issueInstant,
		Version:      SAMLVersion2,
		Status:       status,
		Issuer: &Issuer{
			Format: SAMLIssertFormatEntity,
			Value:  issuer,
		},
	}
}

func NewRequestDeniedErrorResponse(
	issueInstant time.Time,
	issuer string,
	message string,
	details []*etree.Element) *Response {
	var messageEl *StatusMessage
	if message != "" {
		messageEl = &StatusMessage{
			Value: message,
		}
	}
	status := Status{
		StatusCode: StatusCode{
			Value: crewjamsaml.StatusRequester,
			StatusCode: &StatusCode{
				Value: crewjamsaml.StatusRequestDenied,
			},
		},
		StatusMessage: messageEl,
	}
	if len(details) > 0 {
		status.StatusDetail = &StatusDetail{
			Children: details,
		}
	}
	return newResponse(issueInstant, status, issuer)
}

func NewNoPassiveErrorResponse(
	issueInstant time.Time,
	issuer string,
) *Response {
	status := Status{
		StatusCode: StatusCode{
			Value: crewjamsaml.StatusRequester,
			StatusCode: &StatusCode{
				Value: crewjamsaml.StatusNoPassive,
			},
		},
	}
	return newResponse(issueInstant, status, issuer)
}

func NewServerErrorResponse(
	issueInstant time.Time,
	issuer string,
	message string,
	details []*etree.Element) *Response {
	var messageEl *StatusMessage
	if message != "" {
		messageEl = &StatusMessage{
			Value: message,
		}
	}
	status := Status{
		StatusCode: StatusCode{
			Value: crewjamsaml.StatusResponder,
		},
		StatusMessage: messageEl,
	}
	if len(details) > 0 {
		status.StatusDetail = &StatusDetail{
			Children: details,
		}
	}
	return newResponse(issueInstant, status, issuer)
}

func NewUnexpectedServerErrorResponse(issueInstant time.Time, issuer string) *Response {
	return NewServerErrorResponse(issueInstant, issuer, "unexpected error", nil)
}

func NewSuccessResponse(
	issueInstant time.Time,
	issuer string,
	inResponseTo string) *Response {
	status := Status{
		StatusCode: StatusCode{
			Value: crewjamsaml.StatusSuccess,
		},
	}
	response := newResponse(issueInstant, status, issuer)
	response.InResponseTo = inResponseTo
	return response
}
