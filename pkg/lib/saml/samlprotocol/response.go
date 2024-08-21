package samlprotocol

import (
	"time"

	"github.com/beevik/etree"
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
			Value: StatusRequester,
			StatusCode: &StatusCode{
				Value: StatusRequestDenied,
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
			Value: StatusRequester,
			StatusCode: &StatusCode{
				Value: StatusNoPassive,
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
			Value: StatusResponder,
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
			Value: StatusSuccess,
		},
	}
	response := newResponse(issueInstant, status, issuer)
	response.InResponseTo = inResponseTo
	return response
}
