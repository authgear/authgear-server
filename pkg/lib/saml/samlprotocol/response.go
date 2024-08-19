package samlprotocol

import (
	"time"

	"github.com/beevik/etree"
	crewjamsaml "github.com/crewjam/saml"
)

type Response struct {
	crewjamsaml.Response
}

func newResponse(issueInstant time.Time, status Status) *Response {
	return &Response{
		crewjamsaml.Response{
			ID:           GenerateResponseID(),
			IssueInstant: issueInstant,
			Version:      SAMLVersion2,
			Status:       status.status,
		},
	}
}

func NewRequestDeniedErrorResponse(
	issueInstant time.Time,
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
	return newResponse(issueInstant, Status{status: status})
}

type Status struct {
	status crewjamsaml.Status
}

func NewSuccessResponse(issueInstant time.Time) *Response {
	status := crewjamsaml.Status{
		StatusCode: crewjamsaml.StatusCode{
			Value: crewjamsaml.StatusSuccess,
		},
	}
	response := newResponse(issueInstant, Status{status: status})
	return response
}
