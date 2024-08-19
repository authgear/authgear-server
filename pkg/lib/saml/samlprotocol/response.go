package samlprotocol

import (
	"time"

	"github.com/beevik/etree"
	crewjamsaml "github.com/crewjam/saml"
)

type Response struct {
	crewjamsaml.Response
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

func NewInternalServerErrorResponse(issueInstant time.Time, issuer string) *Response {
	status := crewjamsaml.Status{
		StatusCode: crewjamsaml.StatusCode{
			Value: crewjamsaml.StatusResponder,
		},
		StatusMessage: &crewjamsaml.StatusMessage{
			Value: "unexpected error",
		},
	}
	return newResponse(issueInstant, Status{status: status}, issuer)
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
